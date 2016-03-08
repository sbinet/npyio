// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"io/ioutil"
	"testing"

	"github.com/gonum/matrix/mat64"
)

func BenchmarkWriteDense(b *testing.B) {
	data := make([]float64, 1000)
	m := mat64.NewDense(100, 10, data)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, m)
	}
}

func BenchmarkWriteFloat32Slice(b *testing.B) {
	data := make([]float32, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteFloat64Slice(b *testing.B) {
	data := make([]float64, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteBoolSlice(b *testing.B) {
	data := make([]bool, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteUint8Slice(b *testing.B) {
	data := make([]uint8, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteUint16Slice(b *testing.B) {
	data := make([]uint16, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteUint32Slice(b *testing.B) {
	data := make([]uint32, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteUint64Slice(b *testing.B) {
	data := make([]uint64, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteUintSlice(b *testing.B) {
	data := make([]uint, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteInt8Slice(b *testing.B) {
	data := make([]int8, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteInt16Slice(b *testing.B) {
	data := make([]int16, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteInt32Slice(b *testing.B) {
	data := make([]int32, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteInt64Slice(b *testing.B) {
	data := make([]int64, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteIntSlice(b *testing.B) {
	data := make([]int, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteComplex64Slice(b *testing.B) {
	data := make([]complex64, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteComplex128Slice(b *testing.B) {
	data := make([]complex128, 1000)
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}
