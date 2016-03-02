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

type Decoder struct {
	r   io.Reader
	err error // last error

	Header Header
}

func NewDecoder(r io.Reader) (*Decoder, error) {
	dec := &Decoder{r: r}
	dec.decodeHeader()
	if dec.err != nil {
		return nil, dec.err
	}
	return dec, dec.err
}

func (dec *Decoder) decodeHeader() {
	if dec.err != nil {
		return
	}
	var magic [6]byte
	dec.read(&magic)
	if dec.err != nil {
		return
	}
	if magic != Magic {
		dec.err = ErrInvalidNumPyFormat
		return
	}

	var hdrLen int

	dec.read(&dec.Header.Major)
	dec.read(&dec.Header.Minor)
	switch dec.Header.Major {
	case 1:
		var v uint16
		dec.read(&v)
		hdrLen = int(v)
	case 2:
		var v uint32
		dec.read(&v)
		hdrLen = int(v)
	default:
		dec.err = fmt.Errorf("npyio: invalid major version number (%d)", dec.Header.Major)
	}

	if dec.err != nil {
		return
	}

	hdr := make([]byte, hdrLen)
	dec.read(&hdr)
	idx := bytes.LastIndexByte(hdr, '\n')
	hdr = hdr[:idx]
	dec.decodeDescr(hdr)
}

func (dec *Decoder) decodeDescr(buf []byte) {
	if dec.err != nil {
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
		dec.err = fmt.Errorf("npyio: invalid dictionary format")
		return
	}

	descr := string(buf[begDescr+len(descrKey) : begOrder-len(trailer)])
	order := string(buf[begOrder+len(orderKey) : begShape-len(trailer)])
	shape := buf[begShape+len(shapeKey) : endDescr-len(trailer)]
	log.Printf("descr: %q\n", descr)
	log.Printf("order: %q\n", order)
	log.Printf("shape: %q\n", string(shape))

	dec.Header.Descr.Type = descr // FIXME(sbinet): better handling
	switch order {
	case "False":
		dec.Header.Descr.Fortran = false
	case "True":
		dec.Header.Descr.Fortran = true
	default:
		dec.err = fmt.Errorf("npyio: invalid 'fortran_order' value (%v)", order)
		return
	}

	if string(shape) == "()" {
		dec.Header.Descr.Shape = []int{1}
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
			dec.err = err
			return
		}
		dec.Header.Descr.Shape = append(dec.Header.Descr.Shape, int(i))
	}

}

func (dec *Decoder) read(v interface{}) {
	if dec.err != nil {
		return
	}
	dec.err = binary.Read(dec.r, ble, v)
}
