// Copyright 2020 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npz_test

import (
	"fmt"
	"log"
	"os"

	"github.com/sbinet/npyio/npz"
)

func ExampleOpen() {
	f, err := npz.Open("../testdata/data_float64_corder.npz")
	if err != nil {
		log.Fatalf("could not open npz file: %+v", err)
	}
	defer f.Close()

	for _, name := range f.Keys() {
		fmt.Printf("%s: %v\n", name, f.Header(name))
	}

	var f0 []float64
	err = f.Read("arr0.npy", &f0)
	if err != nil {
		log.Fatalf("could not read value from npz file: %+v", err)
	}

	var f1 []float64
	err = f.Read("arr1.npy", &f1)
	if err != nil {
		log.Fatalf("could not read value from npz file: %+v", err)
	}

	fmt.Printf("arr0: %v\n", f0)
	fmt.Printf("arr1: %v\n", f1)

	// Output:
	// arr1.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[6 1]}}
	// arr0.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[2 3]}}
	// arr0: [0 1 2 3 4 5]
	// arr1: [0 1 2 3 4 5]
}

func ExampleReader() {
	f, err := os.Open("../testdata/data_float64_corder.npz")
	if err != nil {
		log.Fatalf("could not open npz file: %+v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Fatalf("could not stat npz file: %+v", err)
	}

	r, err := npz.NewReader(f, stat.Size())
	if err != nil {
		log.Fatalf("could not open npz archive: %+v", err)
	}

	for _, name := range r.Keys() {
		fmt.Printf("%s: %v\n", name, r.Header(name))
	}

	var f0 []float64
	err = r.Read("arr0.npy", &f0)
	if err != nil {
		log.Fatalf("could not read value from npz file: %+v", err)
	}

	var f1 []float64
	err = r.Read("arr1.npy", &f1)
	if err != nil {
		log.Fatalf("could not read value from npz file: %+v", err)
	}

	fmt.Printf("arr0: %v\n", f0)
	fmt.Printf("arr1: %v\n", f1)

	// Output:
	// arr1.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[6 1]}}
	// arr0.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[2 3]}}
	// arr0: [0 1 2 3 4 5]
	// arr1: [0 1 2 3 4 5]
}

func ExampleRead() {
	f, err := os.Open("../testdata/data_float64_corder.npz")
	if err != nil {
		log.Fatalf("could not open npz file: %+v", err)
	}
	defer f.Close()

	var f0 []float64
	err = npz.Read(f, "arr0.npy", &f0)
	if err != nil {
		log.Fatalf("could not read value from npz file: %+v", err)
	}

	var f1 []float64
	err = npz.Read(f, "arr1.npy", &f1)
	if err != nil {
		log.Fatalf("could not read value from npz file: %+v", err)
	}

	fmt.Printf("arr0: %v\n", f0)
	fmt.Printf("arr1: %v\n", f1)

	// Output:
	// arr0: [0 1 2 3 4 5]
	// arr1: [0 1 2 3 4 5]
}

func ExampleCreate() {
	f, err := npz.Create("out.npz")
	if err != nil {
		log.Fatalf("could not create npz file: %+v", err)
	}
	defer f.Close()

	err = f.Write("arr0.npy", []float64{0, 1, 2, 3, 4, 5})
	if err != nil {
		log.Fatalf("could not write value arr0.npy to npz file: %+v", err)
	}

	err = f.Write("arr1.npy", []float32{0, 1, 2, 3, 4, 5})
	if err != nil {
		log.Fatalf("could not write value arr1.npy to npz file: %+v", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("could not close npz file: %+v", err)
	}

	// Output:
}

func ExampleWriter() {
	f, err := os.Create("out.npz")
	if err != nil {
		log.Fatalf("could not create npz file: %+v", err)
	}
	defer f.Close()

	wz := npz.NewWriter(f)
	defer wz.Close()

	err = wz.Write("arr0.npy", []float64{0, 1, 2, 3, 4, 5})
	if err != nil {
		log.Fatalf("could not write value arr0.npy to npz file: %+v", err)
	}

	err = wz.Write("arr1.npy", []float32{0, 1, 2, 3, 4, 5})
	if err != nil {
		log.Fatalf("could not write value arr1.npy to npz file: %+v", err)
	}

	err = wz.Close()
	if err != nil {
		log.Fatalf("could not close npz archive: %+v", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("could not close npz file: %+v", err)
	}

	// Output:
}

func ExampleWrite() {
	err := npz.Write("out.npz", map[string]interface{}{
		"arr0.npy": []float64{0, 1, 2, 3, 4, 5},
		"arr1.npy": []float32{0, 1, 2, 3, 4, 5},
	})
	if err != nil {
		log.Fatalf("could not save to npz file: %+v", err)
	}

	// Output:
}
