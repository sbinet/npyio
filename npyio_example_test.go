package npyio_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/gonum/matrix/mat64"
	"github.com/sbinet/npyio"
)

func ExampleWrite() {
	m := mat64.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5})
	fmt.Printf("-- original data --\n")
	fmt.Printf("data = %v\n", mat64.Formatted(m, mat64.Prefix("       ")))
	buf := new(bytes.Buffer)

	err := npyio.Write(buf, m)
	if err != nil {
		log.Fatalf("error writing data: %v\n", err)
	}

	// modify original data
	m.Set(0, 0, 6)

	var data mat64.Dense
	err = npyio.Read(buf, &data)
	if err != nil {
		log.Fatalf("error reading data: %v\n", err)
	}

	fmt.Printf("-- data read back --\n")
	fmt.Printf("data = %v\n", mat64.Formatted(&data, mat64.Prefix("       ")))

	fmt.Printf("-- modified original data --\n")
	fmt.Printf("data = %v\n", mat64.Formatted(m, mat64.Prefix("       ")))

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
	m := mat64.NewDense(2, 3, []float64{0, 1, 2, 3, 4, 5})
	fmt.Printf("-- original data --\n")
	fmt.Printf("data = %v\n", mat64.Formatted(m, mat64.Prefix("       ")))
	buf := new(bytes.Buffer)

	err := npyio.Write(buf, m)
	if err != nil {
		log.Fatalf("error writing data: %v\n", err)
	}

	// modify original data
	m.Set(0, 0, 6)

	var data mat64.Dense
	err = npyio.Read(buf, &data)
	if err != nil {
		log.Fatalf("error reading data: %v\n", err)
	}

	fmt.Printf("-- data read back --\n")
	fmt.Printf("data = %v\n", mat64.Formatted(&data, mat64.Prefix("       ")))

	fmt.Printf("-- modified original data --\n")
	fmt.Printf("data = %v\n", mat64.Formatted(m, mat64.Prefix("       ")))

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
