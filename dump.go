// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Dump dumps the content of the provided reader to the writer,
// in a human readable format
func Dump(o io.Writer, r io.ReaderAt) error {
	var fname = "input.npy"
	if r, ok := r.(interface{ Name() string }); ok {
		fname = r.Name()
	}
	fmt.Fprintf(o, strings.Repeat("=", 80)+"\n")
	fmt.Fprintf(o, "file: %v\n", fname)
	npzData, npyData, err := Load(r)
	if err != nil {
		return err
	}
	if npyData != nil {
		fmt.Fprintf(o, "npy-header: %v\n", npyData.GetHeader())
		fmt.Fprintf(o, "data = %v\n", npyData.Value)
	} else {
		//sort file names
		fileNames := []string{}
		for fileName := range npzData {
			fileNames = append(fileNames, fileName)
		}
		sort.Slice(fileNames, func(i, j int) bool {
			return fileNames[i] > fileNames[j]
		})
		i := 0
		for _, fileName := range fileNames {
			element := npzData[fileName]
			if i > 0 {
				fmt.Fprintf(o, "\n")
			}
			fmt.Fprintf(o, "entry: %s\n", fileName)
			fmt.Fprintf(o, "npy-header: %v\n", element.GetHeader())
			fmt.Fprintf(o, "data = %v\n", element.Value)
			i++
		}
	}
	return nil
}

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
		return 0, fmt.Errorf("npyio: unsupported  reader: %T", r)
	}
}
