// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"gonum.org/v1/gonum/mat"
)

var (
	rtDense = reflect.TypeOf((*mat.Dense)(nil)).Elem()
)

// Write writes 'val' into 'w' in the NumPy data format.
//
//  - if val is a scalar, it must be of a supported type (bools, (u)ints, floats and complexes)
//  - if val is a slice or array, it must be a slice/array of a supported type.
//    the shape (len,) will be written out.
//  - if val is a mat.Dense, the correct shape will be transmitted. (ie: (nrows, ncols))
//
// The data-array will always be written out in C-order (row-major).
func Write(w io.Writer, val interface{}) error {
	hdr := newHeader()
	rv := reflect.Indirect(reflect.ValueOf(val))
	dt, err := dtypeFrom(rv, rv.Type())
	if err != nil {
		return err
	}
	shape, err := shapeFrom(rv)
	if err != nil {
		return err
	}
	hdr.Descr.Type = dt
	hdr.Descr.Shape = shape

	rdt, err := newDtype(hdr.Descr.Type)
	if err != nil {
		return err
	}

	err = writeHeader(w, hdr, rdt)
	if err != nil {
		return err
	}

	return writeData(w, rv, rdt)
}

func writeHeader(w io.Writer, hdr Header, dt dType) error {
	err := binary.Write(w, dt.order, Magic[:])
	if err != nil {
		return err
	}
	err = binary.Write(w, dt.order, hdr.Major)
	if err != nil {
		return err
	}
	err = binary.Write(w, dt.order, hdr.Minor)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "{'descr': '%s', 'fortran_order': False, 'shape': %s, }",
		hdr.Descr.Type,
		shapeString(hdr.Descr.Shape),
	)
	var hdrSize int
	switch hdr.Major {
	case 1:
		hdrSize = 4 + len(Magic)
	case 2:
		hdrSize = 6 + len(Magic)
	default:
		return fmt.Errorf("npy: imvalid major version number (%d)", hdr.Major)
	}

	padding := (hdrSize + buf.Len() + 1) % 16
	_, err = buf.Write(bytes.Repeat([]byte{'\x20'}, padding))
	if err != nil {
		return err
	}
	_, err = buf.Write([]byte{'\n'})
	if err != nil {
		return err
	}

	buflen := int64(buf.Len())
	switch hdr.Major {
	case 1:
		err = binary.Write(w, dt.order, uint16(buflen))
	case 2:
		err = binary.Write(w, dt.order, uint32(buflen))
	default:
		return fmt.Errorf("npy: invalid major version number (%d)", hdr.Major)
	}

	if err != nil {
		return err
	}

	n, err := io.Copy(w, buf)
	if err != nil {
		return err
	}
	if n < buflen {
		return io.ErrShortWrite
	}

	return nil
}

