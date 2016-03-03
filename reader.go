package npyio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gonum/matrix/mat64"
)

var (
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

// Read reads the array data from the underlying NumPy data file and
// returns a mat64.Matrix, converting array elements to float64.
//
// Only arrays with up to 2 dimensions are supported.
// Only arrays with elements convertible to float64 are supported.
func (r *Reader) Read() (mat64.Matrix, error) {
	nrows := 0
	ncols := 0

	switch len(r.Header.Descr.Shape) {
	default:
		return nil, fmt.Errorf("npyio: array shape not supported %v", r.Header.Descr.Shape)

	case 0:
		nrows = 1
		ncols = 1

	case 1:
		nrows = r.Header.Descr.Shape[0]
		ncols = 1

	case 2:
		nrows = r.Header.Descr.Shape[0]
		ncols = r.Header.Descr.Shape[1]
	}

	m := mat64.NewDense(nrows, ncols, nil)
	set, err := r.setter(m)
	if err != nil {
		return nil, err
	}

	if r.Header.Descr.Fortran {
		for j := 0; j < ncols; j++ {
			for i := 0; i < nrows; i++ {
				set(i, j)
			}
		}
	} else {
		for i := 0; i < nrows; i++ {
			for j := 0; j < ncols; j++ {
				set(i, j)
			}
		}
	}

	return m, nil
}

func (r *Reader) setter(m *mat64.Dense) (func(i, j int), error) {
	var set func(i, j int)
	switch r.Header.Descr.Type {
	case "<u1":
		var v uint8
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<u2":
		var v uint16
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<u4":
		var v uint32
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<u8":
		var v uint64
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<i1":
		var v int8
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<i2":
		var v int16
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<i4":
		var v int32
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<i8":
		var v int64
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}
	case "<f4":
		var v float32
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, float64(v))
		}

	case "<f8":
		var v float64
		set = func(i, j int) {
			r.read(&v)
			m.Set(i, j, v)
		}
	default:
		return nil, fmt.Errorf("npyio: array dtype not supported %q", r.Header.Descr.Type)
	}
	return set, nil
}

func (r *Reader) read(v interface{}) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, ble, v)
}
