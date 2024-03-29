// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestWriter(t *testing.T) {
	for _, test := range []struct {
		name string
		want interface{}
	}{
		{"dense_2x3", mat.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5})},
		{"dense_6x1", mat.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5})},
		{"dense_1x6", mat.NewDense(1, 6, []float64{0, 1, 2, 3, 4, 5})},
		{"dense_1x1", mat.NewDense(1, 1, []float64{42})},

		// scalars
		{"bool-true", true},
		{"bool-false", false},
		{"uint8", uint8(42)},
		{"uint16", uint16(42)},
		{"uint32", uint32(42)},
		{"uint64", uint64(42)},
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
		{"uint8-array", [6]uint8{0, 1, 2, 3, 4, 5}},
		{"uint16-array", [6]uint16{0, 1, 2, 3, 4, 5}},
		{"uint32-array", [6]uint32{0, 1, 2, 3, 4, 5}},
		{"uint64-array", [6]uint64{0, 1, 2, 3, 4, 5}},
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
		{"uint8-slice", []uint8{0, 1, 2, 3, 4, 5}},
		{"uint16-slice", []uint16{0, 1, 2, 3, 4, 5}},
		{"uint32-slice", []uint32{0, 1, 2, 3, 4, 5}},
		{"uint64-slice", []uint64{0, 1, 2, 3, 4, 5}},
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
	want := mat.NewDense(4, 1, []float64{math.NaN(), math.Inf(-1), 0, math.Inf(+1)})

	buf := new(bytes.Buffer)
	err := Write(buf, want)
	if err != nil {
		t.Errorf("error writing data: %v\n", err)
	}

	var m mat.Dense
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

func TestShapeFrom(t *testing.T) {
	for _, tc := range []struct {
		v    interface{}
		want []int
		err  error
	}{
		{
			v:    "hello",
			want: nil,
		},
		{
			v:    1,
			want: nil,
		},
		{
			v:    0.1,
			want: nil,
		},
		{
			v:    [0]int{},
			want: []int{0},
		},
		{
			v:    []int{},
			want: []int{0},
		},
		{
			v:    []int{1},
			want: []int{1},
		},
		{
			v:    [3][]int{nil, nil, nil},
			want: []int{3, 0},
		},
		{
			v:    [3][]int{{}, {}, {}},
			want: []int{3, 0},
		},
		{
			v:    [3][]int{{1, 2}, {3, 4}, {5, 6}},
			want: []int{3, 2},
		},
		{
			v:    [][]int{nil, nil, nil},
			want: []int{3, 0},
		},
		{
			v:    [][]int{{}, {}, {}},
			want: []int{3, 0},
		},
		{
			v:    [][]int{{1, 2}, {3, 4}, {5, 6}},
			want: []int{3, 2},
		},
		{
			v:    [][][]int{{{1}, {2}}, {{3}, {4}}, {{5}, {6}}},
			want: []int{3, 2, 1},
		},
		{
			v:    [][]float64{{1, 2}, {3, 4}, {5, 6}},
			want: []int{3, 2},
		},
		{
			v:    mat.NewDense(2, 3, []float64{1, 2, 3, 4, 5, 6}),
			want: nil, // shapeFrom takes a deref-iface
		},
		{
			v:    *mat.NewDense(2, 3, []float64{1, 2, 3, 4, 5, 6}),
			want: []int{2, 3},
		},
		{
			v:   make(map[int]int),
			err: fmt.Errorf("npy: type map[int]int not supported"),
		},
		{
			v:   make(chan int),
			err: fmt.Errorf("npy: type chan int not supported"),
		},
		{
			v:   struct{}{},
			err: fmt.Errorf("npy: type struct {} not supported"),
		},
	} {
		t.Run("", func(t *testing.T) {
			got, err := shapeFrom(reflect.ValueOf(tc.v))
			switch {
			case err != nil && tc.err != nil:
				if err.Error() != tc.err.Error() {
					t.Fatalf("invalid error:\ngot= %+v\nwant=%+v",
						err, tc.err,
					)
				}
				return
			case err != nil && tc.err == nil:
				t.Fatalf("unexpected error: %+v", err)
			case err == nil && tc.err != nil:
				t.Fatalf("expected an error")
			case err == nil && tc.err == nil:
				// ok.
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("invalid shape.\ngot= %+v\nwant=%+v", got, tc.want)
			}
		})
	}
}
