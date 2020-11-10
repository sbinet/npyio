package npyio

import (
	"archive/zip"
	"io"
	"sort"
)

//SaveNPZ just like numpy.savez, Save several arrays into a single file in compressed .npz format.
//filenames are taken from the keywords of map
func SaveNPZ(writer io.Writer, values map[string]interface{}) (err error) {
	zw := zip.NewWriter(writer)
	defer zw.Close()

	fileNames := []string{}
	//sort map keys
	for fileName := range values {
		fileNames = append(fileNames, fileName)
	}
	sort.Slice(fileNames, func(i, j int) bool {
		return fileNames[i] > fileNames[j]
	})
	for _, fileName := range fileNames {
		f, err := zw.Create(fileName)
		if err != nil {
			return err
		}
		err = Write(f, values[fileName])
		if err != nil {
			return err
		}
	}
	return nil
}
