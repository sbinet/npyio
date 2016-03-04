package npyio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/gonum/matrix/mat64"
)

var (
	errNilPtr = errors.New("npyio: nil pointer")
	errNotPtr = errors.New("npyio: expected a pointer to a value")
	errDims   = errors.New("npyio: invalid dimensions")
	errNoConv = errors.New("npyio: no legal type conversion")

	ble = binary.LittleEndian

	// ErrInvalidNumPyFormat is the error returned by NewReader when
	// the underlying io.Reader is not a valid or recognized NumPy data
	// file format.
	ErrInvalidNumPyFormat = errors.New("npyio: not a valid NumPy file format")

	// Magic header present at the start of a NumPy data file format.
	// See http://docs.scipy.org/doc/numpy-1.10.1/neps/npy-format.html
	Magic = [6]byte{'\x93', 'N', 'U', 'M', 'P', 'Y'}
)

// Header describes the data content of a NumPy data file.
type Header struct {
	Major byte // data file major version
	Minor byte // data file minor version
	Descr struct {
		Type    string // data type of array elements ('<i8', '<f4', ...)
		Fortran bool   // whether the array data is stored in Fortran-order (col-major)
		Shape   []int  // array shape (e.g. [2,3] a 2-rows, 3-cols array
	}
}

func (h Header) String() string {
	return fmt.Sprintf("Header{Major:%v, Minor:%v, Descr:{Type:%v, Fortran:%v, Shape:%v}}",
		int(h.Major),
		int(h.Minor),
		h.Descr.Type,
		h.Descr.Fortran,
		h.Descr.Shape,
	)
}

// Reader reads data from a NumPy data file.
type Reader struct {
	r   io.Reader
	err error // last error

	Header Header
}

// NewReader creates a new NumPy data file format reader.
func NewReader(r io.Reader) (*Reader, error) {
	rr := &Reader{r: r}
	rr.readHeader()
	if rr.err != nil {
		return nil, rr.err
	}
	return rr, rr.err
}

func (r *Reader) readHeader() {
	if r.err != nil {
		return
	}
	var magic [6]byte
	r.read(&magic)
	if r.err != nil {
		return
	}
	if magic != Magic {
		r.err = ErrInvalidNumPyFormat
		return
	}

	var hdrLen int

	r.read(&r.Header.Major)
	r.read(&r.Header.Minor)
	switch r.Header.Major {
	case 1:
		var v uint16
		r.read(&v)
		hdrLen = int(v)
	case 2:
		var v uint32
		r.read(&v)
		hdrLen = int(v)
	default:
		r.err = fmt.Errorf("npyio: invalid major version number (%d)", r.Header.Major)
	}

	if r.err != nil {
		return
	}

	hdr := make([]byte, hdrLen)
	r.read(&hdr)
	idx := bytes.LastIndexByte(hdr, '\n')
	hdr = hdr[:idx]
	r.readDescr(hdr)
}

func (r *Reader) readDescr(buf []byte) {
	if r.err != nil {
		return
	}

	var (
		descrKey = []byte("'descr': ")
		orderKey = []byte("'fortran_order': ")
		shapeKey = []byte("'shape': ")
		trailer  = []byte(", ")
	)

	begDescr := bytes.Index(buf, descrKey)
	begOrder := bytes.Index(buf, orderKey)
	begShape := bytes.Index(buf, shapeKey)
	endDescr := bytes.Index(buf, []byte("}"))
	if begDescr < 0 || begOrder < 0 || begShape < 0 {
		r.err = fmt.Errorf("npyio: invalid dictionary format")
		return
	}

	descr := string(buf[begDescr+len(descrKey)+1 : begOrder-len(trailer)-1])
	order := string(buf[begOrder+len(orderKey) : begShape-len(trailer)])
	shape := buf[begShape+len(shapeKey) : endDescr-len(trailer)]

	r.Header.Descr.Type = descr // FIXME(sbinet): better handling
	switch order {
	case "False":
		r.Header.Descr.Fortran = false
	case "True":
		r.Header.Descr.Fortran = true
	default:
		r.err = fmt.Errorf("npyio: invalid 'fortran_order' value (%v)", order)
		return
	}

	if string(shape) == "()" {
		r.Header.Descr.Shape = nil
		return
	}

	shape = shape[1 : len(shape)-1]
	toks := strings.Split(string(shape), ",")
	for _, tok := range toks {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		i, err := strconv.Atoi(tok)
		if err != nil {
			r.err = err
			return
		}
		r.Header.Descr.Shape = append(r.Header.Descr.Shape, int(i))
	}

}

