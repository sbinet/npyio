// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package npz provides read/write access to files with compressed NumPy data
// file format:
//
//  https://numpy.org/neps/nep-0001-npy-format.html
//
package npz

import (
	"fmt"
	"io"
	"os"
)

func sizeof(r io.ReaderAt) (int64, error) {
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
		return 0, fmt.Errorf("npz: unsupported reader: %T", r)
	}
}
