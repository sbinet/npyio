# npyio

`npyio` provides read/write access to [numpy data files](http://docs.scipy.org/doc/numpy/neps/npy-format.html).

## Installation

Is done via `go get`:

```sh
$> go get github.com/sbinet/npyio
```

## Documentation

Is available on [godoc](https://godoc.org/github.com/sbinet/npyio)

## Example

### Reading a .npz file

Consider a `.npz` file created with the following `python` code:

```python
>>> import numpy as np
>>> arr = np.arange(6).reshape(2,3)
>>> f = open("data.npz", "w")
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

	"github.com/sbinet/npyio"
)

func main() {
	f, err := os.Open("data.npz")
	if err != nil {
		log.Fatal(err)
	}

	r, err := npyio.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("npz-header: %#v\n", r.Header)

	m, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}

	nrows, ncols := m.Dims()
	for i := 0; i < nrows; i++ {
		for j := 0; j < ncols; j++ {
			fmt.Printf("data[%d][%d]= %v\n", i, j, m.At(i, j))
		}
	}
}```

```
$> npyio-read data.npz
npz-header: npyio.Header{Major:0x1, Minor:0x0, Descr:struct { Type string; Fortran bool; Shape []int }{Type:"<i8", Fortran:false, Shape:[]int{2, 3}}}
data[0][0]= 0
data[0][1]= 1
data[0][2]= 2
data[1][0]= 3
data[1][1]= 4
data[1][2]= 5
```

