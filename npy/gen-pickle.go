// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type Code struct {
	Py string
	Go string
}

func main() {
	src := new(bytes.Buffer)

	fmt.Fprintf(src, `// Copyright 2023 The npyio Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Automatically generated. DO NOT EDIT.

package npy

import (
	"encoding/binary"

	"github.com/sbinet/npyio/npy/float16"
)

`)

	dtype(src, []string{
		`np.dtype(">b1")`,
		`np.dtype("<b1")`,
		`np.dtype(">i1")`,
		`np.dtype("<i1")`,
		`np.dtype(">i2")`,
		`np.dtype("<i2")`,
		`np.dtype(">i4")`,
		`np.dtype("<i4")`,
		`np.dtype(">i8")`,
		`np.dtype("<i8")`,
		`np.dtype("int8")`,
		`np.dtype("int16")`,
		`np.dtype("int32")`,
		`np.dtype("int64")`,
		`np.dtype(">u1")`,
		`np.dtype("<u1")`,
		`np.dtype(">u2")`,
		`np.dtype("<u2")`,
		`np.dtype(">u4")`,
		`np.dtype("<u4")`,
		`np.dtype(">u8")`,
		`np.dtype("<u8")`,
		`np.dtype("uint8")`,
		`np.dtype("uint16")`,
		`np.dtype("uint32")`,
		`np.dtype("uint64")`,
		`np.dtype("float16")`,
		`np.dtype("float32")`,
		`np.dtype("float64")`,
		`np.dtype(">f2")`,
		`np.dtype("<f2")`,
		`np.dtype(">f4")`,
		`np.dtype("<f4")`,
		`np.dtype(">f8")`,
		`np.dtype("<f8")`,
		`np.dtype(">c8")`,
		`np.dtype("<c8")`,
		`np.dtype(">c16")`,
		`np.dtype("<c16")`,
		`np.dtype("<S4")`,
		`np.dtype(">S4")`,
		`np.dtype("=S4")`,
		`np.dtype("|S4")`,
		`np.dtype("S4")`,
		`np.dtype("S8")`,
		`np.dtype("S42")`,
		`np.dtype("<O4")`,
		`np.dtype(">O4")`,
		`np.dtype("=O4")`,
		`np.dtype("|O4")`,
		`np.dtype("O4")`,
		`np.dtype("<O8")`,
		`np.dtype(">O8")`,
		`np.dtype("=O8")`,
		`np.dtype("|O8")`,
		`np.dtype("O8")`,
		`np.dtype([('f1', [('f1', np.int16)])])`,
		`np.dtype("i4, (2,3)f8")`,
		`np.dtype("i2, i4, (2,3)f8")`,
		`np.dtype("i2, i8, (2,3)f8")`,
		// UTF
		`np.dtype("<U10")`,
		`np.dtype(">U10")`,
		// time
		`np.dtype("timedelta64")`,
		`np.dtype("<m8")`,
		`np.dtype(">m8")`,
		`np.dtype("datetime64")`,
		`np.dtype("<M8")`,
		`np.dtype(">M8")`,
	})
	ndarray(src, []Code{
		// bool
		{`np.array([True,False,True], dtype="<b1")`, `[]bool{true, false, true}`},
		{`np.array([True,False,True], dtype=">b1")`, `[]bool{true, false, true}`},
		// signed integers
		{`np.array([-1,+2,-3], dtype="<i1")`, `[]int8{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype=">i1")`, `[]int8{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype="<i2")`, `[]int16{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype=">i2")`, `[]int16{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype="<i4")`, `[]int32{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype=">i4")`, `[]int32{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype="<i8")`, `[]int64{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype=">i8")`, `[]int64{-1,+2,-3}`},
		// unsigned integers
		{`np.array([1,2,3], dtype="<u1")`, `[]uint8{1,2,3}`},
		{`np.array([1,2,3], dtype=">u1")`, `[]uint8{1,2,3}`},
		{`np.array([1,2,3], dtype="<u2")`, `[]uint16{1,2,3}`},
		{`np.array([1,2,3], dtype=">u2")`, `[]uint16{1,2,3}`},
		{`np.array([1,2,3], dtype="<u4")`, `[]uint32{1,2,3}`},
		{`np.array([1,2,3], dtype=">u4")`, `[]uint32{1,2,3}`},
		{`np.array([1,2,3], dtype="<u8")`, `[]uint64{1,2,3}`},
		{`np.array([1,2,3], dtype=">u8")`, `[]uint64{1,2,3}`},
		// floats
		{`np.array([-1,+2,-3], dtype="<f2")`, `[]float16.Num{float16.New(-1),float16.New(+2),float16.New(-3)}`},
		{`np.array([-1,+2,-3], dtype=">f2")`, `[]float16.Num{float16.New(-1),float16.New(+2),float16.New(-3)}`},
		{`np.array([-1,+2,-3], dtype="<f4")`, `[]float32{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype=">f4")`, `[]float32{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype="<f8")`, `[]float64{-1,+2,-3}`},
		{`np.array([-1,+2,-3], dtype=">f8")`, `[]float64{-1,+2,-3}`},
		// complexes
		{`np.array([(-1+1j),(+2-2j),(-3+3j)], dtype="<c8")`, `[]complex64{complex(-1,1),complex(2,-2),complex(-3,3)}`},
		{`np.array([(-1+1j),(+2-2j),(-3+3j)], dtype=">c8")`, `[]complex64{complex(-1,1),complex(2,-2),complex(-3,3)}`},
		{`np.array([(-1+1j),(+2-2j),(-3+3j)], dtype="<c16")`, `[]complex128{complex(-1,1),complex(2,-2),complex(-3,3)}`},
		{`np.array([(-1+1j),(+2-2j),(-3+3j)], dtype=">c16")`, `[]complex128{complex(-1,1),complex(2,-2),complex(-3,3)}`},
		// strings
		{`np.array("hello world!", dtype="S12")`, `"hello world!"`},
		{`np.array(["hell","o wo", "rld!"], dtype="S4")`, `[]string{"hell", "o wo", "rld!"}`},
		// utf
		{`np.array("hello, 世界!", dtype="<U10")`, `"hello, 世界!"`},
		{`np.array("hello, 世界!", dtype=">U10")`, `"hello, 世界!"`},
		{`np.array(["hello, 世界!"], dtype="<U10")`, `[]string{"hello, 世界!"}`},
		{`np.array(["hello, 世界!"], dtype=">U10")`, `[]string{"hello, 世界!"}`},
		{`np.array([["hello"], [", 世界!"]], dtype="<U5")`, `[]string{"hello", ", 世界!"}`},
		{`np.array([["hello"], [", 世界!"]], dtype=">U5")`, `[]string{"hello", ", 世界!"}`},
		// n-dim
		{`np.array([[-1,-2,-3],[-4,-5,-6]], dtype="<i8")`, `[]int64{-1,-2,-3,-4,-5,-6}`},
		{`np.array([[-1,-2,-3],[-4,-5,-6]], dtype=">i8")`, `[]int64{-1,-2,-3,-4,-5,-6}`},
		{`np.array([[(-1+1j),(+2-2j),(-3+3j)],[(-4+4j),(-5+5j),(-6+6j)]], dtype="<c16")`,
			`[]complex128{complex(-1,1),complex(2,-2),complex(-3,3),complex(-4,4),complex(-5,5),complex(-6,6)}`},
		{`np.array([[(-1+1j),(+2-2j),(-3+3j)],[(-4+4j),(-5+5j),(-6+6j)]], dtype=">c16")`,
			`[]complex128{complex(-1,1),complex(2,-2),complex(-3,3),complex(-4,4),complex(-5,5),complex(-6,6)}`},
		// ragged-arrays
		{`np.array([[-1],[-2,-3],[-4,-5,-6]], dtype="object")`, `pylist(pylist(-1),pylist(-2,-3),pylist(-4,-5,-6))`},
		{`np.array([[-1],["-2",-3],[-4,-5,"-6"]], dtype="object")`, `pylist(pylist(-1),pylist("-2",-3),pylist(-4,-5,"-6"))`},
	})

	dst, err := format.Source(src.Bytes())
	if err != nil {
		log.Printf("===\n%s\n===\n", src)
		log.Fatalf("could not format go code: %+v", err)
	}

	err = os.WriteFile("zall_test.go", dst, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func dtype(w io.Writer, tests []string) {
	fmt.Fprintf(w, `var dtypeTests = []struct{
	name string
	code string
	pkl  string
	want *ArrayDescr
}{
`)

	gen := func(w io.Writer, i int, code string) error {
		f1, err := os.CreateTemp("", "dtype-*.pkl")
		if err != nil {
			return err
		}
		_ = f1.Close()
		defer os.Remove(f1.Name())

		f2, err := os.CreateTemp("", "dtype-*.json")
		if err != nil {
			return err
		}
		_ = f2.Close()
		defer os.Remove(f2.Name())

		script, err := os.CreateTemp("", "dtype-*.py")
		if err != nil {
			return err
		}
		defer script.Close()
		defer os.Remove(script.Name())

		fmt.Fprintf(script, pyDtype, f1.Name(), f2.Name(), code)

		err = script.Close()
		if err != nil {
			return err
		}

		cmd := exec.Command("python", script.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("could not run %q: %w", script.Name(), err)
		}

		pkl, err := os.ReadFile(f1.Name())
		if err != nil {
			return err
		}

		ref, err := os.ReadFile(f2.Name())
		if err != nil {
			return err
		}

		var data ArrayDescr

		err = json.Unmarshal(ref, &data)
		if err != nil {
			return fmt.Errorf("could not unmarshal json-ref: %w", err)
		}

		dt, err := marshalArrayDescr(data)
		if err != nil {
			return fmt.Errorf("could not marshal dtype from %q: %w", f2.Name(), err)
		}

		fmt.Fprintf(w, `	{
		// pickle.dumps(%[1]s, protocol=4)
		name: "dtype-%[2]d",
		code: `+"`%[1]s`"+`,
		pkl: %[3]q,
		want: &%[4]s,
	},
`, code, i, pkl, dt,
		)

		return nil
	}

	for i, code := range tests {
		err := gen(w, i, code)
		if err != nil {
			panic(fmt.Errorf("could not generate data for %q: %+v", code, err))
		}
	}
	fmt.Fprintf(w, "}\n\n")
}

func ndarray(w io.Writer, tests []Code) {
	fmt.Fprintf(w, `var ndarrayTests = []struct{
	name string
	code string
	pkl  string
	want *Array
}{
`)

	gen := func(w io.Writer, i int, code Code) error {
		f1, err := os.CreateTemp("", "ndarray-*.pkl")
		if err != nil {
			return err
		}
		_ = f1.Close()
		defer os.Remove(f1.Name())

		f2, err := os.CreateTemp("", "ndarray-*.json")
		if err != nil {
			return err
		}
		_ = f2.Close()
		defer os.Remove(f2.Name())

		script, err := os.CreateTemp("", "ndarray-*.py")
		if err != nil {
			return err
		}
		defer script.Close()
		defer os.Remove(script.Name())

		fmt.Fprintf(script, pyNdarray, f1.Name(), f2.Name(), code.Py)

		err = script.Close()
		if err != nil {
			return err
		}

		cmd := exec.Command("python", script.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("could not run %q: %w", script.Name(), err)
		}

		pkl, err := os.ReadFile(f1.Name())
		if err != nil {
			return err
		}

		ref, err := os.ReadFile(f2.Name())
		if err != nil {
			return err
		}

		var data Array

		err = json.Unmarshal(ref, &data)
		if err != nil {
			return fmt.Errorf("could not unmarshal json-ref: %w", err)
		}

		arr, err := marshalArray(data, code.Go)
		if err != nil {
			return fmt.Errorf("could not marshal ndarray from %q: %w", f2.Name(), err)
		}

		fmt.Fprintf(w, `	{
		// pickle.dumps(%[1]s, protocol=4)
		name: "ndarray-%[2]d",
		code: `+"`%[1]s`"+`,
		pkl: %[3]q,
		want: &%[4]s,
	},
`, code.Py, i, pkl, arr,
		)
		return nil
	}

	for i, code := range tests {
		err := gen(w, i, code)
		if err != nil {
			panic(fmt.Errorf("could not generate data for %q: %+v", code, err))
		}
	}
	fmt.Fprintf(w, "}\n\n")
}

type Array struct {
	Descr   ArrayDescr `json:"dtype"`
	Shape   []int      `json:"shape"`
	Strides []int      `json:"strides"`
}

type ArrayDescr struct {
	Descr  []Descr          `json:"descr"`
	Kind   string           `json:"kind"`
	Size   int              `json:"esize"`
	Align  int              `json:"align"`
	Order  string           `json:"order"`
	Fields map[string]Field `json:"fields"`
	Names  []string         `json:"names"`
	Flags  int              `json:"flags"`
	Subarr *Subarray        `json:"subarr"`
}

type Descr struct {
	Name  string      `json:"name"`
	Descr [][2]string `json:"descr"`
	Shape []int       `json:"shape,omitempty"`
}

type Field struct {
	Descr  ArrayDescr `json:"dtype"`
	Offset int        `json:"offset"`
}

type Subarray struct {
	Descr ArrayDescr `json:"dtype"`
	Shape []int      `json:"shape"`
}

func marshalArray(arr Array, godata string) ([]byte, error) {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "Array{\n")
	dt, err := marshalArrayDescr(arr.Descr)
	if err != nil {
		return nil, fmt.Errorf("could not marshal dtype of array: %w", err)
	}
	fmt.Fprintf(w, "\tdescr: %s,\n", dt)
	switch len(arr.Shape) {
	case 0:
		// no-op.
	default:
		fmt.Fprintf(w, "\tshape: []int{")
		for i, v := range arr.Shape {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%d", v)
		}
		fmt.Fprintf(w, "},\n")
	}
	switch len(arr.Strides) {
	case 0:
		// no-op.
	default:
		fmt.Fprintf(w, "\tstrides: []int{")
		for i, v := range arr.Strides {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%d", v)
		}
		fmt.Fprintf(w, "},\n")
	}
	fmt.Fprintf(w, "\tdata: %s,\n", godata)
	fmt.Fprintf(w, "}")
	return w.Bytes(), nil
}

func marshalArrayDescr(dt ArrayDescr) ([]byte, error) {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "ArrayDescr{\n")
	fmt.Fprintf(w, "\tkind: '%s',\n", dt.Kind)
	order := "nativeEndian"
	switch dt.Order {
	case "<":
		order = "binary.LittleEndian"
	case ">":
		order = "binary.BigEndian"
	case "=":
		order = "nativeEndian"
	case "|":
		order = "nil"
		// FIXME(sbinet): handle as not applicable ?
	default:
		return nil, fmt.Errorf("unknown endianness %q", dt.Order)
	}
	if len(dt.Descr) == 1 && (dt.Kind != "V") {
		descr := dt.Descr[0].Descr[0][1]
		switch {
		case strings.HasPrefix(descr, "<"):
			order = "binary.LittleEndian"
		case strings.HasPrefix(descr, ">"):
			order = "binary.BigEndian"
		case strings.HasPrefix(descr, "="):
			order = "nativeEndian"
		case strings.HasPrefix(descr, "|"):
			order = "nil"
		}
	}
	fmt.Fprintf(w, "\torder: %s,\n", order)
	fmt.Fprintf(w, "\tesize: %d,\n", dt.Size)
	fmt.Fprintf(w, "\talign: %d,\n", dt.Align)
	if dt.Flags != 0 {
		fmt.Fprintf(w, "\tflags: %d,\n", dt.Flags)
	}
	if len(dt.Names) > 0 {
		fmt.Fprintf(w, "\tnames: []string{")
		for i, name := range dt.Names {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%q", name)
		}
		fmt.Fprintf(w, "},\n")
	}

	if len(dt.Fields) > 0 {
		fmt.Fprintf(w, "\tfields: map[string]structField{\n")
		keys := make([]string, 0, len(dt.Fields))
		for k := range dt.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			f := dt.Fields[k]
			sub, err := marshalArrayDescr(f.Descr)
			if err != nil {
				return nil, fmt.Errorf("could not marshal field %q: %w", k, err)
			}
			fmt.Fprintf(w, "\t\t%q: {\ndtype: %s,\noffset: %d,\n},\n", k, sub, f.Offset)
		}
		fmt.Fprintf(w, "},\n")
	}

	if dt.Subarr != nil {
		fmt.Fprintf(w, "\tsubarr: &subarrayDescr{\n")
		sub, err := marshalArrayDescr(dt.Subarr.Descr)
		if err != nil {
			return nil, fmt.Errorf("could not marshal subarray dtype %q: %w", dt.Subarr.Descr.Kind, err)
		}
		fmt.Fprintf(w, "\t\tdtype: %s,\n", sub)
		fmt.Fprintf(w, "\t\tshape: []int{")
		for i, dim := range dt.Subarr.Shape {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%d", dim)
		}
		fmt.Fprintf(w, "\t\t},\n")
		fmt.Fprintf(w, "\t},\n")
	}

	fmt.Fprintf(w, "}")
	return w.Bytes(), nil
}

