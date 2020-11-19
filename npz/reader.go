// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npz

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

	"github.com/sbinet/npyio/npy"
)

// Read reads the item named name from the reader r and
// stores the extracted data into ptr.
func Read(r io.ReaderAt, name string, ptr interface{}) error {
	sz, err := sizeof(r)
	if err != nil {
		return fmt.Errorf("npz: could not retrieve size of reader: %w", err)
	}

	rz, err := NewReader(r, sz)
	if err != nil {
		return fmt.Errorf("npz: could not create npz reader: %w", err)
	}

	err = rz.Read(name, ptr)
	if err != nil {
		return fmt.Errorf("npz: could not read from npz reader: %w", err)
	}
	return nil
}

// Reader reads data from a compressed NumPy data file.
type Reader struct {
	r  io.ReaderAt
	rz *zip.Reader
	rc io.Closer

	keys []string
}

// Open opens the named compressed NumPy data file for reading.
func Open(name string) (*Reader, error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("npz: could not open %q: %w", name, err)
	}
	defer func() {
		if err != nil {
			_ = r.Close()
		}
	}()

	stat, err := r.Stat()
	if err != nil {
		return nil, fmt.Errorf("npz: could not stat %q: %w", name, err)
	}

	rz, err := zip.NewReader(r, stat.Size())
	if err != nil {
		return nil, fmt.Errorf("npz: could not open zip file %q: %w", name, err)
	}

	keys := make([]string, len(rz.File))
	for i, f := range rz.File {
		keys[i] = f.Name
	}

	return &Reader{
		r:    r,
		rz:   rz,
		rc:   r,
		keys: keys,
	}, nil
}

// NewReader reads the compressed NumPy data from r, which is assumed
// to have the given size in bytes.
func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
	rz, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("npz: could not create zip reader: %w", err)
	}

	keys := make([]string, len(rz.File))
	for i, f := range rz.File {
		keys[i] = f.Name
	}

	return &Reader{
		r:    r,
		rz:   rz,
		keys: keys,
	}, nil
}

// Close closes the NumPy compressed data reader.
// Close doesn't close the underlying reader.
func (r *Reader) Close() error {
	if r.rc == nil {
		return nil
	}

	rc := r.rc
	r.rc = nil

	err := rc.Close()
	if err != nil {
		return fmt.Errorf("npz: could not close npz reader: %w", err)
	}

	return nil
}

// Keys returns the names of the NumPy data arrays.
func (r *Reader) Keys() []string {
	return r.keys
}

// Header returns the NumPy header metadata for the named array.
func (r *Reader) Header(name string) *npy.Header {
	elm, err := r.get(name)
	if elm == nil || err != nil {
		return nil
	}
	return elm.hdr()
}

// Open opens the named npy section in the npz archive.
func (r *Reader) Open(name string) (io.ReadCloser, error) {
	return r.open(name)
}

func (r *Reader) open(name string) (io.ReadCloser, error) {
	for _, f := range r.rz.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf(
				"npz: could not open item %q from npz: %w",
				name, err,
			)
		}
		return rc, nil
	}
	return nil, fmt.Errorf("npz: could not find %q", name)
}

func (r *Reader) get(name string) (*ritem, error) {
	rc, err := r.open(name)
	if err != nil {
		return nil, err
	}
	rp, err := npy.NewReader(rc)
	if err != nil {
		_ = rc.Close()
		return nil, fmt.Errorf(
			"npz: could not open npy %q from npz: %w",
			name, err,
		)
	}
	return &ritem{
		r:  rc,
		rp: rp,
	}, nil
}

type ritem struct {
	r  io.ReadCloser
	rp *npy.Reader
}

func (r *ritem) hdr() *npy.Header {
	if r.rp == nil {
		return nil
	}
	return &r.rp.Header
}

func (r *ritem) Close() error {
	if r.r == nil {
		return nil
	}
	err := r.r.Close()
	r.r = nil
	r.rp = nil

	if err != nil {
		return fmt.Errorf("npz: could not close read-item: %w", err)
	}
	return nil
}

// Read reads the named NumPy array data into the provided pointer.
//
// Read returns an error if the on-disk data type and the provided one
// don't match.
func (r *Reader) Read(name string, ptr interface{}) error {
	it, err := r.get(name)
	if err != nil {
		return fmt.Errorf("npz: could not read %q: %w", name, err)
	}
	defer it.Close()

	err = it.rp.Read(ptr)
	if err != nil {
		return fmt.Errorf("npz: could not read %q: %w", name, err)
	}

	return nil
}
