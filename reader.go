// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/gonum/matrix/mat64"
)

// Read reads the data from the r NumPy data file io.Reader, into the
// provided pointed at value ptr.
// Read returns an error if the on-disk data type and the one provided
// don't match.
//
// If a *mat64.Dense matrix is passed to Read, the numpy-array data is loaded
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
		r.err = fmt.Errorf("npyio: invalid major version number (%d)", r.Header.Major)
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
		r.err = fmt.Errorf("npyio: invalid dictionary format")
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
		r.err = fmt.Errorf("npyio: invalid 'fortran_order' value (%v)", order)
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
// See npyio.Read() for documentation.
func (r *Reader) Read(ptr interface{}) error {
	rv := reflect.ValueOf(ptr)
	if !rv.IsValid() || rv.Kind() != reflect.Ptr {
		return errNotPtr
	}

	if rv.IsNil() {
		return errNilPtr
	}

	nelems := numElems(r.Header.Descr.Shape)
	dt := TypeFrom(r.Header.Descr.Type)
	if dt == nil {
		return fmt.Errorf("npyio: no reflect.Type for dtype=%v", r.Header.Descr.Type)
	}

	switch vptr := ptr.(type) {
	case *int, *uint, *[]int, *[]uint:
		return ErrInvalidType

	case *mat64.Dense:
		var data []float64
		err := r.Read(&data)
		if err != nil {
			return err
		}
		nrows, ncols, err := dimsFromShape(r.Header.Descr.Shape)
		if err != nil {
			return err
		}
		if r.Header.Descr.Fortran {
			*vptr = *mat64.NewDense(nrows, ncols, nil)
			i := 0
			for icol := 0; icol < ncols; icol++ {
				for irow := 0; irow < nrows; irow++ {
					vptr.Set(irow, icol, data[i])
					i++
				}
			}
		} else {
			*vptr = *mat64.NewDense(nrows, ncols, data)
		}
		return r.err

	case *bool:
		if dt != boolType {
			return ErrTypeMismatch
		}
		var buf [1]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		switch buf[0] {
		case 0:
			*vptr = false
		case 1:
			*vptr = true
		}
		return nil

	case *[]bool:
		if dt != boolType {
			return ErrTypeMismatch
		}
		*vptr = make([]bool, nelems)
		var buf [1]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			switch buf[0] {
			case 0:
				(*vptr)[i] = false
			case 1:
				(*vptr)[i] = true
			}
		}
		return nil

	case *int8:
		if dt != int8Type {
			return ErrTypeMismatch
		}
		var buf [1]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = int8(buf[0])
		return nil

	case *[]int8:
		if dt != int8Type {
			return ErrTypeMismatch
		}
		*vptr = make([]int8, nelems)
		var buf [1]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = int8(buf[0])
		}
		return nil

	case *int16:
		if dt != int16Type {
			return ErrTypeMismatch
		}
		var buf [2]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = int16(ble.Uint16(buf[:]))
		return nil

	case *[]int16:
		if dt != int16Type {
			return ErrTypeMismatch
		}
		*vptr = make([]int16, nelems)
		var buf [2]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = int16(ble.Uint16(buf[:]))
		}
		return nil

	case *int32:
		if dt != int32Type {
			return ErrTypeMismatch
		}
		var buf [4]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = int32(ble.Uint32(buf[:]))
		return nil

	case *[]int32:
		if dt != int32Type {
			return ErrTypeMismatch
		}
		*vptr = make([]int32, nelems)
		var buf [4]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = int32(ble.Uint32(buf[:]))
		}
		return nil

	case *int64:
		if dt != int64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = int64(ble.Uint64(buf[:]))
		return nil

	case *[]int64:
		if dt != int64Type {
			return ErrTypeMismatch
		}
		*vptr = make([]int64, nelems)
		var buf [8]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = int64(ble.Uint64(buf[:]))
		}
		return nil

	case *uint8:
		if dt != uint8Type {
			return ErrTypeMismatch
		}
		var buf [1]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = buf[0]
		return nil

	case *[]uint8:
		if dt != uint8Type {
			return ErrTypeMismatch
		}
		var buf [1]byte
		*vptr = make([]uint8, nelems)
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = buf[0]
		}
		return nil

	case *uint16:
		if dt != uint16Type {
			return ErrTypeMismatch
		}
		var buf [2]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = ble.Uint16(buf[:])
		return nil

	case *[]uint16:
		if dt != uint16Type {
			return ErrTypeMismatch
		}
		*vptr = make([]uint16, nelems)
		var buf [2]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = ble.Uint16(buf[:])
		}
		return nil

	case *uint32:
		if dt != uint32Type {
			return ErrTypeMismatch
		}
		var buf [4]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = ble.Uint32(buf[:])
		return nil

	case *[]uint32:
		if dt != uint32Type {
			return ErrTypeMismatch
		}
		*vptr = make([]uint32, nelems)
		var buf [4]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = ble.Uint32(buf[:])
		}
		return nil

	case *uint64:
		if dt != uint64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = ble.Uint64(buf[:])
		return nil

	case *[]uint64:
		if dt != uint64Type {
			return ErrTypeMismatch
		}
		*vptr = make([]uint64, nelems)
		var buf [8]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = ble.Uint64(buf[:])
		}
		return nil

	case *float32:
		if dt != float32Type {
			return ErrTypeMismatch
		}
		var buf [4]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = math.Float32frombits(ble.Uint32(buf[:]))
		return nil

	case *[]float32:
		if dt != float32Type {
			return ErrTypeMismatch
		}
		*vptr = make([]float32, nelems)
		var buf [4]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = math.Float32frombits(ble.Uint32(buf[:]))
		}
		return nil

	case *float64:
		if dt != float64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		*vptr = math.Float64frombits(ble.Uint64(buf[:]))
		return nil

	case *[]float64:
		if dt != float64Type {
			return ErrTypeMismatch
		}
		*vptr = make([]float64, nelems)
		var buf [8]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			(*vptr)[i] = math.Float64frombits(ble.Uint64(buf[:]))
		}
		return nil

	case *complex64:
		if dt != complex64Type {
			return ErrTypeMismatch
		}
		var buf [8]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		rcplx := math.Float32frombits(ble.Uint32(buf[0:4]))
		icplx := math.Float32frombits(ble.Uint32(buf[4:8]))
		*vptr = complex(rcplx, icplx)
		return nil

	case *[]complex64:
		if dt != complex64Type {
			return ErrTypeMismatch
		}
		*vptr = make([]complex64, nelems)
		var buf [8]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			rcplx := math.Float32frombits(ble.Uint32(buf[0:4]))
			icplx := math.Float32frombits(ble.Uint32(buf[4:8]))
			(*vptr)[i] = complex(rcplx, icplx)
		}
		return nil

	case *complex128:
		if dt != complex128Type {
			return ErrTypeMismatch
		}
		var buf [16]byte
		_, err := r.r.Read(buf[:])
		if err != nil {
			return err
		}
		rcplx := math.Float64frombits(ble.Uint64(buf[0:8]))
		icplx := math.Float64frombits(ble.Uint64(buf[8:16]))
		*vptr = complex(rcplx, icplx)
		return nil

	case *[]complex128:
		if dt != complex128Type {
			return ErrTypeMismatch
		}
		*vptr = make([]complex128, nelems)
		var buf [16]byte
		for i := 0; i < nelems; i++ {
			_, err := r.r.Read(buf[:])
			if err != nil {
				return err
			}
			rcplx := math.Float64frombits(ble.Uint64(buf[0:8]))
			icplx := math.Float64frombits(ble.Uint64(buf[8:16]))
			(*vptr)[i] = complex(rcplx, icplx)
		}
		return nil

	}

	rv = reflect.Indirect(rv)
	switch rv.Kind() {
	case reflect.Slice:
		rv.SetLen(0)
		elt := rv.Type().Elem()
		v := reflect.New(dt).Elem()
		slice := rv
		for i := 0; i < nelems; i++ {
			err := r.Read(v.Addr().Interface())
			if err != nil {
				return err
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
		v := reflect.New(dt).Elem()
		for i := 0; i < nelems; i++ {
			err := r.Read(v.Addr().Interface())
			if err != nil {
				return err
			}
			rv.Index(i).Set(v.Convert(elt))
		}
		return r.err

	case reflect.Bool:
		if !dt.ConvertibleTo(rv.Type()) {
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
		v := reflect.New(dt).Elem()
		if !dt.ConvertibleTo(rv.Type()) {
			return errNoConv
		}
		r.read(v.Addr().Interface())
		rv.Set(v.Convert(rv.Type()))
		return r.err

	case reflect.String, reflect.Map, reflect.Chan, reflect.Interface, reflect.Struct:
		return fmt.Errorf("npyio: type %v not supported", rv.Addr().Type())
	}

	panic("unreachable")
}

func dimsFromShape(shape []int) (int, int, error) {
	nrows := 0
	ncols := 0

	switch len(shape) {
	default:
		return -1, -1, fmt.Errorf("npyio: array shape not supported %v", shape)

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
	r.err = binary.Read(r.r, ble, v)
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
	switch dtype {
	case "b1", "<b1", "|b1", "bool":
		return reflect.TypeOf(false)
	case "u1", "<u1", "|u1", "uint8":
		return reflect.TypeOf(uint8(0))
	case "<u2", "uint16":
		return reflect.TypeOf(uint16(0))
	case "<u4", "uint32":
		return reflect.TypeOf(uint32(0))
	case "<u8", "uint64":
		return reflect.TypeOf(uint64(0))
	case "i1", "|i1", "<i1", "int8":
		return reflect.TypeOf(int8(0))
	case "<i2", "int16":
		return reflect.TypeOf(int16(0))
	case "<i4", "int32":
		return reflect.TypeOf(int32(0))
	case "<i8", "int64":
		return reflect.TypeOf(int64(0))
	case "<f4", "float32":
		return reflect.TypeOf(float32(0))
	case "<f8", "float64":
		return reflect.TypeOf(float64(0))
	case "<c8", "complex64":
		return reflect.TypeOf(complex64(0))
	case "<c16", "complex128":
		return reflect.TypeOf(complex128(0))
	}
	return nil
}