// Read reads the numpy-array data from the underlying NumPy file and
// converts the array elements to the given pointed at value.
//
// If a *mat64.Dense matrix is passed to Read, the numpy-array data is loaded
// into the Dense matrix, honouring Fortran/C-order.
//
// Only numpy-arrays with up to 2 dimensions are supported.
// Only numpy-arrays with elements convertible to float64 are supported.
func (r *Reader) Read(ptr interface{}) error {
	rv := reflect.ValueOf(ptr)
	if !rv.IsValid() || rv.Kind() != reflect.Ptr {
		return errNotPtr
	}

	if rv.IsNil() {
		return errNilPtr
	}

	nelems := numElems(r.Header.Descr.Shape)
	dt := typeFromDType(r.Header.Descr.Type)
	if dt == nil {
		return fmt.Errorf("npyio: no reflect.Type for dtype=%v", r.Header.Descr.Type)
	}

	switch ptr.(type) {
	case *mat64.Dense:
		var data []float64
		err := r.Read(&data)
		if err != nil {
			return err
		}
		nrows, ncols, err := dimsFromShape(r.Header.Descr.Shape)
		if err != nil {
			return err
		}
		var v *mat64.Dense
		if r.Header.Descr.Fortran {
			v = mat64.NewDense(nrows, ncols, nil)
			i := 0
			for icol := 0; icol < ncols; icol++ {
				for irow := 0; irow < nrows; irow++ {
					v.Set(irow, icol, data[i])
					i++
				}
			}
		} else {
			v = mat64.NewDense(nrows, ncols, data)
		}
		rv.Elem().Set(reflect.ValueOf(v).Elem())
		return nil
	}

	rv = rv.Elem()
	switch rv.Kind() {
	case reflect.Slice:
		rv.SetLen(0)
		elt := rv.Type().Elem()
		v := reflect.New(dt).Elem()
		slice := rv
		for i := 0; i < nelems; i++ {
			r.read(v.Addr().Interface())
			slice = reflect.Append(slice, v.Convert(elt))
		}
		rv.Set(slice)
		return nil

	case reflect.Array:
		if nelems > rv.Type().Len() {
			return errDims
		}
		elt := rv.Type().Elem()
		v := reflect.New(dt).Elem()
		for i := 0; i < nelems; i++ {
			r.read(v.Addr().Interface())
			rv.Index(i).Set(v.Convert(elt))
		}
		return nil

	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		v := reflect.New(dt).Elem()
		if !dt.ConvertibleTo(rv.Type()) {
			return errNoConv
		}
		r.read(v.Addr().Interface())
		rv.Set(v.Convert(rv.Type()))
		return nil

	case reflect.String, reflect.Map, reflect.Chan:
		return fmt.Errorf("npyio: type %v not supported", rv.Addr().Type())
	}

	return nil
}

func dimsFromShape(shape []int) (int, int, error) {
	nrows := 0
	ncols := 0

	switch len(shape) {
	default:
		return -1, -1, fmt.Errorf("npyio: array shape not supported %v", shape)

	case 0:
		nrows = 1
		ncols = 1

	case 1:
		nrows = shape[0]
		ncols = 1

	case 2:
		nrows = shape[0]
		ncols = shape[1]
	}

	return nrows, ncols, nil
}

func (r *Reader) read(v interface{}) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, ble, v)
}

func numElems(shape []int) int {
	n := 1
	for _, v := range shape {
		n *= v
	}
	return n
}

func typeFromDType(dtype string) reflect.Type {
	dt := dtype
	switch dt[0] {
	case '<', '|', '>', '=':
		dt = dt[1:]
	}
	switch dt {
	case "u1":
		return reflect.TypeOf(uint8(0))
	case "u2":
		return reflect.TypeOf(uint16(0))
	case "u4":
		return reflect.TypeOf(uint32(0))
	case "u8":
		return reflect.TypeOf(uint64(0))
	case "i1":
		return reflect.TypeOf(int8(0))
	case "i2":
		return reflect.TypeOf(int16(0))
	case "i4":
		return reflect.TypeOf(int32(0))
	case "i8":
		return reflect.TypeOf(int64(0))
	case "f4":
		return reflect.TypeOf(float32(0))
	case "f8":
		return reflect.TypeOf(float64(0))
	}
	return nil
}
