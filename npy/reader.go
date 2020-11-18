// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gonum.org/v1/gonum/mat"
)

// Read reads the data from the r NumPy data file io.Reader, into the
// provided pointed at value ptr.
// Read returns an error if the on-disk data type and the one provided
// don't match.
//
// If a *mat.Dense matrix is passed to Read, the numpy-array data is loaded
// into the Dense matrix, honouring Fortran/C-order and dimensions/shape
// parameters.
//
// Only numpy-arrays with up to 2 dimensions are supported.
// Only numpy-arrays with elements convertible to float64 are supported.
func Read(r io.Reader, ptr interface{}) error {
	rr, err := NewReader(r)
	if err != nil {
		return err
	}

	return rr.Read(ptr)
}

// Reader reads data from a NumPy data file.
type Reader struct {
	r   io.Reader
	err error // last error

	Header Header
	order  binary.ByteOrder
}

// NewReader creates a new NumPy data file format reader.
func NewReader(r io.Reader) (*Reader, error) {
	rr := &Reader{r: r}
	rr.readHeader()
	if rr.err != nil {
		return nil, rr.err
	}
	return rr, rr.err
}

func (r *Reader) readHeader() {
	if r.err != nil {
		return
	}
	r.order = binary.LittleEndian
	var magic [6]byte
	r.read(&magic)
	if r.err != nil {
		return
	}
	if magic != Magic {
		r.err = ErrInvalidNumPyFormat
		return
	}

	var hdrLen int

	r.read(&r.Header.Major)
	r.read(&r.Header.Minor)
	switch r.Header.Major {
	case 1:
		var v uint16
		r.read(&v)
		hdrLen = int(v)
	case 2:
		var v uint32
		r.read(&v)
		hdrLen = int(v)
	default:
		r.err = fmt.Errorf("npy: invalid major version number (%d)", r.Header.Major)
	}

	if r.err != nil {
		return
	}

	hdr := make([]byte, hdrLen)
	r.read(&hdr)
	idx := bytes.LastIndexByte(hdr, '\n')
	hdr = hdr[:idx]
	r.readDescr(hdr)
}

func (r *Reader) readDescr(buf []byte) {
	if r.err != nil {
		return
	}

	var (
		descrKey = []byte("'descr': ")
		orderKey = []byte("'fortran_order': ")
		shapeKey = []byte("'shape': ")
		trailer  = []byte(", ")
	)

	begDescr := bytes.Index(buf, descrKey)
	begOrder := bytes.Index(buf, orderKey)
	begShape := bytes.Index(buf, shapeKey)
	endDescr := bytes.Index(buf, []byte("}"))
	if begDescr < 0 || begOrder < 0 || begShape < 0 {
		r.err = fmt.Errorf("npy: invalid dictionary format")
		return
	}

	descr := string(buf[begDescr+len(descrKey)+1 : begOrder-len(trailer)-1])
	order := string(buf[begOrder+len(orderKey) : begShape-len(trailer)])
	shape := buf[begShape+len(shapeKey) : endDescr-len(trailer)]

	r.Header.Descr.Type = descr // FIXME(sbinet): better handling
	switch order {
	case "False":
		r.Header.Descr.Fortran = false
	case "True":
		r.Header.Descr.Fortran = true
	default:
		r.err = fmt.Errorf("npy: invalid 'fortran_order' value (%v)", order)
		return
	}

	if string(shape) == "()" {
		r.Header.Descr.Shape = nil
		return
	}

	shape = shape[1 : len(shape)-1]
	toks := strings.Split(string(shape), ",")
	for _, tok := range toks {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		i, err := strconv.Atoi(tok)
		if err != nil {
			r.err = err
			return
		}
		r.Header.Descr.Shape = append(r.Header.Descr.Shape, int(i))
	}

}