const pyDtype = `#!/usr/bin/env python
import json
import pickle
import numpy as np

pkl = open("%[1]s", "bw")
dt = np.dtype(%[3]s)
pickle.dump(dt, pkl, protocol=4)
pkl.close()

def todescr(ds):
	o = []
	for _,v in enumerate(ds):
		d = {
			"name": v[0],
			"descr": v[1],
		}
		if type(v[1]) == type(""):
			d["descr"] = [["", v[1]]]
		if len(v) > 2:
			d["shape"] = v[2]
		o.append(d)
		pass
	return o

def tofields(fs):
	o = {}
	if not fs:
		return o
	for k in fs:
		v = fs[k]
		o[k] = {"dtype":todtype(v[0]), "offset":v[1]}
	return o

def todtype(dt):
	## print(">>> dtype: %%s..." %%(dt,))
	orig = dt
	shape = None
	if type(dt) == type(tuple()):
		shape = dt[1]
		dt = dt[0]
		pass
	o = {
		"descr":  todescr(dt.descr),
		"kind":   dt.kind,
		"esize":  dt.itemsize,
		"align":  dt.alignment,
		"order":  str(dt.byteorder),
		"fields": tofields(dt.fields),
		"names":  dt.names or [],
		"flags":  dt.flags,
	}
	if shape != None:
		o["shape"] = [i for i in shape]
	sub = dt.subdtype
	if sub != None:
		## print(" >>> sub: %%s ==> %%s" %%(sub,sub[0].byteorder,))
		o["subarr"] = {
			"dtype": todtype(sub[0]),
			"shape": sub[1],
		}
		pass
	## print(">>> dtype: %%s ==> %%s" %%(orig,o,))
	return o

txt = open("%[2]s", "w")
json.dump(todtype(dt), txt, indent=" ")
txt.close()
`

