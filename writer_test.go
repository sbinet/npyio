// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/gonum/matrix/mat64"
)

func TestWriter(t *testing.T) {
	for _, test := range []struct {
		name string
		want interface{}
	}{
		{"dense_2x3", mat64.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5})},
		{"dense_6x1", mat64.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5})},
		{"dense_1x6", mat64.NewDense(1, 6, []float64{0, 1, 2, 3, 4, 5})},
		{"dense_1x1", mat64.NewDense(1, 1, []float64{42})},

		// scalars
		{"bool-true", true},
		{"bool-false", false},
		{"uint", uint(42)},
		{"uint8", uint8(42)},
		{"uint16", uint16(42)},
		{"uint32", uint32(42)},
		{"uint64", uint64(42)},
		{"int", int(42)},
		{"int8", int8(42)},
		{"int16", int16(42)},
		{"int32", int32(42)},
		{"int64", int64(42)},
		{"float32", float32(42)},
		{"float64", float64(42)},
		{"cplx64", complex64(42 + 66i)},
		{"cplx128", complex128(42 + 66i)},

		// arrays
		{"bool-array", [6]bool{true, true, false, false, true, false}},
		{"uint-array", [6]uint{0, 1, 2, 3, 4, 5}},
		{"uint8-array", [6]uint8{0, 1, 2, 3, 4, 5}},
		{"uint16-array", [6]uint16{0, 1, 2, 3, 4, 5}},
		{"uint32-array", [6]uint32{0, 1, 2, 3, 4, 5}},
		{"uint64-array", [6]uint64{0, 1, 2, 3, 4, 5}},
		{"int-array", [6]int{0, 1, 2, 3, 4, 5}},
		{"int8-array", [6]int8{0, 1, 2, 3, 4, 5}},
		{"int16-array", [6]int16{0, 1, 2, 3, 4, 5}},
		{"int32-array", [6]int32{0, 1, 2, 3, 4, 5}},
		{"int64-array", [6]int64{0, 1, 2, 3, 4, 5}},
		{"float32-array", [6]float32{0, 1, 2, 3, 4, 5}},
		{"float64-array", [6]float64{0, 1, 2, 3, 4, 5}},
		{"cplx64-array", [6]complex64{0, 1 + 1i, 2 + 2i, 3 + 3i, 4 + 4i, 5 + 5i}},
		{"cplx128-array", [6]complex128{0, 1 + 1i, 2 + 2i, 3 + 3i, 4 + 4i, 5 + 5i}},

		// slices
		{"bool-slice", []bool{true, true, false, false, true, false}},
		{"uint-slice", []uint{0, 1, 2, 3, 4, 5}},
		{"uint8-slice", []uint8{0, 1, 2, 3, 4, 5}},
		{"uint16-slice", []uint16{0, 1, 2, 3, 4, 5}},
		{"uint32-slice", []uint32{0, 1, 2, 3, 4, 5}},
		{"uint64-slice", []uint64{0, 1, 2, 3, 4, 5}},
		{"int-slice", []int{0, 1, 2, 3, 4, 5}},
		{"int8-slice", []int8{0, 1, 2, 3, 4, 5}},
		{"int16-slice", []int16{0, 1, 2, 3, 4, 5}},
		{"int32-slice", []int32{0, 1, 2, 3, 4, 5}},
		{"int64-slice", []int64{0, 1, 2, 3, 4, 5}},
		{"float32-slice", []float32{0, 1, 2, 3, 4, 5}},
		{"float64-slice", []float64{0, 1, 2, 3, 4, 5}},
		{"cplx64-slice", []complex64{0, 1 + 1i, 2 + 2i, 3 + 3i, 4 + 4i, 5 + 5i}},
		{"cplx128-slice", []complex128{0, 1 + 1i, 2 + 2i, 3 + 3i, 4 + 4i, 5 + 5i}},
	} {
		buf := new(bytes.Buffer)
		err := Write(buf, test.want)
		if err != nil {
			t.Errorf("%v: error writing data: %v\n", test.name, err)
		}

		got := reflect.New(reflect.Indirect(reflect.ValueOf(test.want)).Type())
		err = Read(buf, got.Interface())
		if err != nil {
			t.Errorf("%v: error reading data: %v\n", test.name, err)
		}

		want := reflect.Indirect(reflect.ValueOf(test.want))
		rv := reflect.Indirect(got)
		if !reflect.DeepEqual(rv.Interface(), want.Interface()) {
			t.Errorf("%v: error.\n got=%v\nwant=%v\n", test.name, rv.Interface(), want.Interface())
		}
	}
}

func TestWriterNaNsInf(t *testing.T) {
	want := mat64.NewDense(4, 1, []float64{math.NaN(), math.Inf(-1), 0, math.Inf(+1)})

	buf := new(bytes.Buffer)
	err := Write(buf, want)
	if err != nil {
		t.Errorf("error writing data: %v\n", err)
	}

	var m mat64.Dense
	err = Read(buf, &m)
	if err != nil {
		t.Errorf("error reading data: %v\n", err)
	}

	for i, v := range []bool{
		math.IsNaN(m.At(0, 0)),
		math.IsInf(m.At(1, 0), -1),
		m.At(2, 0) == 0,
		math.IsInf(m.At(3, 0), +1),
	} {
		if !v {
			t.Errorf("read test m.At(%d,0) failed\n got=%#v\nwant=%#v\n", i, m.At(i, 0), want.At(i, 0))
		}
	}
}
