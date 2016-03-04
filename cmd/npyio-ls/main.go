package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gonum/matrix/mat64"
	"github.com/sbinet/npyio"
)

func main() {
	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(1)
	}

	for i, fname := range os.Args[1:] {
		if i > 0 {
			fmt.Printf("\n")
		}
		f, err := os.Open(fname)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		r, err := npyio.NewReader(f)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(strings.Repeat("=", 80) + "\n")
		fmt.Printf("file: %v\n", fname)
		fmt.Printf("npy-header: %v\n", r.Header)

		var m mat64.Dense
		err = r.Read(&m)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("data = %v\n", mat64.Formatted(&m, mat64.Prefix("       ")))
	}
}
