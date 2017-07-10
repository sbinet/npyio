// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"archive/zip"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/gonum/matrix/mat64"
)

func TestReaderDense(t *testing.T) {
	want := map[string]map[bool]*mat64.Dense{
		"2x3": map[bool]*mat64.Dense{
			false: mat64.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5}), // row-major
			true:  mat64.NewDense(2, 3, []float64{0, 2, 4, 1, 3, 5}), // col-major
		},
		"6x1": map[bool]*mat64.Dense{
			false: mat64.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5}),
			true:  mat64.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5}),
		},
		"1x1": map[bool]*mat64.Dense{
			false: mat64.NewDense(1, 1, []float64{42}),
			true:  mat64.NewDense(1, 1, []float64{42}),
		},
		"scalar": map[bool]*mat64.Dense{
			false: mat64.NewDense(1, 1, []float64{42}),
			true:  mat64.NewDense(1, 1, []float64{42}),
		},
	}

	for _, dt := range []string{
		"float64",
	} {
		for _, order := range []string{"f", "c"} {
			for _, shape := range []string{"2x3", "6x1", "1x1", "scalar"} {

				fname := fmt.Sprintf("testdata/data_%s_%s_%sorder.npy", dt, shape, order)
				f, err := os.Open(fname)
				if err != nil {
					t.Errorf("%v: error: %v\n", fname, err)
				}
				defer f.Close()

				r, err := NewReader(f)
				if err != nil {
					t.Errorf("%v: error: %v\n", fname, err)
				}

				var m mat64.Dense
				err = r.Read(&m)
				if err != nil {
					t.Errorf("%v: error: %v\n", fname, err)
				}

				order := r.Header.Descr.Fortran
				if !mat64.Equal(&m, want[shape][order]) {
					t.Errorf("%v: error.\n got=%v\nwant=%v\n",
						fname,
						&m,
						want[shape][order],
					)
				}
			}
		}
	}
}

func TestReaderSlice(t *testing.T) {
	want := map[string]map[string]interface{}{
		"float32": {
			"2x3":    []float32{0, 1, 2, 3, 4, 5},
			"6x1":    []float32{0, 1, 2, 3, 4, 5},
			"1x1":    []float32{42},
			"scalar": []float32{42},
		},
		"float64": {
			"2x3":    []float64{0, 1, 2, 3, 4, 5},
			"6x1":    []float64{0, 1, 2, 3, 4, 5},
			"1x1":    []float64{42},
			"scalar": []float64{42},
		},
		"int8": {
			"2x3":    []int8{0, 1, 2, 3, 4, 5},
			"6x1":    []int8{0, 1, 2, 3, 4, 5},
			"1x1":    []int8{42},
			"scalar": []int8{42},
		},
		"int16": {
			"2x3":    []int16{0, 1, 2, 3, 4, 5},
			"6x1":    []int16{0, 1, 2, 3, 4, 5},
			"1x1":    []int16{42},
			"scalar": []int16{42},
		},
		"int32": {
			"2x3":    []int32{0, 1, 2, 3, 4, 5},
			"6x1":    []int32{0, 1, 2, 3, 4, 5},
			"1x1":    []int32{42},
			"scalar": []int32{42},
		},
		"int64": {
			"2x3":    []int64{0, 1, 2, 3, 4, 5},
			"6x1":    []int64{0, 1, 2, 3, 4, 5},
			"1x1":    []int64{42},
			"scalar": []int64{42},
		},
		"uint8": {
			"2x3":    []uint8{0, 1, 2, 3, 4, 5},
			"6x1":    []uint8{0, 1, 2, 3, 4, 5},
			"1x1":    []uint8{42},
			"scalar": []uint8{42},
		},
		"uint16": {
			"2x3":    []uint16{0, 1, 2, 3, 4, 5},
			"6x1":    []uint16{0, 1, 2, 3, 4, 5},
			"1x1":    []uint16{42},
			"scalar": []uint16{42},
		},
		"uint32": {
			"2x3":    []uint32{0, 1, 2, 3, 4, 5},
			"6x1":    []uint32{0, 1, 2, 3, 4, 5},
			"1x1":    []uint32{42},
			"scalar": []uint32{42},
		},
		"uint64": {
			"2x3":    []uint64{0, 1, 2, 3, 4, 5},
			"6x1":    []uint64{0, 1, 2, 3, 4, 5},
			"1x1":    []uint64{42},
			"scalar": []uint64{42},
		},
	}

	for _, dt := range []string{
		"float32", "float64",
		"int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64",
	} {
		for _, order := range []string{"f", "c"} {
			for _, shape := range []string{"2x3", "6x1", "1x1", "scalar"} {

				fname := fmt.Sprintf("testdata/data_%s_%s_%sorder.npy", dt, shape, order)
				f, err := os.Open(fname)
				if err != nil {
					t.Errorf("%v: error: %v\n", fname, err)
				}
				defer f.Close()

				r, err := NewReader(f)
				if err != nil {
					t.Errorf("%v: error: %v\n", fname, err)
				}

				rt := TypeFrom(dt)
				if rt == nil {
					t.Errorf("%v: no reflect type for %v\n", fname, dt)
					continue
				}
				data := reflect.New(reflect.SliceOf(rt))
				err = r.Read(data.Interface())
				if err != nil {
					t.Errorf("%v: error: %v\n", fname, err)
				}
				if !reflect.DeepEqual(data.Elem().Interface(), want[dt][shape]) {
					t.Errorf("%v: error.\n got=%v\nwant=%v\n",
						fname,
						data.Elem().Interface(),
						want[dt][shape],
					)
				}
			}
		}
	}
}