func writeData(w io.Writer, rv reflect.Value, dt dType) error {
	rt := rv.Type()
	if rt == rtDense {
		m := rv.Interface().(mat.Dense)
		nrows, ncols := m.Dims()
		var buf [8]byte
		for i := 0; i < nrows; i++ {
			for j := 0; j < ncols; j++ {
				dt.order.PutUint64(buf[:], math.Float64bits(m.At(i, j)))
				_, err := w.Write(buf[:])
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	v := rv.Interface()
	switch v := v.(type) {
	case bool:
		switch v {
		case true:
			_, err := w.Write(trueUint8)
			return err
		case false:
			_, err := w.Write(falseUint8)
			return err
		}

	case []bool:
		for _, vv := range v {
			switch vv {
			case true:
				_, err := w.Write(trueUint8)
				if err != nil {
					return err
				}
			case false:
				_, err := w.Write(falseUint8)
				if err != nil {
					return err
				}
			}
		}
		return nil

	case uint, []uint, int, []int:
		return ErrInvalidType

	case uint8:
		buf := [1]byte{v}
		_, err := w.Write(buf[:])
		return err

	case []uint8:
		_, err := w.Write(v)
		return err

	case uint16:
		var buf [2]byte
		dt.order.PutUint16(buf[:], v)
		_, err := w.Write(buf[:])
		return err

	case []uint16:
		var buf [2]byte
		for _, vv := range v {
			dt.order.PutUint16(buf[:], vv)
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case uint32:
		var buf [4]byte
		dt.order.PutUint32(buf[:], v)
		_, err := w.Write(buf[:])
		return err

	case []uint32:
		var buf [4]byte
		for _, vv := range v {
			dt.order.PutUint32(buf[:], vv)
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case uint64:
		var buf [8]byte
		dt.order.PutUint64(buf[:], v)
		_, err := w.Write(buf[:])
		return err

	case []uint64:
		var buf [8]byte
		for _, vv := range v {
			dt.order.PutUint64(buf[:], vv)
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case int8:
		buf := [1]byte{byte(v)}
		_, err := w.Write(buf[:])
		return err

	case []int8:
		var buf [1]byte
		for _, vv := range v {
			buf[0] = uint8(vv)
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case int16:
		var buf [2]byte
		dt.order.PutUint16(buf[:], uint16(v))
		_, err := w.Write(buf[:])
		return err

	case []int16:
		var buf [2]byte
		for _, vv := range v {
			dt.order.PutUint16(buf[:], uint16(vv))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case int32:
		var buf [4]byte
		dt.order.PutUint32(buf[:], uint32(v))
		_, err := w.Write(buf[:])
		return err

	case []int32:
		var buf [4]byte
		for _, vv := range v {
			dt.order.PutUint32(buf[:], uint32(vv))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case int64:
		var buf [8]byte
		dt.order.PutUint64(buf[:], uint64(v))
		_, err := w.Write(buf[:])
		return err

	case []int64:
		var buf [8]byte
		for _, vv := range v {
			dt.order.PutUint64(buf[:], uint64(vv))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case float32:
		var buf [4]byte
		dt.order.PutUint32(buf[:], math.Float32bits(v))
		_, err := w.Write(buf[:])
		return err

	case []float32:
		var buf [4]byte
		for _, v := range v {
			dt.order.PutUint32(buf[:], math.Float32bits(v))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case float64:
		var buf [8]byte
		dt.order.PutUint64(buf[:], math.Float64bits(v))
		_, err := w.Write(buf[:])
		return err

	case []float64:
		var buf [8]byte
		for _, v := range v {
			dt.order.PutUint64(buf[:], math.Float64bits(v))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case complex64:
		var buf [8]byte
		dt.order.PutUint32(buf[0:4], math.Float32bits(real(v)))
		dt.order.PutUint32(buf[4:8], math.Float32bits(imag(v)))
		_, err := w.Write(buf[:])
		return err

	case []complex64:
		var buf [8]byte
		for _, v := range v {
			dt.order.PutUint32(buf[0:4], math.Float32bits(real(v)))
			dt.order.PutUint32(buf[4:8], math.Float32bits(imag(v)))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case complex128:
		var buf [16]byte
		dt.order.PutUint64(buf[0:8], math.Float64bits(real(v)))
		dt.order.PutUint64(buf[8:16], math.Float64bits(imag(v)))
		_, err := w.Write(buf[:])
		return err

	case []complex128:
		var buf [16]byte
		for _, v := range v {
			dt.order.PutUint64(buf[0:8], math.Float64bits(real(v)))
			dt.order.PutUint64(buf[8:16], math.Float64bits(imag(v)))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case string:
		o := []byte(v)
		o = append(o, 0)
		_, err := w.Write(o)
		if err != nil {
			return err
		}
		return nil

	case []string:
		n := dt.size
		switch {
		case dt.utf:
			for _, str := range v {
				o := make([]byte, n*utf8.UTFMax)
				i := 0
				for _, v := range str {
					dt.order.PutUint32(o[i:i+utf8.UTFMax], uint32(v))
					i += utf8.UTFMax
				}
				_, err := w.Write(o)
				if err != nil {
					return err
				}
			}

		case !dt.utf:
			o := make([]byte, len(v)*n)
			for i, v := range v {
				copy(o[i:i+n], []byte(v))
			}
			_, err := w.Write(o)
			if err != nil {
				return err
			}
		}
		return nil
	}

	switch rt.Kind() {
	case reflect.Array:
		switch rt.Elem().Kind() {
		case reflect.Bool, reflect.Int, reflect.Uint:
			n := rv.Len()
			for i := 0; i < n; i++ {
				elem := rv.Index(i)
				err := writeData(w, elem, dt)
				if err != nil {
					return err
				}
			}
			return nil
		default:
			return binary.Write(w, dt.order, v)
		}

	case reflect.Interface, reflect.Chan, reflect.Map, reflect.Struct:
		return fmt.Errorf("npy: type %v not supported", rt)
	}

	return binary.Write(w, dt.order, v)
}

func dtypeFrom(rv reflect.Value, rt reflect.Type) (string, error) {
	if rt == rtDense {
		return "<f8", nil
	}

	switch rt.Kind() {
	case reflect.Bool:
		return "|b1", nil
	case reflect.Uint8:
		return "|u1", nil
	case reflect.Uint16:
		return "<u2", nil
	case reflect.Uint32:
		return "<u4", nil
	case reflect.Uint, reflect.Uint64:
		return "<u8", nil
	case reflect.Int8:
		return "|i1", nil
	case reflect.Int16:
		return "<i2", nil
	case reflect.Int32:
		return "<i4", nil
	case reflect.Int, reflect.Int64:
		return "<i8", nil
	case reflect.Float32:
		return "<f4", nil
	case reflect.Float64:
		return "<f8", nil
	case reflect.Complex64:
		return "<c8", nil
	case reflect.Complex128:
		return "<c16", nil

	case reflect.Array:
		et := rt.Elem()
		switch et.Kind() {
		default:
			return dtypeFrom(reflect.Value{}, et)
		case reflect.String:
			slice := rv.Slice(0, rt.Len()).Interface().([]string)
			n := 0
			for _, str := range slice {
				if len(str) > n {
					n = len(str)
				}
			}
			return fmt.Sprintf("<U%d", n), nil
		}

	case reflect.Slice:
		rt = rt.Elem()
		switch rt.Kind() {
		default:
			return dtypeFrom(reflect.Value{}, rt)
		case reflect.String:
			slice := rv.Interface().([]string)
			n := 0
			for _, str := range slice {
				if len(str) > n {
					n = len(str)
				}
			}
			return fmt.Sprintf("<U%d", n), nil
		}

	case reflect.String:
		return fmt.Sprintf("<U%d", len(rv.Interface().(string))), nil

	case reflect.Map, reflect.Chan, reflect.Interface, reflect.Struct:
		return "", fmt.Errorf("npy: type %v not supported", rt)
	}

	return "", fmt.Errorf("npy: type %v not supported", rt)
}

func shapeFrom(rv reflect.Value) ([]int, error) {
	if m, ok := rv.Interface().(mat.Dense); ok {
		nrows, ncols := m.Dims()
		return []int{nrows, ncols}, nil
	}

	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		eshape, err := shapeFrom(rv.Index(0))
		if err != nil {
			return nil, err
		}
		return append([]int{rv.Len()}, eshape...), nil

	case reflect.String:
		return nil, nil

	case reflect.Map, reflect.Chan, reflect.Interface, reflect.Struct:
		return nil, fmt.Errorf("npy: type %v not supported", rt)
	}

	// scalar.
	return nil, nil
}

func shapeString(shape []int) string {
	switch len(shape) {
	case 0:
		return "()"
	case 1:
		return fmt.Sprintf("(%d,)", shape[0])
	default:
		var str []string
		for _, v := range shape {
			str = append(str, strconv.Itoa(v))
		}
		return fmt.Sprintf("(%s)", strings.Join(str, ", "))
	}

}
