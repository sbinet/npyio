// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

import (
	"fmt"
	"strings"

	py "github.com/nlpodyssey/gopickle/types"
)

// Array is a multidimensional, homogeneous array of fixed-size items.
type Array struct {
	descr   ArrayDescr
	shape   []int
	strides []int
	fortran bool

	data any
}

var (
	_ py.PyNewable       = (*Array)(nil)
	_ py.PyStateSettable = (*Array)(nil)
)

func (*Array) PyNew(args ...any) (any, error) {
	var (
		subtype = args[0]
		descr   = args[1].(*ArrayDescr)
		shape   = args[2].([]int)
		strides = args[3].([]int)
		data    = args[4].([]byte)
		flags   = args[5].(int)
	)

	return newArray(subtype, *descr, shape, strides, data, flags)
}

func newArray(subtype any, descr ArrayDescr, shape, strides []int, data []byte, flags int) (*Array, error) {
	switch subtype := subtype.(type) {
	case *Array:
		// ok.
	default:
		return nil, fmt.Errorf("subtyping ndarray with %T is not (yet?) supported", subtype)
	}

	arr := &Array{
		descr:   descr,
		shape:   shape,
		strides: strides,
		data:    data,
	}
	return arr, nil
}

func (arr *Array) PySetState(arg any) error {
	tuple, ok := arg.(*py.Tuple)
	if !ok {
		return fmt.Errorf("invalid argument type %T", arg)
	}

	var (
		vers  = 0
		shape py.Tuple
		raw   any
	)
	switch tuple.Len() {
	case 5:
		err := parseTuple(tuple, &vers, &shape, &arr.descr, &arr.fortran, nil)
		if err != nil {
			return fmt.Errorf("could not parse ndarray.__setstate__ tuple: %w", err)
		}
		raw = tuple.Get(4)
	case 4:
		err := parseTuple(tuple, &shape, &arr.descr, &arr.fortran, nil)
		if err != nil {
			return fmt.Errorf("could not parse ndarray.__setstate__ tuple: %w", err)
		}
		raw = tuple.Get(3)
	default:
		return fmt.Errorf("invalid length (%d) for ndarray.__setstate__ tuple", tuple.Len())
	}

	arr.shape = nil
	for i := range shape {
		v, ok := shape.Get(i).(int)
		if !ok {
			return fmt.Errorf("invalid shape[%d]: got=%T, want=int", i, shape.Get(i))
		}
		arr.shape = append(arr.shape, v)
	}

	err := arr.setupStrides()
	if err != nil {
		return fmt.Errorf("ndarray.__setstate__ could not infer strides: %w", err)
	}

	switch raw := raw.(type) {
	case *py.List:
		arr.data = raw

	case []byte:
		data, err := arr.descr.unmarshal(raw, arr.shape)
		if err != nil {
			return fmt.Errorf("ndarray.__setstate__ could not unmarshal raw data: %w", err)
		}
		arr.data = data
	}

	return nil
}

func (arr *Array) setupStrides() error {
	// TODO(sbinet): complete implementation.
	// see: _array_fill_strides in numpy/_core/multiarray/ctors.c

	if arr.shape == nil {
		arr.strides = nil
		return nil
	}

	strides := make([]int, len(arr.shape))
	// FIXME(sbinet): handle non-contiguous arrays
	// FIXME(sbinet): handle FORTRAN arrays

	var (
		// notCFContig bool
		noDim bool // a dimension != 1 was found
	)

	// check if array is both FORTRAN- and C-contiguous
	for _, dim := range arr.shape {
		if dim != 1 {
			if noDim {
				//	notCFContig = true
				break
			}
			noDim = true
		}
	}

	itemsize := arr.descr.itemsize()
	switch {
	case arr.fortran:
		for i, dim := range arr.shape {
			strides[i] = itemsize
			switch {
			case dim != 0:
				itemsize *= dim
			default:
				// notCFContig = false
			}
		}

	default:
		for i := len(arr.shape) - 1; i >= 0; i-- {
			dim := arr.shape[i]
			strides[i] = itemsize
			switch {
			case dim != 0:
				itemsize *= dim
			default:
				// notCFContig = false
			}
		}
	}

	arr.strides = strides
	return nil
}

// Descr returns the array's data type descriptor.
func (arr Array) Descr() ArrayDescr {
	return arr.descr
}

// Shape returns the array's shape.
func (arr Array) Shape() []int {
	return arr.shape
}

// Strides returns the array's strides in bytes.
func (arr Array) Strides() []int {
	return arr.strides
}

// Fortran returns whether the array's data is stored in FORTRAN-order
// (ie: column-major) instead of C-order (ie: row-major.)
func (arr Array) Fortran() bool {
	return arr.fortran
}

// Data returns the array's underlying data.
func (arr Array) Data() any {
	return arr.data
}

func (arr Array) String() string {
	o := new(strings.Builder)
	fmt.Fprintf(o, "Array{descr: %v, ", arr.descr)
	switch arr.shape {
	case nil:
		fmt.Fprintf(o, "shape: nil, ")
	default:
		fmt.Fprintf(o, "shape: %v, ", arr.shape)
	}
	switch arr.strides {
	case nil:
		fmt.Fprintf(o, "strides: nil, ")
	default:
		fmt.Fprintf(o, "strides: %v, ", arr.strides)
	}
	fmt.Fprintf(o, "fortran: %v, data: %+v}",
		arr.fortran,
		arr.data,
	)
	return o.String()
}
