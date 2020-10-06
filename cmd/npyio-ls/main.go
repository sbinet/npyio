// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sbinet/npyio"
)

func main() {
	log.SetPrefix("npyio-ls: ")
	log.SetFlags(0)

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

		f, err := os.Open(fname)
		if err != nil {
			log.Printf("could not open %q: %+v", fname, err)
			allgood = false
			continue
		}
		defer f.Close()

		err = npyio.Dump(os.Stdout, f)
		if err != nil {
			log.Printf("could not dump %q: %+v\n", fname, err)
			allgood = false
			continue
		}
	}

	if !allgood {
		os.Exit(1)
	}
}
