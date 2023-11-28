// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package float16

import (
	"testing"
)

func TestFloat16(t *testing.T) {
	for _, tc := range []struct {
		num  Num
		str  string
		want float32
	}{
		{
			num:  New(0),
			str:  "0",
			want: 0,
		},
		{
			num:  New(-1),
			str:  "-1",
			want: -1,
		},
		{
			num:  Float16Frombits(0xdead),
			str:  "-427.25",
			want: -427.25,
		},
	} {
		t.Run("", func(t *testing.T) {
			got := tc.num.Float32()
			if got != tc.want {
				t.Fatalf("invalid float16 value:\ngot= %v\nwant=%v", got, tc.want)
			}

			f16 := Float16Frombits(tc.num.Uint16())
			if got, want := f16, tc.num; got != want {
				t.Fatalf("roundtrip failed:\ngot= %v\nwant=%v", got, want)
			}

			str := tc.num.String()
			if got, want := str, tc.str; got != want {
				t.Fatalf("invalid float16 string representation:\ngot= %q\nwant=%q", got, want)
			}
		})
	}
}