func TestReaderNDimSlice(t *testing.T) {
	want := make([]float64, 2*3*4)
	for i := range want {
		want[i] = float64(i)
	}

	f, err := os.Open("testdata/data_float64_2x3x4_corder.npy")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	defer f.Close()

	var data []float64
	err = Read(f, &data)
	if err != nil {
		t.Errorf("error reading data: %v\n", err)
	}

	if !reflect.DeepEqual(data, want) {
		t.Errorf("error.\n got=%v\nwant=%v\n", data, want)
	}
}

func TestReaderNaNsInf(t *testing.T) {
	want := mat64.NewDense(4, 1, []float64{math.NaN(), math.Inf(-1), 0, math.Inf(+1)})
	f, err := os.Open("testdata/nans_inf.npy")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	defer f.Close()

	var m mat64.Dense
	err = Read(f, &m)
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

func TestReaderNpz(t *testing.T) {
	want := map[string]map[bool]*mat64.Dense{
		"arr0.npy": map[bool]*mat64.Dense{
			false: mat64.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5}), // row-major
			true:  mat64.NewDense(2, 3, []float64{0, 2, 4, 1, 3, 5}), // col-major
		},
		"arr1.npy": map[bool]*mat64.Dense{
			false: mat64.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5}),
			true:  mat64.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5}),
		},
	}

	for _, order := range []string{"c", "f"} {
		fname := fmt.Sprintf("testdata/data_float64_%sorder.npz", order)

		zr, err := zip.OpenReader(fname)
		if err != nil {
			t.Errorf("%s: error: %v\n", fname, err)
			continue
		}
		defer zr.Close()

		for _, zip := range zr.File {
			f, err := zip.Open()
			if err != nil {
				t.Errorf("%s: error opening %s entry: %v\n", fname, zip.Name, err)
				continue
			}
			defer f.Close()

			r, err := NewReader(f)
			if err != nil {
				t.Errorf("%s: error creating %s reader: %v\n", fname, zip.Name, err)
				continue
			}

			var m mat64.Dense
			err = r.Read(&m)
			if err != nil {
				t.Errorf("%s: error reading %s data: %v\n", fname, zip.Name, err)
				continue
			}

			corder := r.Header.Descr.Fortran
			if !mat64.Equal(&m, want[zip.Name][corder]) {
				t.Errorf("%s: error comparing %s.\n got=%v\nwant=%v\n",
					fname,
					zip.Name,
					&m,
					want[zip.Name][corder],
				)
				continue
			}
		}
	}
}

func TestStringLenDtype(t *testing.T) {
	for _, test := range []struct {
		dtype string
		want  int
		err   bool
	}{
		{
			dtype: "S66",
			want:  66,
		},
		{
			dtype: "S6",
			want:  6,
		},
		{
			dtype: "6S",
			want:  6,
		},
		{
			dtype: "66S",
			want:  66,
		},
		{
			dtype: "|S6",
			want:  6,
		},
		{
			dtype: "|S66",
			want:  66,
		},
		{
			dtype: "|6S",
			want:  6,
		},
		{
			dtype: "|66S",
			want:  66,
		},
		{
			dtype: "a6",
			want:  6,
		},
		{
			dtype: "6a",
			want:  6,
		},
		{
			dtype: "|a6",
			want:  6,
		},
		{
			dtype: "|6a",
			want:  6,
		},
		{
			dtype: "<U25",
			want:  25,
		},
		{
			dtype: "|U25",
			want:  25,
		},
		{
			dtype: ">U25",
			want:  25,
		},
		{
			dtype: "<25U",
			want:  25,
		},
		{
			dtype: "|25U",
			want:  25,
		},
		{
			dtype: ">25U",
			want:  25,
		},
		{
			dtype: "6S6",
			err:   true,
		},
		{
			dtype: "6a6",
			err:   true,
		},
		{
			dtype: "6U6",
			err:   true,
		},
		{
			dtype: "<i4",
			err:   true,
		},
	} {
		n, err := stringLen(test.dtype)
		if err == nil && test.err {
			t.Errorf("%s: expected an error", test.dtype)
			continue
		}
		if err != nil && !test.err {
			t.Errorf("%s: error=%v", test.dtype, err)
			continue
		}
		if n != test.want {
			t.Errorf("%s: got=%d. want=%d", test.dtype, n, test.want)
			continue
		}
	}
}
