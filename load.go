package npyio

import (
	"archive/zip"
	"bytes"
	"fmt"
	"gonum.org/v1/gonum/mat"
	"io"
	"reflect"
)

//NumpyElement the data read from  a NumPy data file.
type NumpyElement struct {
	Value  interface{} // data  values ,arrays ,slice,other types
	header Header      // numpy npy file header
}

//ToMatrix convert npy data to mat.Dense
//if forceTypeConvert if true , convert number types like  int,uint,float32 to float64
func (n *NumpyElement) ToMatrix(forceTypeConvert bool) (*mat.Dense, error) {
	shape := n.header.Descr.Shape
	if len(shape) == 0 {
		return nil, fmt.Errorf("shape is nil")
	}
	var rowLen = shape[0]
	dataLen := 1
	for _, i2 := range shape {
		dataLen *= i2
	}
	data := make([]float64, dataLen)
	if t, ok := n.Value.([]float64); ok {
		return mat.NewDense(rowLen, dataLen/rowLen, n.Value.([]float64)), nil
	} else {
		if !forceTypeConvert {
			return nil, fmt.Errorf("cannot conver %type to []float64 ", reflect.TypeOf(t))
		}
	}

	switch n.Value.(type) {
	case []int:
		t := n.Value.([]int)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []uint:
		t := n.Value.([]uint)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []int8:
		t := n.Value.([]int8)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []uint8:
		t := n.Value.([]uint8)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []int16:
		t := n.Value.([]int16)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []uint16:
		t := n.Value.([]uint16)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []int32:
		t := n.Value.([]int32)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []uint32:
		t := n.Value.([]uint32)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []uint64:
		t := n.Value.([]uint64)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []int64:
		t := n.Value.([]int64)
		for i, v := range t {
			data[i] = float64(v)
		}
	case []float32:
		t := n.Value.([]float32)
		for i, v := range t {
			data[i] = float64(v)
		}
	default:
		return nil, fmt.Errorf("cannot conver %type to []float64 ", reflect.TypeOf(n.Value))

	}
	m := mat.NewDense(rowLen, dataLen/rowLen, data)
	return m, nil
}

//GetHeader get header to
func (n *NumpyElement) GetHeader() Header {
	return n.header
}

//Load  just like numpy.load ,load data from io reader
//If the file is a .npy file, then  a single numpy element npyData   is returned.
//If the file is a .npz file, then a map npzData is returned , containing  {filename:numpy element}
//one for each file in the archive.
func Load(r io.ReaderAt) (npzData map[string]*NumpyElement, npyData *NumpyElement, err error) {
	var (
		zipMagic = [4]byte{'P', 'K', 3, 4}
		fname    = ""
	)
	if r, ok := r.(interface{ Name() string }); ok {
		fname = r.Name()
	}
	// detect .npz files (check if we find a ZIP file magic header)
	var hdr [6]byte
	_, err = r.ReadAt(hdr[:], 0)
	if err != nil {
		return nil, nil, fmt.Errorf("npyio: could not infer format: %w", err)
	}

	sz, err := sizeof(r)
	if err != nil {
		return nil, nil, fmt.Errorf("npyio: could not infer file size: %w", err)
	}

	switch {
	// .npy file
	case bytes.Equal(Magic[:], hdr[:]):
		elem, err := readNumpyElement(io.NewSectionReader(r, 0, sz), fname)
		if err != nil {
			return nil, nil, fmt.Errorf("npyio: could not display ile: %w", err)

		}
		return nil, elem, nil
		//.npz file
	case bytes.Equal(zipMagic[:], hdr[:len(zipMagic)]):
		npzData = make(map[string]*NumpyElement)
		zr, err := zip.NewReader(r, sz)
		if err != nil {
			return nil, nil, fmt.Errorf("npyio: could not create zip file reader: %w", err)
		}

		for _, f := range zr.File {
			r, err := f.Open()
			if err != nil {
				return nil, nil, fmt.Errorf(
					"npyio: could not open zip file entry %s: %w",
					f.Name, err,
				)
			}
			defer r.Close()
			elem, err := readNumpyElement(r, f.Name)
			if err != nil {
				return nil, nil, fmt.Errorf(
					"npyio: could not display zip file entry %s: %w",
					f.Name, err,
				)
			}
			err = r.Close()
			if err != nil {
				return nil, nil, fmt.Errorf(
					"npyio: could not close zip file entry %s: %w",
					f.Name, err,
				)
			}
			npzData[f.Name] = elem
		}
		return npzData, nil, nil
	default:
		return nil, nil, fmt.Errorf("npyio: unknown magic header %q", string(hdr[:]))
	}

}

func readNumpyElement(f io.Reader, fname string) (*NumpyElement, error) {
	r, err := NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("npyio: could not create npy reader %s: %w", fname, err)
	}
	rt := TypeFrom(r.Header.Descr.Type)
	if rt == nil {
		return nil, fmt.Errorf("npyio: no reflect type for %q", r.Header.Descr.Type)
	}
	rv := reflect.New(reflect.SliceOf(rt))
	err = r.Read(rv.Interface())
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("npyio: read error: %w", err)
	}
	elem := &NumpyElement{
		Value:  rv.Elem().Interface(),
		header: r.Header,
	}
	return elem, nil
}
