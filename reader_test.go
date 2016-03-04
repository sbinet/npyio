package npyio

import (
	"fmt"
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
