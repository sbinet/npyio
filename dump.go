// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Dump dumps the content of the provided reader to the writer,
// in a human readable format
func Dump(o io.Writer, r io.ReaderAt) error {
	var (
		err      error
		zipMagic = [4]byte{'P', 'K', 3, 4}
		fname    = "input.npy"
	)

	if r, ok := r.(interface{ Name() string }); ok {
		fname = r.Name()
	}

	fmt.Fprintf(o, strings.Repeat("=", 80)+"\n")
	fmt.Fprintf(o, "file: %v\n", fname)

	// detect .npz files (check if we find a ZIP file magic header)
	var hdr [6]byte
	_, err = r.ReadAt(hdr[:], 0)
	if err != nil {
		return fmt.Errorf("npyio: could not infer format: %w", err)
	}

	sizeof := func(r io.ReaderAt) (int64, error) {
		switch r := r.(type) {
		case interface{ Stat() (os.FileInfo, error) }:
			fi, err := r.Stat()
			if err != nil {
				return 0, err
			}
			return fi.Size(), nil
		case io.Seeker:
			pos, err := r.Seek(0, io.SeekCurrent)
			if err != nil {
				return 0, err
			}
			sz, err := r.Seek(0, io.SeekEnd)
			if err != nil {
				return 0, err
			}
			_, err = r.Seek(pos, io.SeekStart)
			if err != nil {
				return 0, err
			}
			return sz, nil
		default:
			return 0, fmt.Errorf("npyio: unsupported reader: %T", r)
		}
	}

	sz, err := sizeof(r)
	if err != nil {
		return fmt.Errorf("npyio: could not infer file size: %w", err)
	}

	switch {
	case bytes.Equal(Magic[:], hdr[:]):
		err = display(o, io.NewSectionReader(r, 0, sz), fname)
		if err != nil {
			return fmt.Errorf("npyio: could not display ile: %w", err)
		}

	case bytes.Equal(zipMagic[:], hdr[:len(zipMagic)]):
		zr, err := zip.NewReader(r, sz)
		if err != nil {
			return fmt.Errorf("npyio: could not create zip file reader: %w", err)
		}

		for ii, f := range zr.File {
			r, err := f.Open()
			if err != nil {
				return fmt.Errorf(
					"npyio: could not open zip file entry %s: %w",
					f.Name, err,
				)
			}
			defer r.Close()
			if ii > 0 {
				fmt.Fprintf(o, "\n")
			}
			fmt.Fprintf(o, "entry: %s\n", f.Name)
			err = display(o, r, fname+"@"+f.Name)
			if err != nil {
				return fmt.Errorf(
					"npyio: could not display zip file entry %s: %w",
					f.Name, err,
				)
			}
			err = r.Close()
			if err != nil {
				return fmt.Errorf(
					"npyio: could not close zip file entry %s: %w",
					f.Name, err,
				)
			}
		}
	default:
		return fmt.Errorf("npyio: unknown magic header %q", string(hdr[:]))
	}

	return nil
}

func display(o io.Writer, f io.Reader, fname string) error {
	r, err := NewReader(f)
	if err != nil {
		return fmt.Errorf("npyio: could not create npy reader %s: %w", fname, err)
	}

	fmt.Fprintf(o, "npy-header: %v\n", r.Header)

	rt := TypeFrom(r.Header.Descr.Type)
	if rt == nil {
		return fmt.Errorf("npyio: no reflect type for %q", r.Header.Descr.Type)
	}
	rv := reflect.New(reflect.SliceOf(rt))
	err = r.Read(rv.Interface())
	if err != nil && err != io.EOF {
		return fmt.Errorf("npyio: read error: %w", err)
	}
	fmt.Fprintf(o, "data = %v\n", rv.Elem().Interface())
	return nil
}
