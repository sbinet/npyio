// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{
			name: "testdata/data_float32_2x3_corder.npy",
			want: "testdata/data_float32_2x3_corder.npy.txt",
		},
		{
			name: "testdata/data_float32_2x3_forder.npy",
			want: "testdata/data_float32_2x3_forder.npy.txt",
		},
		{
			name: "testdata/data_float64_2x3x4_corder.npy",
			want: "testdata/data_float64_2x3x4_corder.npy.txt",
		},
		{
			name: "testdata/data_float64_corder.npz",
			want: "testdata/data_float64_corder.npz.txt",
		},
		{
			name: "testdata/data_float64_forder.npz",
			want: "testdata/data_float64_forder.npz.txt",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open(tc.name)
			if err != nil {
				t.Fatalf("could not open %q: %+v", tc.name, err)
			}
			defer f.Close()

			o := new(strings.Builder)
			err = Dump(o, f)
			if err != nil {
				t.Fatalf("could not dump %q: %+v", tc.name, err)
			}

			want, err := ioutil.ReadFile(tc.want)
			if err != nil {
				t.Fatalf("could not read reference file %q: %+v", tc.want, err)
			}

			if got, want := o.String(), string(want); got != want {
				t.Fatalf(
					"invalid dump:\ngot:\n%s\nwant:\n%s\n",
					got, want,
				)
			}
		})
	}
}

func TestDumpSeeker(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{
			name: "testdata/data_float32_2x3_corder.npy",
			want: "testdata/data_float32_2x3_corder.npy.txt",
		},
		{
			name: "testdata/data_float32_2x3_forder.npy",
			want: "testdata/data_float32_2x3_forder.npy.txt",
		},
		{
			name: "testdata/data_float64_2x3x4_corder.npy",
			want: "testdata/data_float64_2x3x4_corder.npy.txt",
		},
		{
			name: "testdata/data_float64_corder.npz",
			want: "testdata/data_float64_corder.npz.txt",
		},
		{
			name: "testdata/data_float64_forder.npz",
			want: "testdata/data_float64_forder.npz.txt",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open(tc.name)
			if err != nil {
				t.Fatalf("could not open %q: %+v", tc.name, err)
			}
			defer f.Close()

			type namer interface{ Name() string }

			r := struct {
				io.Seeker
				io.ReaderAt
				namer
			}{
				Seeker:   f,
				ReaderAt: f,
				namer:    f,
			}
			o := new(strings.Builder)
			err = Dump(o, r)
			if err != nil {
				t.Fatalf("could not dump %q: %+v", tc.name, err)
			}

			want, err := ioutil.ReadFile(tc.want)
			if err != nil {
				t.Fatalf("could not read reference file %q: %+v", tc.want, err)
			}

			if got, want := o.String(), string(want); got != want {
				t.Fatalf(
					"invalid dump:\ngot:\n%s\nwant:\n%s\n",
					got, want,
				)
			}
		})
	}
}

func TestUnzipNpz(t *testing.T) {
	for _, tc := range []struct {
		name         string
		wantFileName []string
		wantShape    [][2]int
		wantData     [][]float64
	}{
		{
			name:         "testdata/data_float64_corder.npz",
			wantFileName: []string{"arr1.npy", "arr0.npy"},
			wantShape: [][2]int{
				{6, 1}, {2, 3},
			},
			wantData: [][]float64{
				{0, 1, 2, 3, 4, 5}, {0, 1, 2, 3, 4, 5},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open(tc.name)
			if err != nil {
				t.Fatalf("could not open %q: %+v", tc.name, err)
			}
			defer f.Close()
			files, err := UnzipNpz(f)
			if err != nil {
				t.Fatalf("could not unzip %q: %+v", tc.name, err)
			}
			for i, file := range files {
				if got, want := file.Name, tc.wantFileName[i]; got != want {
					t.Fatalf(
						"filename mismatch:\ngot:\n%s\nwant:\n%s\n",
						got, want,
					)
				}
				r, err := file.Open()
				if err != nil {
					t.Fatalf("could not open %q: %+v", file.Name, err)
				}
				defer r.Close()
				npyReader, err := NewReader(r)
				if err != nil {
					t.Fatalf("could not read %v: %+v", file.Name, err)
				}
				if got, want := npyReader.Header.Descr.Shape, tc.wantShape[i]; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
					t.Fatalf(
						"shape  mismatch in file %s :\ngot:\n%v\nwant:\n%v\n", file.Name,
						got, want,
					)
				}
				length := npyReader.Header.Descr.Shape[0] * npyReader.Header.Descr.Shape[1]
				var raw = make([]float64, length)
				err = npyReader.Read(&raw)
				if err != nil {
					t.Fatalf("could not read %q: %+v", file.Name, err)
				}
				if got, want := raw, tc.wantData[i]; !equal(got, want) {
					t.Fatalf(
						"data  mismatch in file %s :\ngot:\n%v\nwant:\n%v\n", file.Name,
						got, want,
					)
				}
			}
		})
	}
}

func equal(a, b []float64) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
