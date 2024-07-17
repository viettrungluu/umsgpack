// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains (testable) examples.

package umsgpack_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/viettrungluu/umsgpack"
)

// Marshal:

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

func ExampleMarshal_applicationExtension() {
	// Marshals a time.Duration to a extension type 42, containing a 64-bit value, big-endian.
	marshalDuration := func(obj any) (any, error) {
		if duration, ok := obj.(time.Duration); !ok {
			return obj, nil
		} else {
			data := make([]byte, 8)
			binary.BigEndian.PutUint64(data, uint64(duration))
			return &umsgpack.UnresolvedExtensionType{ExtensionType: 42, Data: data}, nil
		}
	}
	opts := &umsgpack.MarshalOptions{
		ApplicationMarshalTransformer: marshalDuration,
	}

	input := time.Duration(123)
	buf := &bytes.Buffer{}
	if err := umsgpack.Marshal(opts, buf, input); err != nil {
		panic(err)
	} else {
		fmt.Println(buf.Bytes())
	}
	// Output: [215 42 0 0 0 0 0 0 0 123]
}

func ExampleDefaultStructMarshalTransformer() {
	opts := &umsgpack.MarshalOptions{
		ApplicationMarshalTransformer: umsgpack.DefaultStructMarshalTransformer,
	}

	input := struct {
		Foo string
		Bar int
		baz int
	}{"hello", 123, 0}
	if output, err := umsgpack.MarshalToBytes(opts, input); err != nil {
		panic(err)
	} else {
		// NOTE: output isn't deterministic since map iteration order isn't deterministic.
		// But its length should be deterministic.
		fmt.Println(len(output))
	}
	// Output: 16
}

// Unmarshal:

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

func ExampleUnmarshal_applicationExtension() {
	// Unmarshals a time.Duration represented as a 64-bit value, big-endian.
	unmarshalDuration := func(data []byte) (any, bool, error) {
		if len(data) != 8 {
			return nil, false, errors.New("invalid")
		} else {
			return time.Duration(binary.BigEndian.Uint64(data)), true, nil
		}
	}
	opts := &umsgpack.UnmarshalOptions{
		ApplicationUnmarshalTransformer: umsgpack.MakeExtensionTypeUnmarshalTransformer(
			map[int8]umsgpack.UnmarshalExtensionTypeFn{
				42: unmarshalDuration, // Extension type 42.
			},
		),
	}

	input := []byte{
		215, 42, 0, 0, 0, 0, 0, 0, 0, 123,
	}
	buf := bytes.NewBuffer(input)
	if output, err := umsgpack.Unmarshal(opts, buf); err != nil {
		panic(err)
	} else {
		fmt.Println(output)
	}
	// Output: 123ns
}
