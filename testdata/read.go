// +build ignore

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

	fmt.Printf("npz-header: %v\n", r.Header)

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
}
