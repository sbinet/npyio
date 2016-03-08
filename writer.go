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
		return binary.Write(w, ble, bool2uint(v))
	case int:
		return binary.Write(w, ble, int64(v))
	case uint:
		return binary.Write(w, ble, uint64(v))
	case []bool:
		for _, vv := range v {
			err := binary.Write(w, ble, bool2uint(vv))
			if err != nil {
				return err
			}
		}
		return nil
	case []int:
		for _, vv := range v {
			err := binary.Write(w, ble, int64(vv))
			if err != nil {
				return err
			}
		}
		return nil

	case []uint:
		for _, vv := range v {
			err := binary.Write(w, ble, uint64(vv))
			if err != nil {
				return err
			}
		}
		return nil

	case []uint8:
		_, err := w.Write(v)
		return err
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

func bool2uint(b bool) uint8 {
	if b {
		return 1
	}
	return 0
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
