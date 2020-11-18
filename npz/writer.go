// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npz

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/sbinet/npyio/npy"
)

// Write writes the values vs to the named npz archive file.
//
// The data-array will always be written out in C-order (row-major).
func Write(name string, vs map[string]interface{}) error {
	w, err := Create(name)
	if err != nil {
		return err
	}
	defer w.Close()

	ks := make([]string, 0, len(vs))
	for k := range vs {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	for _, k := range ks {
		err = w.Write(k, vs[k])
		if err != nil {
			return err
		}
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

// Writer writes data to a compressed NumPy data file.
type Writer struct {
	w  io.Writer
	wz *zip.Writer
	wc io.Closer
}

// Create creates the named compressed NumPy data file for writing.
func Create(name string) (*Writer, error) {
	w, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("npz: could not create %q: %w", name, err)
	}

	wz := zip.NewWriter(w)

	return &Writer{
		w:  w,
		wz: wz,
		wc: w,
	}, nil
}

// NewWriter returns a new npz writer.
//
// The returned npz writer won't close the underlying writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:  w,
		wz: zip.NewWriter(w),
	}
}

// Close closes the npz archive.
// Close flushes the data to disk.
func (w *Writer) Close() error {
	if w.w == nil {
		return nil
	}

	var (
		errz error
		errc error
	)

	errz = w.wz.Close()
	if w.wc != nil {
		wc := w.wc
		w.wc = nil
		errc = wc.Close()
	}

	w.w = nil
	w.wz = nil

	if errz != nil {
		return fmt.Errorf("npz: could not close npz archive: %w", errz)
	}

	if errc != nil {
		return fmt.Errorf("npz: could not close npz file: %w", errc)
	}

	return nil
}

// Write writes the named NumPy array data to the npz archive.
func (w *Writer) Write(name string, v interface{}) error {
	ww, err := w.wz.Create(name)
	if err != nil {
		return fmt.Errorf("npz: could not create npz entry %q: %w", name, err)
	}

	err = npy.Write(ww, v)
	if err != nil {
		return fmt.Errorf("npz: could not write npz entry %q: %w", name, err)
	}

	return nil
}