// Read reads the numpy-array data from the underlying NumPy file.
// Read returns an error if the on-disk data type and the provided one
// don't match.
//
// See npy.Read() for documentation.
func (r *Reader) Read(ptr interface{}) error {
	if r.err != nil {
		return r.err
	}

	rv := reflect.ValueOf(ptr)
	if !rv.IsValid() || rv.Kind() != reflect.Ptr {
		return errNotPtr
	}

	if rv.IsNil() {
		return errNilPtr
	}

	nelems := numElems(r.Header.Descr.Shape)
	dt, err := newDtype(r.Header.Descr.Type)
	if err != nil {
		return err
	}
	r.order = dt.order

	switch vptr := ptr.(type) {
	case *int, *uint, *[]int, *[]uint:
		return ErrInvalidType

	case *mat.Dense:
		var data []float64
		err := r.Read(&data)
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		nrows, ncols, err := dimsFromShape(r.Header.Descr.Shape)
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		if r.Header.Descr.Fortran {
			*vptr = *mat.NewDense(nrows, ncols, nil)
			i := 0
			for icol := 0; icol < ncols; icol++ {
				for irow := 0; irow < nrows; irow++ {
					vptr.Set(irow, icol, data[i])
					i++
				}
			}
		} else {
			*vptr = *mat.NewDense(nrows, ncols, data)
		}
		return r.err

	case *bool:
		if dt.rt != boolType {
			return ErrTypeMismatch
		}
		var buf [1]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		switch buf[0] {
		case 0:
			*vptr = false
		case 1:
			*vptr = true
		}
		return r.err

	case *[]bool:
		if dt.rt != boolType {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]bool, n)
		}
		var buf [1]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			switch buf[0] {
			case 0:
				(*vptr)[i] = false
			case 1:
				(*vptr)[i] = true
			}
		}
		return r.err

	case *int8:
		if dt.rt != int8Type {
			return ErrTypeMismatch
		}
		var buf [1]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = int8(buf[0])
		return r.err

	case *[]int8:
		if dt.rt != int8Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]int8, n)
		}
		var buf [1]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = int8(buf[0])
		}
		return r.err

	case *int16:
		if dt.rt != int16Type {
			return ErrTypeMismatch
		}
		var buf [2]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = int16(dt.order.Uint16(buf[:]))
		return r.err

	case *[]int16:
		if dt.rt != int16Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]int16, n)
		}
		var buf [2]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = int16(dt.order.Uint16(buf[:]))
		}
		return r.err

	case *int32:
		if dt.rt != int32Type {
			return ErrTypeMismatch
		}
		var buf [4]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = int32(dt.order.Uint32(buf[:]))
		return r.err

	case *[]int32:
		if dt.rt != int32Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]int32, n)
		}
		var buf [4]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = int32(dt.order.Uint32(buf[:]))
		}
		return r.err

	case *int64:
		if dt.rt != int64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = int64(dt.order.Uint64(buf[:]))
		return r.err

	case *[]int64:
		if dt.rt != int64Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]int64, n)
		}
		var buf [8]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = int64(dt.order.Uint64(buf[:]))
		}
		return r.err

	case *uint8:
		if dt.rt != uint8Type {
			return ErrTypeMismatch
		}
		var buf [1]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = buf[0]
		return r.err

	case *[]uint8:
		if dt.rt != uint8Type {
			return ErrTypeMismatch
		}
		var buf [1]byte
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]uint8, n)
		}
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = buf[0]
		}
		return r.err

	case *uint16:
		if dt.rt != uint16Type {
			return ErrTypeMismatch
		}
		var buf [2]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = dt.order.Uint16(buf[:])
		return r.err

	case *[]uint16:
		if dt.rt != uint16Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]uint16, n)
		}
		var buf [2]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = dt.order.Uint16(buf[:])
		}
		return r.err

	case *uint32:
		if dt.rt != uint32Type {
			return ErrTypeMismatch
		}
		var buf [4]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = dt.order.Uint32(buf[:])
		return r.err

	case *[]uint32:
		if dt.rt != uint32Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]uint32, n)
		}
		var buf [4]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = dt.order.Uint32(buf[:])
		}
		return r.err

	case *uint64:
		if dt.rt != uint64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = dt.order.Uint64(buf[:])
		return r.err

	case *[]uint64:
		if dt.rt != uint64Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]uint64, n)
		}
		var buf [8]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = dt.order.Uint64(buf[:])
		}
		return r.err

	case *float32:
		if dt.rt != float32Type {
			return ErrTypeMismatch
		}
		var buf [4]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = math.Float32frombits(dt.order.Uint32(buf[:]))
		return r.err

	case *[]float32:
		if dt.rt != float32Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]float32, n)
		}
		var buf [4]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = math.Float32frombits(dt.order.Uint32(buf[:]))
		}
		return r.err

	case *float64:
		if dt.rt != float64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		*vptr = math.Float64frombits(dt.order.Uint64(buf[:]))
		return r.err

	case *[]float64:
		if dt.rt != float64Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]float64, n)
		}
		var buf [8]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			(*vptr)[i] = math.Float64frombits(dt.order.Uint64(buf[:]))
		}
		return r.err

	case *complex64:
		if dt.rt != complex64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		rcplx := math.Float32frombits(dt.order.Uint32(buf[0:4]))
		icplx := math.Float32frombits(dt.order.Uint32(buf[4:8]))
		*vptr = complex(rcplx, icplx)
		return r.err

	case *[]complex64:
		if dt.rt != complex64Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]complex64, n)
		}
		var buf [8]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			rcplx := math.Float32frombits(dt.order.Uint32(buf[0:4]))
			icplx := math.Float32frombits(dt.order.Uint32(buf[4:8]))
			(*vptr)[i] = complex(rcplx, icplx)
		}
		return r.err

	case *complex128:
		if dt.rt != complex128Type {
			return ErrTypeMismatch
		}
		var buf [16]byte
		_, err := r.r.Read(buf[:])
		if err != nil && err != io.EOF {
			r.err = err
			return r.err
		}
		rcplx := math.Float64frombits(dt.order.Uint64(buf[0:8]))
		icplx := math.Float64frombits(dt.order.Uint64(buf[8:16]))
		*vptr = complex(rcplx, icplx)
		return r.err

	case *[]complex128:
		if dt.rt != complex128Type {
			return ErrTypeMismatch
		}
		n := min(len(*vptr), nelems)
		if n == 0 {
			n = nelems
			*vptr = make([]complex128, n)
		}
		var buf [16]byte
		for i := 0; i < n; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			rcplx := math.Float64frombits(dt.order.Uint64(buf[0:8]))
			icplx := math.Float64frombits(dt.order.Uint64(buf[8:16]))
			(*vptr)[i] = complex(rcplx, icplx)
		}
		return r.err

	case *string:
		if dt.rt != stringType {
			return ErrTypeMismatch
		}

		switch {
		case dt.utf:
			raw, err := ioutil.ReadAll(io.LimitReader(r.r, utf8.UTFMax*int64(dt.size)))
			if err != nil {
				r.err = err
				return r.err
			}
			var str string
			for len(raw) > 0 {
				r, size := utf8.DecodeRune(raw)
				str += string(r)
				raw = raw[size:]
			}
			*vptr = str
			return r.err

		case !dt.utf:
			buf, err := ioutil.ReadAll(io.LimitReader(r.r, int64(dt.size)))
			if err != nil {
				r.err = err
				return r.err
			}
			n := bytes.Index(buf, []byte{0})
			if n > 0 {
				buf = buf[:n]
			}
			*vptr = string(buf)
			return r.err
		}
	}

	rv = reflect.Indirect(rv)
	switch rv.Kind() {
	case reflect.Slice:
		rv.SetLen(0)
		elt := rv.Type().Elem()
		v := reflect.New(dt.rt).Elem()
		slice := rv
		for i := 0; i < nelems; i++ {
			err := r.Read(v.Addr().Interface())
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			slice = reflect.Append(slice, v.Convert(elt))
		}
		rv.Set(slice)
		return r.err

	case reflect.Array:
		if nelems > rv.Type().Len() {
			return errDims
		}

		elt := rv.Type().Elem()
		v := reflect.New(dt.rt).Elem()
		for i := 0; i < nelems; i++ {
			err := r.Read(v.Addr().Interface())
			if err != nil && err != io.EOF {
				r.err = err
				return r.err
			}
			rv.Index(i).Set(v.Convert(elt))
		}
		return r.err

	case reflect.Bool:
		if !dt.rt.ConvertibleTo(rv.Type()) {
			return errNoConv
		}
		var v uint8
		r.read(&v)
		rv.SetBool(v == 1)
		return r.err

	case reflect.Int, reflect.Uint,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		v := reflect.New(dt.rt).Elem()
		if !dt.rt.ConvertibleTo(rv.Type()) {
			return errNoConv
		}
		r.read(v.Addr().Interface())
		rv.Set(v.Convert(rv.Type()))
		return r.err

	case reflect.String, reflect.Map, reflect.Chan, reflect.Interface, reflect.Struct:
		return fmt.Errorf("npy: type %v not supported", rv.Addr().Type())
	}

	panic("unreachable")
}

