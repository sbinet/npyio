package npyio

import (
	"fmt"
	"os"
	"testing"

	"github.com/gonum/matrix/mat64"
)

func TestReader(t *testing.T) {
	want := map[bool]*mat64.Dense{
		false: mat64.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5}), // row-major
		true:  mat64.NewDense(2, 3, []float64{0, 2, 4, 1, 3, 5}), // col-major
	}

	for _, dt := range []string{
		"float32", "float64",
		"int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64",
	} {
		for _, order := range []string{"f", "c"} {

			fname := fmt.Sprintf("testdata/data_%s_2x3_%sorder.npz", dt, order)
			f, err := os.Open(fname)
			if err != nil {
				t.Errorf("%v: error: %v\n", fname, err)
			}
			defer f.Close()

			r, err := NewReader(f)
			if err != nil {
				t.Errorf("%v: error: %v\n", fname, err)
			}

			m, err := r.Read()
			if err != nil {
				t.Errorf("%v: error: %v\n", fname, err)
			}

			order := r.Header.Descr.Fortran
			if !mat64.Equal(m, want[order]) {
				t.Errorf("%v: error.\n got=%v\nwant=%v\n",
					fname,
					m,
					want[order],
				)
			}
		}
	}
}
