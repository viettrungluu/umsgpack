// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains (testable) examples.

package umsgpack_test

import (
	"bytes"
	"fmt"

	"github.com/viettrungluu/umsgpack"
)

func ExampleMarshal() {
	input := []any{
		map[any]any{
			"foo": "bar",
		},
		123,
		4.5,
	}
	buf := &bytes.Buffer{}
	if err := umsgpack.Marshal(nil, buf, input); err != nil {
		panic(err)
	} else {
		fmt.Println(buf.Bytes())
	}
	// Output: [147 129 163 102 111 111 163 98 97 114 123 203 64 18 0 0 0 0 0 0]
}

func ExampleMarshalToBytes() {
	input := map[any]any{
		"hello":   "world",
		"one_two": 3,
		45:        []any{"six", 7, 8.9},
	}
	if output, err := umsgpack.MarshalToBytes(nil, input); err != nil {
		panic(err)
	} else {
		// NOTE: output isn't deterministic since map iteration order isn't deterministic.
		// But its length should be deterministic.
		fmt.Println(len(output))
	}
	// Output: 38
}

func ExampleUnmarshal() {
	input := []byte{
		147, 129, 163, 102, 111, 111, 163, 98, 97, 114, 123, 203, 64, 18, 0, 0, 0, 0, 0, 0,
	}
	buf := bytes.NewBuffer(input)
	if output, err := umsgpack.Unmarshal(nil, buf); err != nil {
		panic(err)
	} else {
		fmt.Println(output)
	}
	// Output: [map[foo:bar] 123 4.5]
}

func ExampleUnmarshalBytes() {
	input := []byte{
		147, 129, 163, 102, 111, 111, 163, 98, 97, 114, 123, 203, 64, 18, 0, 0, 0, 0, 0, 0,
	}
	if output, err := umsgpack.UnmarshalBytes(nil, input); err != nil {
		panic(err)
	} else {
		fmt.Println(output)
	}
	// Output: [map[foo:bar] 123 4.5]
}
