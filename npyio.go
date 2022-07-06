// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate embedmd -w README.md

// Package npyio provides read/write access to files following the NumPy data file format:
//
//	https://numpy.org/neps/nep-0001-npy-format.html
//
// # Supported types
//
// npyio supports r/w of scalars, arrays, slices and gonum/mat.Dense.
// Supported scalars are:
//   - bool,
//   - (u)int{8,16,32,64},
//   - float{32,64},
//   - complex{64,128}
//
// # Reading
//
// Reading from a NumPy data file can be performed like so:
//
//	f, err := os.Open("data.npy")
//	var m mat.Dense
//	err = npyio.Read(f, &m)
//	fmt.Printf("data = %v\n", mat.Formatted(&m, mat.Prefix("       "))))
//
// npyio can also read data directly into slices, arrays or scalars, provided
// the on-disk data type and the provided one match.
//
// Example:
//
//	var data []float64
//	err = npyio.Read(f, &data)
//
//	var data uint64
//	err = npyio.Read(f, &data)
//
// # Writing
//
// Writing into a NumPy data file can be done like so:
//
//	f, err := os.Create("data.npy")
//	var m mat.Dense = ...
//	err = npyio.Write(f, m)
//
// Scalars, arrays and slices are also supported:
//
//	var data []float64 = ...
//	err = npyio.Write(f, data)
//
//	var data int64 = 42
//	err = npyio.Write(f, data)
//
//	var data [42]complex128 = ...
//	err = npyio.Write(f, data)
package npyio

import (
	"io"
	"reflect"

	"github.com/sbinet/npyio/npy"
)

var (
	// ErrInvalidNumPyFormat is the error returned by NewReader when
	// the underlying io.Reader is not a valid or recognized NumPy data
	// file format.
	ErrInvalidNumPyFormat = npy.ErrInvalidNumPyFormat

	// ErrTypeMismatch is the error returned by Reader when the on-disk
	// data type and the user provided one do NOT match.
	ErrTypeMismatch = npy.ErrTypeMismatch

	// ErrInvalidType is the error returned by Reader and Writer when
	// confronted with a type that is not supported or can not be
	// reliably (de)serialized.
	ErrInvalidType = npy.ErrInvalidType

	// Magic header present at the start of a NumPy data file format.
	// See https://numpy.org/neps/nep-0001-npy-format.html
	Magic = npy.Magic
)

// Header describes the data content of a NumPy data file.
type Header = npy.Header

// Reader reads data from a NumPy data file.
type Reader = npy.Reader

// NewReader creates a new NumPy data file format reader.
func NewReader(r io.Reader) (*Reader, error) {
	return npy.NewReader(r)
}

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
	return npy.Read(r, ptr)
}

// TypeFrom returns the reflect.Type corresponding to the numpy-dtype string, if any.
func TypeFrom(dtype string) reflect.Type {
	return npy.TypeFrom(dtype)
}

// Write writes 'val' into 'w' in the NumPy data format.
//
//   - if val is a scalar, it must be of a supported type (bools, (u)ints, floats and complexes)
//   - if val is a slice or array, it must be a slice/array of a supported type.
//     the shape (len,) will be written out.
//   - if val is a mat.Dense, the correct shape will be transmitted. (ie: (nrows, ncols))
//
// The data-array will always be written out in C-order (row-major).
func Write(w io.Writer, val interface{}) error {
	return npy.Write(w, val)
}
