// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"io/ioutil"
	"testing"
)

func BenchmarkWriteFloat32Slice(b *testing.B) {
	data := make([]float32, 1000)
	for i := 0; i < len(data); i++ {
		data[i] = float32(i)
	}

	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteFloat64Slice(b *testing.B) {
	data := make([]float64, 1000)
	for i := 0; i < len(data); i++ {
		data[i] = float64(i)
	}

	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteInt16Slice(b *testing.B) {
	data := make([]int16, 1000)
	for i := 0; i < len(data); i++ {
		data[i] = int16(i)
	}

	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteInt32Slice(b *testing.B) {
	data := make([]int32, 1000)
	for i := 0; i < len(data); i++ {
		data[i] = int32(i)
	}

	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteInt64Slice(b *testing.B) {
	data := make([]int64, 1000)
	for i := 0; i < len(data); i++ {
		data[i] = int64(i)
	}

	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}

func BenchmarkWriteIntSlice(b *testing.B) {
	data := make([]int, 1000)
	for i := 0; i < len(data); i++ {
		data[i] = int(i)
	}

	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, data)
	}
}
