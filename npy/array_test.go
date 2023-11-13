// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestArrayStringer(t *testing.T) {
	f, err := os.Open("../testdata/data_float64_2x3x4_corder.npy")
	if err != nil {
		t.Fatalf("could not open testdata: %+v", err)
	}
	defer f.Close()

	var arr Array
	err = Read(f, &arr)
	if err != nil {
		t.Fatalf("could not read data: %+v", err)
	}

	var (
		want = `Array{descr: ArrayDescr{kind: 'f', order: '<', flags: 0, esize: 8, align: 8, subarr: <nil>, names: [], fields: {}, meta: map[]}, shape: [2 3 4], strides: [96 32 8], fortran: false, data: [0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23]}`
		got  = fmt.Sprintf("%v", arr)
	)

	if got != want {
		t.Fatalf("invalid array display:\ngot= %s\nwant=%s", got, want)
	}

	if got, want := arr.Descr().kind, byte('f'); got != want {
		t.Fatalf("invalid kind: got=%c, want=%c", got, want)
	}

	if got, want := arr.Shape(), []int{2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("invalid shape:\ngot= %+v\nwant=%+v", got, want)
	}

	if got, want := arr.Strides(), []int{96, 32, 8}; !reflect.DeepEqual(got, want) {
		t.Fatalf("invalid strides:\ngot= %+v\nwant=%+v", got, want)
	}

	if got, want := arr.Fortran(), false; got != want {
		t.Fatalf("invalid fortran:\ngot= %+v\nwant=%+v", got, want)
	}
}
