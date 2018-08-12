// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio_test

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"gonum.org/v1/gonum/mat"

	"github.com/sbinet/npyio"
)

func ExampleWrite() {
	m := mat.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5})
	fmt.Printf("-- original data --\n")
	fmt.Printf("data = %v\n", mat.Formatted(m, mat.Prefix("       ")))
	buf := new(bytes.Buffer)

	err := npyio.Write(buf, m)
	if err != nil {
		log.Fatalf("error writing data: %v\n", err)
	}

	// modify original data
	m.Set(0, 0, 6)

	var data mat.Dense
	err = npyio.Read(buf, &data)
	if err != nil {
		log.Fatalf("error reading data: %v\n", err)
	}

	fmt.Printf("-- data read back --\n")
	fmt.Printf("data = %v\n", mat.Formatted(&data, mat.Prefix("       ")))

	fmt.Printf("-- modified original data --\n")
	fmt.Printf("data = %v\n", mat.Formatted(m, mat.Prefix("       ")))

	// Output:
	// -- original data --
	// data = ⎡0  1  2⎤
	//        ⎣3  4  5⎦
	// -- data read back --
	// data = ⎡0  1  2⎤
	//        ⎣3  4  5⎦
	// -- modified original data --
	// data = ⎡6  1  2⎤
	//        ⎣3  4  5⎦
}

func ExampleRead() {
	m := mat.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5})
	fmt.Printf("-- original data --\n")
	fmt.Printf("data = %v\n", mat.Formatted(m, mat.Prefix("       ")))
	buf := new(bytes.Buffer)

	err := npyio.Write(buf, m)
	if err != nil {
		log.Fatalf("error writing data: %v\n", err)
	}

	// modify original data
	m.Set(0, 0, 6)

	var data mat.Dense
	err = npyio.Read(buf, &data)
	if err != nil {
		log.Fatalf("error reading data: %v\n", err)
	}

	fmt.Printf("-- data read back --\n")
	fmt.Printf("data = %v\n", mat.Formatted(&data, mat.Prefix("       ")))

	fmt.Printf("-- modified original data --\n")
	fmt.Printf("data = %v\n", mat.Formatted(m, mat.Prefix("       ")))

	// Output:
	// -- original data --
	// data = ⎡0  1  2⎤
	//        ⎣3  4  5⎦
	// -- data read back --
	// data = ⎡0  1  2⎤
	//        ⎣3  4  5⎦
	// -- modified original data --
	// data = ⎡6  1  2⎤
	//        ⎣3  4  5⎦
}

func ExamplePartialRead() {
	out, err := os.Create("data.npy")
	if err != nil {
		log.Fatal(err)
	}

	f := []float64{0, 1, 2, 3, 4, 5}
	fmt.Printf("-- original data --\n")
	fmt.Printf("data = %v\n", f)
	err = npyio.Write(out, f)
	if err != nil {
		log.Fatalf("error writing data: %v\n", err)
	}
	out.Close()

	in, err := os.Open("data.npy")
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()
	r, err := npyio.NewReader(in)
	if err != nil {
		log.Fatal(err)
	}

	data := make([]float64, 3)
	err = r.Read(&data)
	if err != nil {
		log.Fatalf("error reading data: %v\n", err)
	}

	fmt.Printf("-- partial data read back --\n")
	fmt.Printf("data = %v\n", data)

	err = r.Read(&data)
	if err != nil {
		log.Fatalf("error reading data: %v\n", err)
	}

	fmt.Printf("-- rest of data read back --\n")
	fmt.Printf("data = %v\n", data)

	// Output:
	// -- original data --
	// data = [0 1 2 3 4 5]
	// -- partial data read back --
	// data = [0 1 2]
	// -- rest of data read back --
	// data = [3 4 5]
}
