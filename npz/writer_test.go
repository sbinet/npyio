// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npz

import (
	"bytes"
	"reflect"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestWrite(t *testing.T) {
	for _, tc := range []struct {
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
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			wz := NewWriter(buf)
			err := wz.Write(tc.name, tc.want)
			if err != nil {
				t.Fatalf("could not write value: %+v", err)
			}

			err = wz.Close()
			if err != nil {
				t.Fatalf("could not close writer: %+v", err)
			}

			got := reflect.New(reflect.Indirect(reflect.ValueOf(tc.want)).Type())
			err = Read(bytes.NewReader(buf.Bytes()), tc.name, got.Interface())
			if err != nil {
				t.Fatalf("could not read value: %+v", err)
			}

			got = reflect.Indirect(got)
			want := reflect.Indirect(reflect.ValueOf(tc.want))

			if got, want := got.Interface(), want.Interface(); !reflect.DeepEqual(got, want) {
				t.Fatalf(
					"invalid r/w round-trip:\ngot= %#v\nwant=%#v",
					got, want,
				)
			}
		})
	}
}
