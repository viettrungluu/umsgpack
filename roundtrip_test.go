// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests Marshal[ToBytes] and Unmarshal[Bytes], together, checking that (certain) objects
// are properly round-tripped. (Note that some objects wouldn't be fully round-tripped, losing type
// information.)

package umsgpack_test

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/viettrungluu/umsgpack"
)

var roundTrippableObjects = []any{
	// nil:
	nil,
	// boolean:
	false,
	true,
	// int:
	// - positive fixint:
	int(0),
	int(1),
	int(42),
	int(0x7f),
	// - negative fixint:
	int(-1),
	int(-12),
	int(-32),
	// - int 8:
	int(-33),
	int(-0x42),
	int(-0x80),
	// - int 16:
	int(0x80),
	int(0x1234),
	int(0x7fff),
	int(-0x81),
	int(-0x1234),
	int(-0x8000),
	// - int 32:
	int(0x8000),
	int(0x123456),
	int(0x7fffffff),
	int(-0x8001),
	int(-0x123456),
	int(-0x80000000),
	// - int 64:
	int(0x800000000),
	int(0x123456789abcd),
	int(0x7fffffffffffffff),
	int(-0x800000001),
	int(-0x123456789abcd),
	int(-0x8000000000000000),
	// uint:
	// - uint 8:
	uint(0),
	uint(1),
	uint(42),
	uint(0xff),
	// - uint 16:
	uint(0x100),
	uint(0x1234),
	uint(0xffff),
	// - uint 32:
	uint(0x10000),
	uint(0x123456),
	uint(0xffffffff),
	// - uint 64:
	uint(0x100000000),
	uint(0x123456789abcd),
	uint(0xffffffffffffffff),
	// float32:
	float32(0),
	float32(1),
	float32(123.45),
	float32(math.MaxFloat32),
	float32(math.Inf(1)),
	float32(-1),
	float32(-123.45),
	float32(math.Inf(-1)),
	// float64:
	float64(0),
	float64(1),
	float64(123.45),
	float64(math.MaxFloat64),
	float64(math.Inf(1)),
	float64(-1),
	float64(-123.45),
	float64(math.Inf(-1)),
	// string:
	// - fixstr:
	"",
	"a",
	"bc",
	strings.Repeat("x", 31),
	// - str 8:
	strings.Repeat("x", 32),
	strings.Repeat("x", 0x42),
	strings.Repeat("x", 0xff),
	// - str 16:
	strings.Repeat("x", 0x100),
	strings.Repeat("x", 0x1234),
	strings.Repeat("x", 0xffff),
	// - str 32:
	strings.Repeat("x", 0x10000),
	strings.Repeat("x", 0x123456),
	// binary:
	// - bin 8:
	[]byte{},
	[]byte("a"),
	[]byte("bc"),
	[]byte(strings.Repeat("x", 0x42)),
	[]byte(strings.Repeat("x", 0xff)),
	// - bin 16:
	[]byte(strings.Repeat("x", 0x100)),
	[]byte(strings.Repeat("x", 0x1234)),
	[]byte(strings.Repeat("x", 0xffff)),
	// - bin 32:
	[]byte(strings.Repeat("x", 0x10000)),
	[]byte(strings.Repeat("x", 0x12345)),
	// array:
	// - fixarray:
	[]any{},
	[]any{123},
	[]any{123, "hi"},
	append([]any{123, "hi"}, genArray(0xf-2)...),
	// - array 16:
	append([]any{123, "hi"}, genArray(0x10-2)...),
	append([]any{123, "hi"}, genArray(0x1234-2)...),
	append([]any{123, "hi"}, genArray(0xffff-2)...),
	// - array 32:
	append([]any{123, "hi"}, genArray(0x10000-2)...),
	append([]any{123, "hi"}, genArray(0x12345-2)...),
	// map:
	// - fixmap:
	map[any]any{},
	map[any]any{"foo": 123},
	map[any]any{"foo": 123, 45: "bar"},
	genMap(0xf),
	// - map 16:
	genMap(0x10),
	genMap(0x42),
	genMap(0xffff),
	// - map 32
	genMap(0x10000),
	genMap(0x12345),
	// Timestamp extension type (-1):
	// - timestamp 32
	time.Unix(0, 0),
	time.Unix(0x12345678, 0),
	// - timestamp 64
	time.Unix(0x23456789a, 123456789),
	// - timestamp 96
	time.Unix(0x123456789abcdef0, 0x12345678),
}

func testRoundtripObj(t *testing.T, name string, obj any) {
	encoded, err := MarshalToBytes(nil, obj)
	if err != nil {
		t.Errorf("%v: %#v: MarshalToBytes returned error: %v", name, obj, err)
		return
	}

	decoded, err := UnmarshalBytes(nil, encoded)
	if err != nil {
		t.Errorf("%v: %#v: UnmarshalBytes returned error: %v", name, obj, err)
		return
	}

	if !reflect.DeepEqual(decoded, obj) {
		t.Errorf("%v: %#v: roundtrip mismatch: %#v", name, obj, decoded)
		return
	}
}

func TestRoundtrip(t *testing.T) {
	for i, obj := range roundTrippableObjects {
		testRoundtripObj(t, strconv.Itoa(i), obj)
	}

	// The whole array is roundtrippable!
	testRoundtripObj(t, "everything", roundTrippableObjects)
}
