package npyio

import (
	"gonum.org/v1/gonum/mat"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestSaveZ(t *testing.T) {
	for _, tc := range []struct {
		name         string
		wantFileName []string
		wantShape    [][]int
		wantData     [][]float64
	}{
		{
			name:         ".data_float64_forder.npz",
			wantFileName: []string{"arr_0.npy", "arr_1.npy", "arr_2.npy"},
			wantShape: [][]int{
				{6}, {2, 3}, {3, 3},
			},
			wantData: [][]float64{
				{0, 1, 2, 3, 4, 5}, {0, 1, 2, 3, 4, 5}, {1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var values = make(map[string]interface{})
			for i, val := range tc.wantData {
				if len(tc.wantShape[i]) == 2 {
					m := mat.NewDense(tc.wantShape[i][0], tc.wantShape[i][1], val)
					values[tc.wantFileName[i]] = m
				} else {
					values[tc.wantFileName[i]] = val
				}
			}
			f, err := os.Create(tc.name)
			if err != nil {
				t.Fatalf("could not open %q: %+v", tc.name, err)
			}
			err = SaveNPZ(f, values)
			f.Close()
			defer os.Remove(tc.name)
			if err != nil {
				t.Fatalf("could not write   %+v", err)
			}
			fr, err := os.Open(tc.name)
			if err != nil {
				t.Fatalf("could not open %q: %+v", tc.name, err)
			}
			defer fr.Close()

			npzData, _, err := Load(fr)
			if err != nil {
				t.Fatalf("could not read   %+v", err)
			}
			var fileNames []string
			for key := range npzData {
				fileNames = append(fileNames, key)
			}
			sort.Strings(fileNames)
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
