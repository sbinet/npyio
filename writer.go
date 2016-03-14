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

var (
	rtDense = reflect.TypeOf((*mat64.Dense)(nil)).Elem()
)

// Write writes 'val' into 'w' in the NumPy data format.
//
//  - if val is a scalar, it must be of a supported type (bools, (u)ints, floats and complexes)
//  - if val is a slice or array, it must be a slice/array of a supported type.
//    the shape (len,) will be written out.
//  - if val is a mat64.Dense, the correct shape will be transmitted. (ie: (nrows, ncols))
//
// The data-array will always be written out in C-order (row-major).
func Write(w io.Writer, val interface{}) error {
	hdr := newHeader()
	rv := reflect.Indirect(reflect.ValueOf(val))
	dt, err := dtypeFrom(rv.Type())
	if err != nil {
		return err
	}
	shape, err := shapeFrom(rv)
	if err != nil {
		return err
	}
	hdr.Descr.Type = dt
	hdr.Descr.Shape = shape

	err = writeHeader(w, hdr)
	if err != nil {
		return err
	}

	return writeData(w, rv)
}

func writeHeader(w io.Writer, hdr Header) error {
	err := binary.Write(w, ble, Magic[:])
	if err != nil {
		return err
	}
	err = binary.Write(w, ble, hdr.Major)
	if err != nil {
		return err
	}
	err = binary.Write(w, ble, hdr.Minor)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "{'descr': '%s', 'fortran_order': False, 'shape': %s, }",
		hdr.Descr.Type,
		shapeString(hdr.Descr.Shape),
	)
	var hdrSize = 0
	switch hdr.Major {
	case 1:
		hdrSize = 4 + len(Magic)
	case 2:
		hdrSize = 6 + len(Magic)
	default:
		return fmt.Errorf("npyio: imvalid major version number (%d)", hdr.Major)
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
		err = binary.Write(w, ble, uint16(buflen))
	case 2:
		err = binary.Write(w, ble, uint32(buflen))
	default:
		return fmt.Errorf("npyio: invalid major version number (%d)", hdr.Major)
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

func writeData(w io.Writer, rv reflect.Value) error {
	rt := rv.Type()
	if rt == rtDense {
		m := rv.Interface().(mat64.Dense)
		nrows, ncols := m.Dims()
		var buf [8]byte
		for i := 0; i < nrows; i++ {
			for j := 0; j < ncols; j++ {
				ble.PutUint64(buf[:], math.Float64bits(m.At(i, j)))
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
		ble.PutUint16(buf[:], v)
		_, err := w.Write(buf[:])
		return err

	case []uint16:
		var buf [2]byte
		for _, vv := range v {
			ble.PutUint16(buf[:], vv)
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case uint32:
		var buf [4]byte
		ble.PutUint32(buf[:], v)
		_, err := w.Write(buf[:])
		return err

	case []uint32:
		var buf [4]byte
		for _, vv := range v {
			ble.PutUint32(buf[:], vv)
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case uint64:
		var buf [8]byte
		ble.PutUint64(buf[:], v)
		_, err := w.Write(buf[:])
		return err

	case []uint64:
		var buf [8]byte
		for _, vv := range v {
			ble.PutUint64(buf[:], vv)
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
		ble.PutUint16(buf[:], uint16(v))
		_, err := w.Write(buf[:])
		return err

	case []int16:
		var buf [2]byte
		for _, vv := range v {
			ble.PutUint16(buf[:], uint16(vv))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case int32:
		var buf [4]byte
		ble.PutUint32(buf[:], uint32(v))
		_, err := w.Write(buf[:])
		return err

	case []int32:
		var buf [4]byte
		for _, vv := range v {
			ble.PutUint32(buf[:], uint32(vv))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case int64:
		var buf [8]byte
		ble.PutUint64(buf[:], uint64(v))
		_, err := w.Write(buf[:])
		return err

	case []int64:
		var buf [8]byte
		for _, vv := range v {
			ble.PutUint64(buf[:], uint64(vv))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case float32:
		var buf [4]byte
		ble.PutUint32(buf[:], math.Float32bits(v))
		_, err := w.Write(buf[:])
		return err

	case []float32:
		var buf [4]byte
		for _, v := range v {
			ble.PutUint32(buf[:], math.Float32bits(v))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case float64:
		var buf [8]byte
		ble.PutUint64(buf[:], math.Float64bits(v))
		_, err := w.Write(buf[:])
		return err

	case []float64:
		var buf [8]byte
		for _, v := range v {
			ble.PutUint64(buf[:], math.Float64bits(v))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case complex64:
		var buf [8]byte
		ble.PutUint32(buf[0:4], math.Float32bits(real(v)))
		ble.PutUint32(buf[4:8], math.Float32bits(imag(v)))
		_, err := w.Write(buf[:])
		return err

	case []complex64:
		var buf [8]byte
		for _, v := range v {
			ble.PutUint32(buf[0:4], math.Float32bits(real(v)))
			ble.PutUint32(buf[4:8], math.Float32bits(imag(v)))
			_, err := w.Write(buf[:])
			if err != nil {
				return err
			}
		}
		return nil

	case complex128:
		var buf [16]byte
		ble.PutUint64(buf[0:8], math.Float64bits(real(v)))
		ble.PutUint64(buf[8:16], math.Float64bits(imag(v)))
		_, err := w.Write(buf[:])
		return err

	case []complex128:
		var buf [16]byte
		for _, v := range v {
			ble.PutUint64(buf[0:8], math.Float64bits(real(v)))
			ble.PutUint64(buf[8:16], math.Float64bits(imag(v)))
			_, err := w.Write(buf[:])
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
				err := writeData(w, elem)
				if err != nil {
					return err
				}
			}
			return nil
		default:
			return binary.Write(w, ble, v)
		}

	case reflect.Interface, reflect.String, reflect.Chan, reflect.Map, reflect.Struct:
		return fmt.Errorf("npyio: type %v not supported", rt)
	}

	return binary.Write(w, ble, v)
}

func dtypeFrom(rt reflect.Type) (string, error) {
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

	case reflect.Array, reflect.Slice:
		return dtypeFrom(rt.Elem())

	case reflect.String, reflect.Map, reflect.Chan, reflect.Interface, reflect.Struct:
		return "", fmt.Errorf("npyio: type %v not supported", rt)
	}

	return "", fmt.Errorf("npyio: type %v not supported", rt)
}

func shapeFrom(rv reflect.Value) ([]int, error) {
	if m, ok := rv.Interface().(mat64.Dense); ok {
		nrows, ncols := m.Dims()
		return []int{nrows, ncols}, nil
	}

	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		return []int{rv.Len()}, nil

	case reflect.String, reflect.Map, reflect.Chan, reflect.Interface, reflect.Struct:
		return nil, fmt.Errorf("npyio: type %v not supported", rt)
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
