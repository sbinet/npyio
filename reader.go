package npyio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

var (
	ble = binary.LittleEndian

	ErrInvalidNumPyFormat = errors.New("npyio: not a valid NumPy file format")

	Magic = [6]byte{'\x93', 'N', 'U', 'M', 'P', 'Y'}
)

type Header struct {
	Major byte
	Minor byte
	Descr struct {
		Type    string
		Fortran bool
		Shape   []int
	}
}

type Reader struct {
	r   io.Reader
	err error // last error

	Header Header
}

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

	descr := string(buf[begDescr+len(descrKey) : begOrder-len(trailer)])
	order := string(buf[begOrder+len(orderKey) : begShape-len(trailer)])
	shape := buf[begShape+len(shapeKey) : endDescr-len(trailer)]
	log.Printf("descr: %q\n", descr)
	log.Printf("order: %q\n", order)
	log.Printf("shape: %q\n", string(shape))

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
		r.Header.Descr.Shape = []int{1}
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

func (r *Reader) read(v interface{}) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, ble, v)
}
