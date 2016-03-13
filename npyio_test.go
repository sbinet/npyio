// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"bytes"
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

func BenchmarkWriteIntArray(b *testing.B) {
	var data [1000]int
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, &data)
	}
}

func BenchmarkWriteFloat64Array(b *testing.B) {
	var data [1000]float64
	w := ioutil.Discard
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = Write(w, &data)
	}
}

type creader struct {
	buf []byte
	pos int
}

func (r *creader) Read(data []byte) (int, error) {
	n := copy(data, r.buf[r.pos:r.pos+len(data)])
	r.pos += n
	return n, nil
}

func (r *creader) reset() {
	r.pos = 0
}

func BenchmarkReadDense(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, mat64.NewDense(100, 10, make([]float64, 1000)))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var m mat64.Dense
		_ = Read(r, &m)
		r.reset()
	}
}

func BenchmarkReadFloat32Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]float32, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []float32
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadFloat64Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]float64, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []float64
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadBoolSlice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]bool, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []bool
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadUintSlice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]uint, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []uint
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadUint8Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]uint8, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []uint8
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadUint16Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]uint16, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []uint16
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadUint32Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]uint32, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []uint32
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadUint64Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]uint64, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []uint64
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadIntSlice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]int, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []int
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadInt8Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]int8, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []int8
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadInt16Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]int16, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []int16
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadInt32Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]int32, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []int32
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadInt64Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]int64, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []int64
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadComplex64Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]complex64, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []complex64
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadComplex128Slice(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]complex128, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data []complex128
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadIntArray(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]int, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data [1000]int
		_ = Read(r, &data)
		r.reset()
	}
}

func BenchmarkReadFloat64Array(b *testing.B) {
	buf := new(bytes.Buffer)
	_ = Write(buf, make([]float64, 1000))
	r := &creader{buf: buf.Bytes()}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var data [1000]float64
		_ = Read(r, &data)
		r.reset()
	}
}