// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gonum/matrix/mat64"
	"github.com/sbinet/npyio"
)

func main() {
	f, err := os.Open("data.npy")
	if err != nil {
		log.Fatal(err)
	}

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
