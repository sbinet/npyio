// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package npy

import (
	"reflect"
	"strings"
	"testing"

	"github.com/nlpodyssey/gopickle/pickle"
	py "github.com/nlpodyssey/gopickle/types"
)

func TestUnpickleDtype(t *testing.T) {
	for _, tc := range dtypeTests {
		t.Run(tc.name, func(t *testing.T) {
			pkl := pickle.NewUnpickler(strings.NewReader(tc.pkl))
			pkl.FindClass = ClassLoader

			got, err := pkl.Load()
			if err != nil {
				t.Fatalf("could not unpickle: %+v", err)
			}

			if got, want := got, tc.want; !reflect.DeepEqual(got, want) {
				t.Fatalf("invalid unpickled data for %q:\ngot= %+v\nwant=%+v", tc.code, got, want)
			}
		})
	}
}

func TestUnpickleNdarray(t *testing.T) {
	for _, tc := range ndarrayTests {
		t.Run(tc.name, func(t *testing.T) {
			pkl := pickle.NewUnpickler(strings.NewReader(tc.pkl))
			pkl.FindClass = ClassLoader

			got, err := pkl.Load()
			if err != nil {
				t.Fatalf("could not unpickle: %+v", err)
			}
			if got, want := got, tc.want; !reflect.DeepEqual(got, want) {
				t.Fatalf("invalid unpickled data for %q:\ngot= %+v\nwant=%+v", tc.code, got, want)
			}
		})
	}
}

func pylist(sli ...any) *py.List {
	return py.NewListFromSlice(sli)
}