func dimsFromShape(shape []int) (int, int, error) {
	nrows := 0
	ncols := 0

	switch len(shape) {
	default:
		return -1, -1, fmt.Errorf("npy: array shape not supported %v", shape)

	case 0:
		nrows = 1
		ncols = 1

	case 1:
		nrows = shape[0]
		ncols = 1

	case 2:
		nrows = shape[0]
		ncols = shape[1]
	}

	return nrows, ncols, nil
}

func (r *Reader) read(v interface{}) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, r.order, v)
}

func numElems(shape []int) int {
	n := 1
	for _, v := range shape {
		n *= v
	}
	return n
}

// TypeFrom returns the reflect.Type corresponding to the numpy-dtype string, if any.
func TypeFrom(dtype string) reflect.Type {
	dt, err := newDtype(dtype)
	if err != nil {
		return nil
	}
	return dt.rt
}

var (
	reStrPre  = regexp.MustCompile(`^[|]*?(\d.*)[Sa]$`)
	reStrPost = regexp.MustCompile(`^[|]*?[Sa](\d.*)$`)
	reUniPre  = regexp.MustCompile(`^[<|>]*?(\d.*)U$`)
	reUniPost = regexp.MustCompile(`^[<|>]*?U(\d.*)$`)
)

func stringLen(dtype string) (int, error) {
	if m := reStrPre.FindStringSubmatch(dtype); m != nil {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
	if m := reStrPost.FindStringSubmatch(dtype); m != nil {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
	if m := reUniPre.FindStringSubmatch(dtype); m != nil {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
	if m := reUniPost.FindStringSubmatch(dtype); m != nil {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
	return 0, fmt.Errorf("npy: %q is not a string-like dtype", dtype)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
