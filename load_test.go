package npyio

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name         string
		wantFileName []string
		wantShape    [][]int
		wantData     [][]float64
	}{
		{
			name:         "testdata/data_float64_corder.npz",
			wantFileName: []string{"arr1.npy", "arr0.npy"},
			wantShape: [][]int{
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
			npzData, _, err := Load(f)
			if err != nil {
				t.Fatalf("could not load %q: %+v", tc.name, err)
			}
			var fileNames []string
			for key := range npzData {
				fileNames = append(fileNames, key)
			}
			sort.Slice(fileNames, func(i, j int) bool {
				return fileNames[i] > fileNames[j]
			})
			if got, want := fileNames, tc.wantFileName; !reflect.DeepEqual(got, want) {
				t.Fatalf(
					"filename mismatch:\ngot:\n%v\nwant:\n%v\n",
					got, want,
				)
			}
			for i, fileName := range tc.wantFileName {
				elem := npzData[fileName]
				if got, want := elem.GetHeader().Descr.Shape, tc.wantShape[i]; !reflect.DeepEqual(got, want) {
					t.Fatalf(
						"shape  mismatch in file %s :\ngot:\n%v\nwant:\n%v\n", fileName,
						got, want,
					)
				}
				if got, want := elem.Value, tc.wantData[i]; !reflect.DeepEqual(got, want) {
					t.Fatalf(
						"data  mismatch in file %s :\ngot:\n%v\nwant:\n%v\n", fileName,
						got, want,
					)
				}
			}
		})
	}
}

func TestNumpyElement_ToMatrix(t *testing.T) {
	s := []int16{1, 2, 3, 45, 2, 3, 4, 5}
	a := NumpyElement{
		Value: s,
	}
	a.header.Descr.Shape = []int{4, 2}
	shape := a.header.Descr.Shape
	_, err := a.ToMatrix(false)
	if err == nil {
		t.Fatal("err should be not nil")
	}
	m, err := a.ToMatrix(true)
	if err != nil {
		t.Fatal(err)
	}
	r, c := m.Dims()
	if r != shape[0] && c != shape[1] {
		t.Fatalf("want %v got %v", shape, []int{r, c})
	}
	k := 0
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if m.At(i, j) != float64(s[k]) {
				t.Fatalf("want %d got %f", s[k], m.At(i, j))
			}
			k++
		}
	}
}
