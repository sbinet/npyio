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

func TestReader(t *testing.T) {
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
	want := map[string][]float64{
		"2x3":    []float64{0, 1, 2, 3, 4, 5},
		"6x1":    []float64{0, 1, 2, 3, 4, 5},
		"1x1":    []float64{42},
		"scalar": []float64{42},
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

				var data []float64
				err = r.Read(&data)
				if err != nil {
					t.Errorf("%v: error: %v\n", fname, err)
				}
				if !reflect.DeepEqual(data, want[shape]) {
					t.Errorf("%v: error.\n got=%v\nwant=%v\n",
						fname,
						data,
						want[shape],
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
