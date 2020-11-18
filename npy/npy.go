// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package npy provides read/write access to files following the NumPy data file format:
//  https://numpy.org/neps/nep-0001-npy-format.html
//
// Supported types
//
// npy supports r/w of scalars, arrays, slices and gonum/mat.Dense.
// Supported scalars are:
//  - bool,
//  - (u)int{8,16,32,64},
//  - float{32,64},
//  - complex{64,128}
//
// Reading
//
// Reading from a NumPy data file can be performed like so:
//
//  f, err := os.Open("data.npy")
//  var m mat.Dense
//  err = npy.Read(f, &m)
//  fmt.Printf("data = %v\n", mat.Formatted(&m, mat.Prefix("       "))))
//
// npy can also read data directly into slices, arrays or scalars, provided
// the on-disk data type and the provided one match.
//
// Example:
//  var data []float64
//  err = npy.Read(f, &data)
//
//  var data uint64
//  err = npy.Read(f, &data)
//
// Writing
//
// Writing into a NumPy data file can be done like so:
//
//  f, err := os.Create("data.npy")
//  var m mat.Dense = ...
//  err = npy.Write(f, m)
//
// Scalars, arrays and slices are also supported:
//
//  var data []float64 = ...
//  err = npy.Write(f, data)
//
//  var data int64 = 42
//  err = npy.Write(f, data)
//
//  var data [42]complex128 = ...
//  err = npy.Write(f, data)
package npy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

var (
	errNilPtr = errors.New("npy: nil pointer")
	errNotPtr = errors.New("npy: expected a pointer to a value")
	errDims   = errors.New("npy: invalid dimensions")
	errNoConv = errors.New("npy: no legal type conversion")

	// ErrInvalidNumPyFormat is the error returned by NewReader when
	// the underlying io.Reader is not a valid or recognized NumPy data
	// file format.
	ErrInvalidNumPyFormat = errors.New("npy: not a valid NumPy file format")

	// ErrTypeMismatch is the error returned by Reader when the on-disk
	// data type and the user provided one do NOT match.
	ErrTypeMismatch = errors.New("npy: types don't match")

	// ErrInvalidType is the error returned by Reader and Writer when
	// confronted with a type that is not supported or can not be
	// reliably (de)serialized.
	ErrInvalidType = errors.New("npy: invalid or unsupported type")

	// Magic header present at the start of a NumPy data file format.
	// See https://numpy.org/neps/nep-0001-npy-format.html
	Magic = [6]byte{'\x93', 'N', 'U', 'M', 'P', 'Y'}
)

// Header describes the data content of a NumPy data file.
type Header struct {
	Major byte // data file major version
	Minor byte // data file minor version
	Descr struct {
		Type    string // data type of array elements ('<i8', '<f4', ...)
		Fortran bool   // whether the array data is stored in Fortran-order (col-major)
		Shape   []int  // array shape (e.g. [2,3] a 2-rows, 3-cols array
	}
}

// newHeader creates a new Header with the major/minor version numbers that
// npy currently supports.
func newHeader() Header {
	return Header{
		Major: 2,
		Minor: 0,
	}
}

func (h Header) String() string {
	return fmt.Sprintf("Header{Major:%v, Minor:%v, Descr:{Type:%v, Fortran:%v, Shape:%v}}",
		int(h.Major),
		int(h.Minor),
		h.Descr.Type,
		h.Descr.Fortran,
		h.Descr.Shape,
	)
}

var (
	boolType       = reflect.TypeOf(true)
	uint8Type      = reflect.TypeOf((*uint8)(nil)).Elem()
	uint16Type     = reflect.TypeOf((*uint16)(nil)).Elem()
	uint32Type     = reflect.TypeOf((*uint32)(nil)).Elem()
	uint64Type     = reflect.TypeOf((*uint64)(nil)).Elem()
	int8Type       = reflect.TypeOf((*int8)(nil)).Elem()
	int16Type      = reflect.TypeOf((*int16)(nil)).Elem()
	int32Type      = reflect.TypeOf((*int32)(nil)).Elem()
	int64Type      = reflect.TypeOf((*int64)(nil)).Elem()
	float32Type    = reflect.TypeOf((*float32)(nil)).Elem()
	float64Type    = reflect.TypeOf((*float64)(nil)).Elem()
	complex64Type  = reflect.TypeOf((*complex64)(nil)).Elem()
	complex128Type = reflect.TypeOf((*complex128)(nil)).Elem()
	stringType     = reflect.TypeOf((*string)(nil)).Elem()

	trueUint8  = []byte{1}
	falseUint8 = []byte{0}
)

type dType struct {
	str   string
	utf   bool
	size  int
	order binary.ByteOrder
	rt    reflect.Type
}

func newDtype(str string) (dType, error) {
	var (
		err error
		dt  = dType{
			str:   str,
			order: nativeEndian,
		}
	)
	switch str {
	case "b1", "<b1", "|b1", "bool":
		dt.rt = boolType
		dt.size = 1

	case "u1", "<u1", "|u1", "uint8":
		dt.rt = uint8Type
		dt.size = 1

	case "u2", "<u2", "|u2", ">u2", "uint16":
		dt.rt = uint16Type
		dt.size = 2

	case "u4", "<u4", "|u4", ">u4", "uint32":
		dt.rt = uint32Type
		dt.size = 4

	case "u8", "<u8", "|u8", ">u8", "uint64":
		dt.rt = uint64Type
		dt.size = 8

	case "i1", "<i1", "|i1", ">i1", "int8":
		dt.rt = int8Type
		dt.size = 1

	case "i2", "<i2", "|i2", ">i2", "int16":
		dt.rt = int16Type
		dt.size = 2

	case "i4", "<i4", "|i4", ">i4", "int32":
		dt.rt = int32Type
		dt.size = 4

	case "i8", "<i8", "|i8", ">i8", "int64":
		dt.rt = int64Type
		dt.size = 8

	case "f4", "<f4", "|f4", ">f4", "float32":
		dt.rt = float32Type
		dt.size = 4

	case "f8", "<f8", "|f8", ">f8", "float64":
		dt.rt = float64Type
		dt.size = 8

	case "c8", "<c8", "|c8", ">c8", "complex64":
		dt.rt = complex64Type
		dt.size = 8

	case "c16", "<c16", "|c16", ">c16", "complex128":
		dt.rt = complex128Type
		dt.size = 16
	}

	switch {
	case reStrPre.MatchString(str), reStrPost.MatchString(str):
		dt.rt = stringType
		dt.size, err = stringLen(str)
		if err != nil {
			return dt, err
		}

	case reUniPre.MatchString(str), reUniPost.MatchString(str):
		dt.rt = stringType
		dt.utf = true
		dt.size, err = stringLen(str)
		if err != nil {
			return dt, err
		}
	}
	if dt.rt == nil {
		return dt, fmt.Errorf("npy: no reflect.Type for dtype=%v", str)
	}

	switch dt.str[0] {
	case '<':
		dt.order = binary.LittleEndian
	case '>':
		dt.order = binary.BigEndian
	default:
		dt.order = nativeEndian
	}
	return dt, nil
}

var nativeEndian binary.ByteOrder

func init() {
	v := uint16(1)
	switch byte(v >> 8) {
	case 0:
		nativeEndian = binary.LittleEndian
	case 1:
		nativeEndian = binary.BigEndian
	}
}
