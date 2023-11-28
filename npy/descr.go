// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	py "github.com/nlpodyssey/gopickle/types"
	"github.com/sbinet/npyio/npy/float16"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode/utf32"
)

// ArrayDescr describes a numpy data type.
type ArrayDescr struct {
	kind  byte
	order binary.ByteOrder
	flags int // flags describing data type
	esize int // element size in bytes
	align int // alignment needed for this type

	subarr *subarrayDescr // non-nil if this type is an array (C-continguous) of some other type.
	names  []string       // fields' names (if any)
	fields structFields   // fields (if any)
	meta   map[string]any
}

func newDescrFrom(v any, flags int) (*ArrayDescr, error) {
	switch v := v.(type) {
	case nil:
		return &ArrayDescr{kind: 'f', order: binary.LittleEndian, esize: 8, align: 8, flags: flags}, nil
	case *ArrayDescr:
		return v, nil
	case string:
		return newDescrFromStr(v, flags)
	default:
		return nil, fmt.Errorf("invalid type %T for dtype ctor", v)
	}
}

func newDescrFromStr(typ string, flags int) (*ArrayDescr, error) {
	dt := &ArrayDescr{order: nil, esize: -1, align: -1, flags: flags}

	if len(typ) == 0 {
		return nil, fmt.Errorf("data type %q not understood", typ)
	}

	descr := typ
	switch {
	case strings.HasPrefix(typ, "<"):
		descr = descr[1:]
		dt.order = binary.LittleEndian
	case strings.HasPrefix(typ, "="):
		descr = descr[1:]
		dt.order = nativeEndian
	case strings.HasPrefix(typ, "|"):
		descr = descr[1:]
		dt.order = nil
	case strings.HasPrefix(typ, ">"):
		descr = descr[1:]
		dt.order = binary.BigEndian
	}

	if len(descr) == 0 {
		return nil, fmt.Errorf("data type %v not understood", typ)
	}

	if isDatetimeStr(descr) {
		// FIXME(sbinet)
		return nil, fmt.Errorf("datetime string not implemented")
	}

	err := dt.init(descr)
	if err != nil {
		return nil, err
	}

	return dt, nil
}

func isDatetimeStr(typ string) bool {
	if len(typ) < 2 {
		return false
	}
	switch {
	case strings.HasPrefix(typ, "M8"), strings.HasPrefix(typ, "datetime64"):
		return true
	case strings.HasPrefix(typ, "m8"), strings.HasPrefix(typ, "timedelta64"):
		return true
	}
	return false
}

func (dt *ArrayDescr) init(descr string) error {
	switch len(descr) {
	case 0:
		return fmt.Errorf("invalid typecode %q", descr)

	case 1: // a typecode like "d", "f", ...
		dt.kind = descr[0]

	default:
		dt.kind = descr[0]
		v, err := strconv.ParseUint(descr[1:], 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse typecode %q: %w", descr, err)
		}
		dt.esize = int(v)
		switch dt.kind {
		case 'b', 'i', 'u', 'f', 'm', 'M':
			dt.align = dt.esize
		case 'c':
			dt.align = dt.esize / 2
		case 'O':
			// dt.esize = 8
			dt.align = 8
		case 'S':
			dt.align = 1
		case 'V':
			dt.align = 1
		}
	}

	return nil
}

type structFields map[string]structField
type structField struct {
	dtype  ArrayDescr
	offset uint32
}

type subarrayDescr struct {
	dtype ArrayDescr
	shape []int
}

var (
	_ py.Callable        = (*ArrayDescr)(nil)
	_ py.PyStateSettable = (*ArrayDescr)(nil)
)

func (*ArrayDescr) Call(args ...any) (any, error) {
	switch sz := len(args); {
	case sz < 1, sz > 3:
		return nil, fmt.Errorf("invalid tuple length (got=%d)", sz)
	}

	descr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid descr type %T", args[0])
	}

	const flags = 0
	return newDescrFromStr(descr, flags)
}

