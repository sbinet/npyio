// Package npyio provides read/write access to files following the NumPy data file format:
//  http://docs.scipy.org/doc/numpy-1.10.1/neps/npy-format.html
//
// npyio supports r/w of scalars, arrays, slices and mat64.Dense.
// Supported scalars are:
//  - bool,
//  - (u)int{,8,16,32,64},
//  - float{32,64},
//  - complex{64,128}
//
// Reading from a NumPy data file can be performed like so:
//
//  f, err := os.Open("data.npy")
//  var m mat64.Dense
//  err = npyio.Read(f, &m)
//  fmt.Printf("data = %v\n", mat64.Formatted(&m, mat64.Prefix("       ")))
//
// npyio can also read data directly into slices, arrays or scalars, provided
// there is a valid type conversion [numpy-data-type]->[go-type].
//
// Example:
//  var data []float64
//  err = npyio.Read(f, &data)
//
//  var data uint64
//  err = npyio.Read(f, &data)
//
//
// Writing into a NumPy data file can be done like so:
//
//  f, err := os.Create("data.npy")
//  var m mat64.Dense = ...
//  err = npyio.Write(f, m)
//
// Scalars, arrays and slices are also supported:
//
//  var data []float64 = ...
//  err = npyio.Write(f, data)
//
//  var data int64 = 42
//  err = npyio.Write(f, data)
//
//  var data [42]complex128 = ...
//  err = npyio.Write(f, data)
package npyio

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	errNilPtr = errors.New("npyio: nil pointer")
	errNotPtr = errors.New("npyio: expected a pointer to a value")
	errDims   = errors.New("npyio: invalid dimensions")
	errNoConv = errors.New("npyio: no legal type conversion")

	ble = binary.LittleEndian

	// ErrInvalidNumPyFormat is the error returned by NewReader when
	// the underlying io.Reader is not a valid or recognized NumPy data
	// file format.
	ErrInvalidNumPyFormat = errors.New("npyio: not a valid NumPy file format")

	// Magic header present at the start of a NumPy data file format.
	// See http://docs.scipy.org/doc/numpy-1.10.1/neps/npy-format.html
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

// newHeader creates a new Header with the major/minor version numbers that npyio currently supports.
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
