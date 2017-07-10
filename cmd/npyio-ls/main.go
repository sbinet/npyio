// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/sbinet/npyio"
)

var (
	zipMagic = [4]byte{'P', 'K', 3, 4}
)

func main() {
	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}

	allgood := true
	for i, fname := range os.Args[1:] {
		if i > 0 {
			fmt.Printf("\n")
		}
		fmt.Printf(strings.Repeat("=", 80) + "\n")
		fmt.Printf("file: %v\n", fname)

		f, err := os.Open(fname)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			allgood = false
		}
		defer f.Close()

		// detect .npz files (check if we find a ZIP file magic header)
		var hdr [6]byte
		_, err = f.Read(hdr[:])
		if err != nil {
			fmt.Printf("error sniffing file format: %v\n", err)
			allgood = false
			continue
		}
		_, err = f.Seek(0, 0)
		if err != nil {
			fmt.Printf("error rewinding file: %v\n", err)
			allgood = false
			continue
		}

		switch {
		case bytes.Equal(npyio.Magic[:], hdr[:]):
			allgood = display(f, fname) && allgood
		case bytes.Equal(zipMagic[:], hdr[:len(zipMagic)]):
			fi, err := f.Stat()
			if err != nil {
				fmt.Printf("error stat-ing file: %v\n", err)
				allgood = false
				continue
			}
			zr, err := zip.NewReader(f, fi.Size())
			if err != nil {
				fmt.Printf("error creating zip-reader: %v\n", err)
				allgood = false
				continue
			}

			for ii, f := range zr.File {
				r, err := f.Open()
				if err != nil {
					fmt.Printf("error opening entry %s: %v\n", f.Name, err)
					allgood = false
					continue
				}
				defer r.Close()
				if ii > 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("entry: %s\n", f.Name)
				allgood = display(r, fname+"@"+f.Name) && allgood
				err = r.Close()
				if err != nil {
					fmt.Printf("error closing entry %s: %v\n", f.Name, err)
					allgood = false
					continue
				}
			}
		default:
			fmt.Printf("error: unknown magic header %q\n", string(hdr[:]))
			allgood = false
			continue
		}
	}

	if !allgood {
		os.Exit(1)
	}
}

func display(f io.Reader, fname string) bool {
	r, err := npyio.NewReader(f)
	if err != nil {
		fmt.Printf("error creating reader %s: %v\n", fname, err)
		return false
	}

	fmt.Printf("npy-header: %v\n", r.Header)

	rt := npyio.TypeFrom(r.Header.Descr.Type)
	if rt == nil {
		fmt.Printf("error: no reflect.Type for %q\n", r.Header.Descr.Type)
		return false
	}
	rv := reflect.New(reflect.SliceOf(rt))
	err = r.Read(rv.Interface())
	if err != nil && err != io.EOF {
		fmt.Printf("read error: %v\n", err)
		return false
	}
	fmt.Printf("data = %v\n", rv.Elem().Interface())
	return true
}
