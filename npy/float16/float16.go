// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package float16

import (
	"math"
	"strconv"
)

// Num represents a half-precision floating point value (float16)
// stored on 16 bits.
//
// See https://en.wikipedia.org/wiki/Half-precision_floating-point_format for more informations.
type Num struct {
	bits uint16
}

// New creates a new half-precision floating point value from the provided
// float32 value.
func New(f float32) Num {
	var (
		bits = math.Float32bits(f)
		sign = uint16((bits >> 31) & 0x1)
		exp  = (bits >> 23) & 0xff
		res  = int16(exp) - 127 + 15
		fc   = uint16(bits>>13) & 0x3ff
	)
	switch {
	case exp == 0:
		res = 0
	case exp == 0xff:
		res = 0x1f
	case res > 0x1e:
		res = 0x1f
		fc = 0
	case res < 0x01:
		res = 0
		fc = 0
	}
	return Num{bits: (sign << 15) | uint16(res<<10) | fc}
}

// Float16frombits returns a new half-precision floating point value from the provided bits.
func Float16Frombits(bits uint16) Num {
	return Num{bits: bits}
}

func (f Num) Float32() float32 {
	var (
		sign = uint32((f.bits >> 15) & 0x1)
		exp  = (f.bits >> 10) & 0x1f
		res  = uint32(exp) + 127 - 15
		fc   = uint32(f.bits & 0x3ff)
	)
	switch {
	case exp == 0:
		res = 0
	case exp == 0x1f:
		res = 0xff
	}
	return math.Float32frombits((sign << 31) | (res << 23) | (fc << 13))
}

func (f Num) Uint16() uint16 { return f.bits }
func (f Num) String() string { return strconv.FormatFloat(float64(f.Float32()), 'g', -1, 32) }