const pyNdarray = `#!/usr/bin/env python
import json
import pickle
import numpy as np

pkl = open("%[1]s", "bw")
arr = %[3]s
pickle.dump(arr, pkl, protocol=4)
pkl.close()

def todescr(ds):
	o = []
	for _,v in enumerate(ds):
		d = {
			"name": v[0],
			"descr": v[1],
		}
		if type(v[1]) == type(""):
			d["descr"] = [["", v[1]]]
		if len(v) > 2:
			d["shape"] = v[2]
		o.append(d)
		pass
	return o

def tofields(fs):
	o = {}
	if not fs:
		return o
	for k in fs:
		v = fs[k]
		o[k] = {"dtype":todtype(v[0]), "offset":v[1]}
	return o

def todtype(dt):
	## print(">>> dtype: %%s..." %%(dt,))
	orig = dt
	shape = None
	if type(dt) == type(tuple()):
		shape = dt[1]
		dt = dt[0]
		pass
	o = {
		"descr":  todescr(dt.descr),
		"kind":   dt.kind,
		"esize":  dt.itemsize,
		"align":  dt.alignment,
		"order":  str(dt.byteorder),
		"fields": tofields(dt.fields),
		"names":  dt.names or [],
		"flags":  dt.flags,
	}
	if shape != None:
		o["shape"] = [i for i in shape]
	sub = dt.subdtype
	if sub != None:
		## print(" >>> sub: %%s ==> %%s" %%(sub,sub[0].byteorder,))
		o["subarr"] = {
			"dtype": todtype(sub[0]),
			"shape": sub[1],
		}
		pass
	## print(">>> dtype: %%s ==> %%s" %%(orig,o,))
	return o

def toarray(arr):
	return {
		"dtype": todtype(arr.dtype),
		"shape": list(arr.shape),
		"strides": list(arr.strides),
	}

txt = open("%[2]s", "w")
json.dump(toarray(arr), txt, indent=" ")
txt.close()
`
