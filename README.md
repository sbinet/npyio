# npyio [![GoDoc](https://godoc.org/github.com/sbinet/npyio?status.svg)](https://godoc.org/github.com/sbinet/npyio)

`npyio` provides read/write access to [numpy data files](http://docs.scipy.org/doc/numpy/neps/npy-format.html).

## Installation

Is done via `go get`:

```sh
$> go get github.com/sbinet/npyio
```

## Documentation

Is available on [godoc](https://godoc.org/github.com/sbinet/npyio)

## Example

### Reading a .npy file

Consider a `.npy` file created with the following `python` code:

```python
>>> import numpy as np
>>> arr = np.arange(6).reshape(2,3)
>>> f = open("data.npy", "w")
>>> np.save(f, arr)
>>> f.close()
```

The (int64) data array can be loaded into a (float64) `mat64.Matrix` by the following code:

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

	m, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("data = %v\n", mat64.Formatted(m, mat64.Prefix("       ")))
}
```

```
$> npyio-ls data.npy
npy-header: Header{Major:1, Minor:0, Descr:{Type:<i8, Fortran:false, Shape:[2 3]}}
data = ⎡0  1  2⎤
       ⎣3  4  5⎦
```

