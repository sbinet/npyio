# npyio [![GoDoc](https://godoc.org/github.com/sbinet/npyio?status.svg)](https://godoc.org/github.com/sbinet/npyio)

`npyio` provides read/write access to [numpy data files](http://docs.scipy.org/doc/numpy/neps/npy-format.html).

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

The (float64) data array can be loaded into a (float64) `mat64.Matrix` by the following code:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gonum/matrix/mat64"
	"github.com/sbinet/npyio"
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

	m := mat64.NewDense(shape[0], shape[1], raw)
	fmt.Printf("data = %v\n", mat64.Formatted(m, mat64.Prefix("       ")))
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