func (dt *ArrayDescr) PySetState(arg any) error {
	tuple, ok := arg.(*py.Tuple)
	if !ok {
		return fmt.Errorf("invalid argument type %T", arg)
	}

	if tuple == nil {
		// FIXME(sbinet): is a nil tuple a valid value ?
		return fmt.Errorf("invalid __setstate__ argument %+v", tuple)
	}

	var (
		vers   int = 4
		order  string
		subarr py.Tuple
		names  py.Tuple
		fields py.Dict
		meta   py.Dict
		esize  = -1
		align  = -1
		flags  = 0
	)

	switch tuple.Len() {
	case 9:
		err := parseTuple(tuple, &vers, &order, &subarr, &names, &fields, &esize, &align, &flags, &meta)
		if err != nil {
			return fmt.Errorf("could not parse tuple: %w", err)
		}
	case 8:
		err := parseTuple(tuple, &vers, &order, &subarr, &names, &fields, &esize, &align, &flags)
		if err != nil {
			return fmt.Errorf("could not parse tuple: %w", err)
		}
	case 7:
		err := parseTuple(tuple, &vers, &order, &subarr, &names, &fields, &esize, &align)
		if err != nil {
			return fmt.Errorf("could not parse tuple: %w", err)
		}
	case 6:
		err := parseTuple(tuple, &vers, &order, &subarr, &fields, &esize, &align)
		if err != nil {
			return fmt.Errorf("could not parse tuple: %w", err)
		}
	case 5:
		vers = 0
		err := parseTuple(tuple, &order, &subarr, &fields, &esize, &align)
		if err != nil {
			return fmt.Errorf("could not parse tuple: %w", err)
		}
	default:
		switch {
		case tuple.Len() > 5:
			v, ok := tuple.Get(0).(int)
			if !ok {
				return fmt.Errorf("invalid __setstate__ arg[0]: got=%T, want=int", tuple.Get(0))
			}
			vers = v
		default:
			vers = -1
		}
	}

	if vers < 0 || vers > 4 {
		return fmt.Errorf("invalid version=%d for numpy.dtype pickle", vers)
	}

	if vers == 0 || vers == 1 {
		return fmt.Errorf("unhandled version=%d for numpy.dtype pickle", vers)
	}

	switch order {
	case "<":
		dt.order = binary.LittleEndian
	case ">":
		dt.order = binary.BigEndian
	case "=":
		dt.order = nativeEndian
	case "|":
		dt.order = nil
	}

	if subarr.Len() > 0 {
		var (
			subdt ArrayDescr
			tuple py.Tuple
			shape []int
		)
		err := parseTuple(&subarr, &subdt, &tuple)
		if err != nil {
			return fmt.Errorf("could not parse subarray tuple: %w", err)
		}
		for i := range tuple {
			v, ok := tuple[i].(int)
			if !ok {
				return fmt.Errorf("could not parse subarray shape[%d]: type=%T", i, tuple[i])
			}
			shape = append(shape, v)
		}
		dt.subarr = &subarrayDescr{
			dtype: subdt,
			shape: shape,
		}
	}

	if names.Len() > 0 {
		for _, v := range names {
			name, ok := v.(string)
			if !ok {
				return fmt.Errorf("invalid field name type %T", v)
			}
			dt.names = append(dt.names, name)
		}
	}

	if fields.Len() > 0 {
		dt.fields = make(structFields, fields.Len())
		for i := 0; i < fields.Len(); i++ {
			v, ok := fields.Get(dt.names[i])
			if !ok {
				return fmt.Errorf("invalid field offset name %q", dt.names[i])
			}
			tup, ok := v.(*py.Tuple)
			if !ok {
				return fmt.Errorf("invalid field offset type %T", v)
			}
			if got, want := tup.Len(), 2; got != want {
				return fmt.Errorf("invalid field offset tuple length (got=%d, want=%d)", got, want)
			}
			fdt, ok := tup.Get(0).(*ArrayDescr)
			if !ok {
				return fmt.Errorf("invalid field offset dtype")
			}
			offset, ok := tup.Get(1).(int)
			if !ok {
				return fmt.Errorf("invalid field offset")
			}
			dt.fields[dt.names[i]] = structField{*fdt, uint32(offset)}
		}
	}

	if esize >= 0 {
		dt.esize = esize
	}
	if align >= 0 {
		dt.align = align
	}
	if flags >= 0 {
		dt.flags = flags
	}

	if meta.Len() > 0 {
		return fmt.Errorf("dtype with metadata not handled (yet?)")
	}

	return nil
}

