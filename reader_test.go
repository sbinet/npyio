package npyio

import (
	"fmt"
	"os"
	"testing"
)

func TestReader(t *testing.T) {
	f, err := os.Open("testdata/data_float64.npz")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	defer f.Close()

	dec, err := NewReader(f)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}

	fmt.Printf("dec: %v\n", dec)
}
