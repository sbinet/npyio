// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npz

import (
	"fmt"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestReader(t *testing.T) {
	want := map[string]map[bool]*mat.Dense{
		"arr0.npy": {
			false: mat.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5}), // row-major
			true:  mat.NewDense(2, 3, []float64{0, 2, 4, 1, 3, 5}), // col-major
		},
		"arr1.npy": {
			false: mat.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5}),
			true:  mat.NewDense(6, 1, []float64{0, 1, 2, 3, 4, 5}),
		},
	}

	for _, order := range []string{"c", "f"} {
		fname := fmt.Sprintf("../testdata/data_float64_%sorder.npz", order)

		t.Run(fname, func(t *testing.T) {
			zr, err := Open(fname)
			if err != nil {
				t.Fatalf("error: %+v", err)
			}
			defer zr.Close()

			for _, name := range zr.Keys() {
				var m mat.Dense
				err = zr.Read(name, &m)
				if err != nil {
					t.Fatalf("error reading %s data: %+v", name, err)
				}

				corder := zr.Header(name).Descr.Fortran
				if !mat.Equal(&m, want[name][corder]) {
					t.Errorf("%s: error comparing %s.\n got=%v\nwant=%v\n",
						fname,
						name,
						&m,
						want[name][corder],
					)
					continue
				}
			}
		})
	}
}
