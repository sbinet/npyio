// Copyright 2016 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npyio

import (
	"io/ioutil"
	"testing"
)

func BenchmarkWriteFloatSlice(b *testing.B) {
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
