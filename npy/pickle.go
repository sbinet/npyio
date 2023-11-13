// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

//go:generate go run ./gen-pickle.go

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/nlpodyssey/gopickle/pickle"
	py "github.com/nlpodyssey/gopickle/types"
)

// FIXME(sbinet): use errors.ErrUnsupported when Go>=1.21.
var errUnsupported = errors.New("unsupported operation")

func newUnpickler(r io.Reader) pickle.Unpickler {
	u := pickle.NewUnpickler(r)
	u.FindClass = ClassLoader
	return u
}

// ClassLoader provides a python class loader mechanism for python pickles
// containing numpy.dtype and numpy.ndarray values.
func ClassLoader(module, name string) (any, error) {
	switch module + "." + name {
	case "numpy.dtype":
		return &ArrayDescr{}, nil
	case "numpy.ndarray":
		return &Array{}, nil
	case "numpy.core.multiarray._reconstruct":
		return reconstruct{}, nil
	}

	// FIXME(sbinet): use errors.ErrUnsupported when Go>=1.21.
	// return nil, fmt.Errorf("could not unpickle %q: %w", module+"."+name, errors.ErrUnsupported)
	return nil, fmt.Errorf("could not unpickle %q: %w", module+"."+name, errUnsupported)
}

type reconstruct struct{}

var _ py.Callable = (*reconstruct)(nil)

func (reconstruct) Call(args ...any) (any, error) {
	switch sz := len(args); sz {
	case 3:
		// ok.
	default:
		return nil, fmt.Errorf("invalid tuple length (got=%d)", sz)
	}

	var (
		subtype = args[0] // ex: numpy.ndarray
		// shape   = args[1] // a tuple, usually (0,)
		// dtype   = args[2] // a dummy dtype (usually "b")
	)

	switch v := subtype.(type) {
	case py.PyNewable:
		var (
			dtype   = "b"
			shape   []int
			strides []int
			data    []byte
			flags   int
		)
		descr, err := newDescrFrom(dtype, flags)
		if err != nil {
			return nil, fmt.Errorf("could not convert %v (type=%T) to dtype: %w", dtype, dtype, err)
		}
		return v.PyNew(subtype, descr, shape, strides, data, flags)
	}

	return subtype, nil
}

func parseTuple(tup *py.Tuple, args ...any) error {
	if want, got := tup.Len(), len(args); want != got {
		return fmt.Errorf("invalid number of arguments: got=%d, want=%d", got, want)
	}

	for i := range args {
		src := tup.Get(i)
		if src == nil {
			continue
		}
		dst := args[i]
		if dst == nil {
			continue
		}
		rsrc := reflect.Indirect(reflect.ValueOf(src))
		rdst := reflect.Indirect(reflect.ValueOf(dst))
		if !rdst.CanSet() {
			return fmt.Errorf("can not set arg[%d] destination: type=%T", i, dst)
		}
		if rsrc.Type() != rdst.Type() {
			return fmt.Errorf("can not convert arg[%d] %T to %T", i, rsrc.Interface(), rdst.Interface())
		}
		rdst.Set(rsrc)
	}

	return nil
}
