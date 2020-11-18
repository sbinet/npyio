# npyio

[![Build Status](https://travis-ci.org/sbinet/npyio.svg?branch=master)](https://travis-ci.org/sbinet/npyio)
[![codecov](https://codecov.io/gh/sbinet/npyio/branch/master/graph/badge.svg)](https://codecov.io/gh/sbinet/npyio)
[![Go Report Card](https://goreportcard.com/badge/github.com/sbinet/npyio)](https://goreportcard.com/report/github.com/sbinet/npyio)
[![License](https://img.shields.io/badge/License-BSD--3-blue.svg)](https://github.com/sbinet/npyio/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/sbinet/npyio?status.svg)](https://godoc.org/github.com/sbinet/npyio)

`npyio` provides read/write access to [numpy data files](https://numpy.org/neps/nep-0001-npy-format.html).

## Installation

Is done via `go get`:

```sh
$> go get github.com/sbinet/npyio
```

## Documentation

Is available on [godoc](https://godoc.org/github.com/sbinet/npyio)

## npyio-ls

`npyio-ls` is a command using `github.com/sbinet/npyio` (located under
`github.com/sbinet/npyio/cmd/npyio-ls`) to display the content of a (list of)
`NumPy` data file(s).

```
$> npyio-ls testdata/data_float64_2x3_?order.npy 
================================================================================
file: testdata/data_float64_2x3_corder.npy
npy-header: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[2 3]}}
data = [0 1 2 3 4 5]

================================================================================
file: testdata/data_float64_2x3_forder.npy
npy-header: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:true, Shape:[2 3]}}
data = [0 1 2 3 4 5]

$> npyio-ls testdata/data_float64_2x3x4_corder.npy 
================================================================================
file: testdata/data_float64_2x3x4_corder.npy
npy-header: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[2 3 4]}}
data = [0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23]
```

`npyio-ls` automatically detects `.npz` archive files and inspects them too:

```
$> npyio-ls testdata/data_float64_corder.npz 
================================================================================
file: testdata/data_float64_corder.npz
entry: arr1.npy
npy-header: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[6 1]}}
data = [0 1 2 3 4 5]

entry: arr0.npy
npy-header: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[2 3]}}
data = [0 1 2 3 4 5]
```

## Example

### Reading a .npy file

Consider a `.npy` file created with the following `python` code:

```python
>>> import numpy as np
>>> arr = np.arange(6, dtype="float64").reshape(2,3)
>>> f = open("data.npy", "w")
>>> np.save(f, arr)
>>> f.close()
```

The (float64) data array can be loaded into a (float64) `mat.Matrix` by the following code:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sbinet/npyio"
	"gonum.org/v1/gonum/mat"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r, err := npyio.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("npy-header: %v\n", r.Header)
	shape := r.Header.Descr.Shape
	raw := make([]float64, shape[0]*shape[1])

	err = r.Read(&raw)
	if err != nil {
		log.Fatal(err)
	}

	m := mat.NewDense(shape[0], shape[1], raw)
	fmt.Printf("data = %v\n", mat.Formatted(m, mat.Prefix("       ")))
}
```

```
$> my-binary data.npy
npy-header: Header{Major:1, Minor:0, Descr:{Type:<i8, Fortran:false, Shape:[2 3]}}
data = ⎡0  1  2⎤
       ⎣3  4  5⎦
```

### Reading a .npy file with npyio.Read

Alternatively, one can use the convenience function `npyio.Read`:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sbinet/npyio"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var m []float64
	err = npyio.Read(f, &m)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("data = %v\n", m)
}
```

```
$> my-binary ./data.npy
data = [0 1 2 3 4 5]
```

### Writing a .npy file with npyio.Write

```go
package main

import (
	"log"
	"os"

	"github.com/sbinet/npyio"
)

func main() {
	f, err := os.Create("data.npy")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	m := []float64{0, 1, 2, 3, 4, 5}
	err = npyio.Write(f, m)
	if err != nil {
		log.Fatalf("error writing to file: %v\n", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("error closing file: %v\n", err)
	}
}
```

### Reading a .npz file with npyio/npz

[embedmd]:# (npz/npz_example_test.go go /func ExampleOpen/ /\n}/)
```go
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
	// arr0.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[2 3]}}
	// arr1.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[6 1]}}
	// arr0: [0 1 2 3 4 5]
	// arr1: [0 1 2 3 4 5]
}
```

[embedmd]:# (npz/npz_example_test.go go /func ExampleReader/ /\n}/)
```go
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
	// arr0.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[2 3]}}
	// arr1.npy: Header{Major:1, Minor:0, Descr:{Type:<f8, Fortran:false, Shape:[6 1]}}
	// arr0: [0 1 2 3 4 5]
	// arr1: [0 1 2 3 4 5]
}
```

[embedmd]:# (npz/npz_example_test.go go /func ExampleRead\(/ /\n}/)
```go
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
```

### Writing a .npz file with npyio/npz

[embedmd]:# (npz/npz_example_test.go go /func ExampleCreate/ /\n}/)
```go
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
```

[embedmd]:# (npz/npz_example_test.go go /func ExampleWriter/ /\n}/)
```go
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
```

[embedmd]:# (npz/npz_example_test.go go /func ExampleWrite\(/ /\n}/)
```go
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
```