func (dt ArrayDescr) unmarshal(raw []byte, shape []int) (any, error) {
	// FIXME(sbinet): handle ndims
	// FIXME(sbinet): handle sub-arrays ?
	// FIXME(sbinet): handle strides

	if dt.subarr != nil {
		return nil, fmt.Errorf("sub-arrays not handled")
	}

	switch dt.kind {
	case 'b':
		data := make([]bool, len(raw))
		for i, v := range raw {
			if v == 0 {
				continue
			}
			data[i] = true
		}
		return data, nil

	case 'i':
		switch dt.esize {
		case 1:
			data := make([]int8, len(raw))
			for i, v := range raw {
				data[i] = int8(v)
			}
			return data, nil

		case 2:
			data := make([]int16, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, int16(dt.order.Uint16(raw[i:])))
			}
			return data, nil

		case 4:
			data := make([]int32, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, int32(dt.order.Uint32(raw[i:])))
			}
			return data, nil

		case 8:
			data := make([]int64, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, int64(dt.order.Uint64(raw[i:])))
			}
			return data, nil

		default:
			return nil, fmt.Errorf("unhandled esize=%d for kind=%q", dt.esize, dt.kind)
		}

	case 'u':
		switch dt.esize {
		case 1:
			data := make([]uint8, len(raw))
			copy(data, raw)
			return data, nil

		case 2:
			data := make([]uint16, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, dt.order.Uint16(raw[i:]))
			}
			return data, nil

		case 4:
			data := make([]uint32, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, dt.order.Uint32(raw[i:]))
			}
			return data, nil

		case 8:
			data := make([]uint64, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, dt.order.Uint64(raw[i:]))
			}
			return data, nil

		default:
			return nil, fmt.Errorf("unhandled esize=%d for kind=%q", dt.esize, dt.kind)
		}

	case 'f':
		switch dt.esize {
		case 2:
			data := make([]float16.Num, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, float16.Float16Frombits(dt.order.Uint16(raw[i:])))
			}
			return data, nil

		case 4:
			data := make([]float32, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, math.Float32frombits(dt.order.Uint32(raw[i:])))
			}
			return data, nil

		case 8:
			data := make([]float64, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, math.Float64frombits(dt.order.Uint64(raw[i:])))
			}
			return data, nil

		default:
			return nil, fmt.Errorf("unhandled esize=%d for kind=%q", dt.esize, dt.kind)
		}

	case 'c':
		switch dt.esize {
		case 8:
			data := make([]complex64, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, complex(
					math.Float32frombits(dt.order.Uint32(raw[i+0:])),
					math.Float32frombits(dt.order.Uint32(raw[i+4:])),
				))
			}
			return data, nil

		case 16:
			data := make([]complex128, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, complex(
					math.Float64frombits(dt.order.Uint64(raw[i+0:])),
					math.Float64frombits(dt.order.Uint64(raw[i+8:])),
				))
			}
			return data, nil

		default:
			return nil, fmt.Errorf("unhandled esize=%d for kind=%q", dt.esize, dt.kind)
		}

	case 'S':
		switch len(shape) {
		case 0:
			return string(raw), nil // FIXME(sbinet): use subset ? (shape/dims/...)
		default:
			data := make([]string, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				data = append(data, string(raw[i:i+dt.esize])) // FIXME(sbinet): no-alloc ?
			}
			return data, nil
		}

	case 'U':
		order := utf32.BigEndian
		if dt.order == binary.LittleEndian {
			order = utf32.LittleEndian
		}
		dec := utf32.UTF32(order, utf32.IgnoreBOM).NewDecoder()
		switch len(shape) {
		case 0:
			data, err := decodeUTF(dec, raw)
			if err != nil {
				return nil, fmt.Errorf("could not decode utf array: %w", err)
			}
			return data, nil

		default:
			data := make([]string, 0, len(raw)/dt.esize)
			for i := 0; i < len(raw); i += dt.esize {
				v, err := decodeUTF(dec, raw[i:i+dt.esize])
				if err != nil {
					return nil, fmt.Errorf("could not decode utf array element %d: %w", i/dt.esize, err)
				}
				data = append(data, v)
			}
			return data, nil
		}

	case 'O':
		pkl := newUnpickler(bytes.NewReader(raw))
		data, err := pkl.Load()
		if err != nil {
			return nil, fmt.Errorf("could not unpickle data: %w", err)
		}
		return data, nil

	default:
		return nil, fmt.Errorf("unknown dtype [%c%d]", dt.kind, dt.esize)
	}
}

func (dt ArrayDescr) itemsize() int {
	if dt.esize < 0 {
		panic(fmt.Errorf("unknown dtype [%c%d]", dt.kind, dt.esize))
	}
	return dt.esize
}

func (dt ArrayDescr) String() string {
	o := new(strings.Builder)
	fmt.Fprintf(o,
		"ArrayDescr{kind: '%s', order: '%s', flags: %d, esize: %d, align: %d, subarr: %v, names: %v, fields: %v, meta: %v}",
		string(dt.kind),
		orderToString(dt.order),
		dt.flags,
		dt.esize,
		dt.align,
		dt.subarr,
		dt.names,
		dt.fields,
		dt.meta,
	)
	return o.String()
}

func (sfs structFields) String() string {
	o := new(strings.Builder)
	fmt.Fprintf(o, "{")
	keys := make([]string, 0, len(sfs))
	for k := range sfs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i > 0 {
			fmt.Fprintf(o, ", ")
		}
		v := sfs[k]
		fmt.Fprintf(o, "%q: %v", k, v)
	}
	fmt.Fprintf(o, "}")
	return o.String()
}

func (sf structField) String() string {
	o := new(strings.Builder)
	fmt.Fprintf(o, "field{dtype: %v, offset: %d}", sf.dtype, sf.offset)
	return o.String()
}

func (sub subarrayDescr) String() string {
	o := new(strings.Builder)
	fmt.Fprintf(o, "subarr{dtype: %v, shape: %v}", sub.dtype, sub.shape)
	return o.String()
}

func decodeUTF(dec *encoding.Decoder, raw []byte) (string, error) {
	// FIXME(sbinet): use subset ? (shape/dims/...)
	vs := make([]byte, 0, utf8.RuneCount(raw))
	raw, err := dec.Bytes(raw)
	if err != nil {
		return "", err
	}
	i := 0
loop:
	for {
		r, sz := utf8.DecodeRune(raw[i:])
		switch r {
		case utf8.RuneError:
			if sz == 0 {
				break loop
			}
			return string(vs), fmt.Errorf("invalid rune")
		default:
			vs = utf8.AppendRune(vs, r)
			i += sz
		}
	}
	return string(vs), nil
}
