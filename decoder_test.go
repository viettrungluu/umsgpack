// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests decoder.go.

package umsgpack_test

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"strconv"
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

// fillerChars generates n filler characters in the pattern 012345678901234....
func fillerChars(n int) []byte {
	rv := make([]byte, n)
	for i := 0; i < n; i += 1 {
		rv[i] = byte('0' + i%10)
	}
	return rv
}

// fillerBytes generates n filler bytes in the pattern 0, 1, 2, ..., 255, 0, 1, ....
func fillerBytes(n int) []byte {
	rv := make([]byte, n)
	for i := 0; i < n; i += 1 {
		rv[i] = byte(i % 256)
	}
	return rv
}

// genArrayData generates test array data with n entries.
func genArrayData(n int) []byte {
	rv := []byte{}
	for i := 0; i < n; i += 1 {
		s := strconv.Itoa(i)
		rv = append(rv, byte(0xa0+len(s)))
		rv = append(rv, []byte(s)...)
	}
	return rv
}

// genArrayData generates a test array with n entries.
func genArray(n int) []any {
	rv := []any{}
	for i := 0; i < n; i += 1 {
		rv = append(rv, strconv.Itoa(i))
	}
	return rv
}

// genMapData generates test map data with n key-value pairs.
func genMapData(n int) []byte {
	rv := []byte{}
	for i := 0; i < n; i += 1 {
		s := strconv.Itoa(i)
		rv = append(rv, byte(0xa0+len(s)))
		rv = append(rv, []byte(s)...)
		j := i % 10000
		rv = append(rv, 0xd1, byte(j>>8), byte(j))
	}
	return rv
}

// genMap generates test map with n key-value pairs.
func genMap(n int) map[any]any {
	rv := map[any]any{}
	for i := 0; i < n; i += 1 {
		rv[strconv.Itoa(i)] = i % 10000
	}
	return rv
}

type testCase struct {
	encoded []byte
	decoded any
	err     error
}

func runTestCases(t *testing.T, opts *UnmarshalOptions, tCs []testCase) {
	for _, tC := range tCs {
		buf := bytes.NewBuffer(tC.encoded)
		if actualDecoded, actualErr := Unmarshal(opts, buf); actualErr != tC.err {
			t.Errorf("unexected error for encoded=%q (decoded=%#v, err=%v): actualErr=%v", tC.encoded, tC.decoded, tC.err, actualErr)
		} else if tC.err == nil && !reflect.DeepEqual(actualDecoded, tC.decoded) {
			t.Errorf("unexected result for encoded=%q (decoded=%#v): actualDecoded=%#v", tC.encoded, tC.decoded, actualDecoded)
		}
	}
}

func TestUnmarshal_defaultOpts(t *testing.T) {
	opts := &UnmarshalOptions{}
	tCs := []testCase{
		// never used (0xc1):
		{encoded: []byte{0xc1}, err: InvalidFormatError},
		// nil:
		{encoded: []byte{0xc0}, decoded: nil},
		// boolean:
		{encoded: []byte{0xc2}, decoded: false},
		{encoded: []byte{0xc3}, decoded: true},
		// int:
		// - positive fixint:
		{encoded: []byte{0x00}, decoded: int(0)},
		{encoded: []byte{0x01}, decoded: int(1)},
		{encoded: []byte{0x02}, decoded: int(2)},
		{encoded: []byte{0x7f}, decoded: int(127)},
		// - negative fixint:
		{encoded: []byte{0xff}, decoded: int(-1)},
		{encoded: []byte{0xfe}, decoded: int(-2)},
		{encoded: []byte{0xe1}, decoded: int(-31)},
		{encoded: []byte{0xe0}, decoded: int(-32)},
		// - int 8:
		{encoded: []byte{0xd0, 0x00}, decoded: int(0)},
		{encoded: []byte{0xd0, 0x01}, decoded: int(1)},
		{encoded: []byte{0xd0, 0x7f}, decoded: int(127)},
		{encoded: []byte{0xd0, 0xff}, decoded: int(-1)},
		{encoded: []byte{0xd0, 0xfe}, decoded: int(-2)},
		{encoded: []byte{0xd0, 0x80}, decoded: int(-128)},
		{encoded: []byte{0xd0}, err: io.EOF},
		// - int 16:
		{encoded: []byte{0xd1, 0x00, 0x00}, decoded: int(0)},
		{encoded: []byte{0xd1, 0x00, 0x01}, decoded: int(1)},
		{encoded: []byte{0xd1, 0x7f, 0xff}, decoded: int(32767)},
		{encoded: []byte{0xd1, 0xff, 0xff}, decoded: int(-1)},
		{encoded: []byte{0xd1, 0xff, 0xfe}, decoded: int(-2)},
		{encoded: []byte{0xd1, 0x80, 0x00}, decoded: int(-32768)},
		{encoded: []byte{0xd1}, err: io.EOF},
		{encoded: []byte{0xd1, 0x00}, err: io.ErrUnexpectedEOF},
		// - int 32:
		{encoded: []byte{0xd2, 0x00, 0x00, 0x00, 0x00}, decoded: int(0)},
		{encoded: []byte{0xd2, 0x00, 0x00, 0x00, 0x01}, decoded: int(1)},
		{encoded: []byte{0xd2, 0x7f, 0xff, 0xff, 0xff}, decoded: int(1<<31 - 1)},
		{encoded: []byte{0xd2, 0xff, 0xff, 0xff, 0xff}, decoded: int(-1)},
		{encoded: []byte{0xd2, 0xff, 0xff, 0xff, 0xfe}, decoded: int(-2)},
		{encoded: []byte{0xd2, 0x80, 0x00, 0x00, 0x00}, decoded: int(-(1 << 31))},
		{encoded: []byte{0xd2}, err: io.EOF},
		{encoded: []byte{0xd2, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		// - int 64:
		{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: int(0)},
		{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: int(1)},
		{encoded: []byte{0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: int(1<<63 - 1)},
		{encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: int(-1)},
		{encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}, decoded: int(-2)},
		{encoded: []byte{0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: int(-(1 << 63))},
		{encoded: []byte{0xd3}, err: io.EOF},
		{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		// uint:
		// - uint 8:
		{encoded: []byte{0xcc, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xcc, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xcc, 0xfe}, decoded: uint(254)},
		{encoded: []byte{0xcc, 0xff}, decoded: uint(255)},
		{encoded: []byte{0xcc}, err: io.EOF},
		// - uint 16:
		{encoded: []byte{0xcd, 0x00, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xcd, 0x00, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xcd, 0xff, 0xfe}, decoded: uint(65534)},
		{encoded: []byte{0xcd, 0xff, 0xff}, decoded: uint(65535)},
		{encoded: []byte{0xcd}, err: io.EOF},
		{encoded: []byte{0xcd, 0x00}, err: io.ErrUnexpectedEOF},
		// - uint 32:
		{encoded: []byte{0xce, 0x00, 0x00, 0x00, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xce, 0x00, 0x00, 0x00, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xfe}, decoded: uint(1<<32 - 2)},
		{encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xff}, decoded: uint(1<<32 - 1)},
		{encoded: []byte{0xce}, err: io.EOF},
		{encoded: []byte{0xce, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		// - uint 64:
		{encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}, decoded: uint(1<<64 - 2)},
		{encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: uint(1<<64 - 1)},
		{encoded: []byte{0xcf}, err: io.EOF},
		{encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		// float32:
		{encoded: []byte{0xca, 0x00, 0x00, 0x00, 0x00}, decoded: float32(0)},
		{encoded: []byte{0xca, 0x00, 0x00, 0x00, 0x01}, decoded: float32(math.SmallestNonzeroFloat32)},
		{encoded: []byte{0xca, 0x3f, 0x40, 0x00, 0x00}, decoded: float32(0.75)},
		{encoded: []byte{0xca, 0x3f, 0x80, 0x00, 0x00}, decoded: float32(1)},
		{encoded: []byte{0xca, 0x3f, 0x80, 0x00, 0x01}, decoded: math.Nextafter32(1, 2)},
		{encoded: []byte{0xca, 0x7f, 0x7f, 0xff, 0xff}, decoded: float32(math.MaxFloat32)},
		{encoded: []byte{0xca, 0x7f, 0x80, 0x00, 0x00}, decoded: float32(math.Inf(1))},
		{encoded: []byte{0xca, 0xbf, 0x80, 0x00, 0x00}, decoded: float32(-1)},
		{encoded: []byte{0xca, 0xff, 0x80, 0x00, 0x00}, decoded: float32(math.Inf(-1))},
		{encoded: []byte{0xca}, err: io.EOF},
		{encoded: []byte{0xca, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		// float64:
		{encoded: []byte{0xcb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: float64(0)},
		{encoded: []byte{0xcb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: float64(math.SmallestNonzeroFloat64)},
		{encoded: []byte{0xcb, 0x3f, 0xe8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: float64(0.75)},
		{encoded: []byte{0xcb, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: float64(1)},
		{encoded: []byte{0xcb, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: math.Nextafter(1, 2)},
		{encoded: []byte{0xcb, 0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: float64(math.MaxFloat64)},
		{encoded: []byte{0xcb, 0x7f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: math.Inf(1)},
		{encoded: []byte{0xcb, 0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: float64(-1)},
		{encoded: []byte{0xcb, 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: math.Inf(-1)},
		{encoded: []byte{0xcb}, err: io.EOF},
		{encoded: []byte{0xcb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		// string:
		// - fixstr:
		{encoded: []byte{0xa0}, decoded: ""},
		{encoded: []byte{0xa1, 0x30}, decoded: "0"},
		{encoded: []byte{0xa2, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xbf}, fillerChars(31)...), decoded: "0123456789012345678901234567890"},
		{encoded: []byte{0xa1}, err: io.EOF},
		{encoded: append([]byte{0xbf}, fillerChars(30)...), err: io.ErrUnexpectedEOF},
		// - str 8:
		{encoded: []byte{0xd9, 0x00}, decoded: ""},
		{encoded: []byte{0xd9, 0x01, 0x30}, decoded: "0"},
		{encoded: []byte{0xd9, 0x02, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xd9, 0xff}, fillerChars(255)...), decoded: string(fillerChars(255))},
		{encoded: []byte{0xd9}, err: io.EOF},
		{encoded: []byte{0xd9, 0x01}, err: io.EOF},
		{encoded: []byte{0xd9, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		// - str 16:
		{encoded: []byte{0xda, 0x00, 0x00}, decoded: ""},
		{encoded: []byte{0xda, 0x00, 0x01, 0x30}, decoded: "0"},
		{encoded: []byte{0xda, 0x00, 0x02, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xda, 0xff, 0xff}, fillerChars(65535)...), decoded: string(fillerChars(65535))},
		{encoded: []byte{0xda}, err: io.EOF},
		{encoded: []byte{0xda, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xda, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xda, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		// - str 32:
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x00}, decoded: ""},
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x01, 0x30}, decoded: "0"},
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x02, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xdb, 0x00, 0x01, 0x86, 0xa0}, fillerChars(100000)...), decoded: string(fillerChars(100000))},
		{encoded: []byte{0xdb}, err: io.EOF},
		{encoded: []byte{0xdb, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		// binary:
		// - bin 8:
		{encoded: []byte{0xc4, 0x00}, decoded: []byte{}},
		{encoded: []byte{0xc4, 0x01, 0x00}, decoded: []byte{0}},
		{encoded: []byte{0xc4, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
		{encoded: append([]byte{0xc4, 0xff}, fillerBytes(255)...), decoded: fillerBytes(255)},
		{encoded: []byte{0xc4}, err: io.EOF},
		{encoded: []byte{0xc4, 0x01}, err: io.EOF},
		{encoded: []byte{0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		// - bin 16:
		{encoded: []byte{0xc5, 0x00, 0x00}, decoded: []byte{}},
		{encoded: []byte{0xc5, 0x00, 0x01, 0x00}, decoded: []byte{0}},
		{encoded: []byte{0xc5, 0x00, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
		{encoded: append([]byte{0xc5, 0xff, 0xff}, fillerBytes(65535)...), decoded: fillerBytes(65535)},
		{encoded: []byte{0xc5}, err: io.EOF},
		{encoded: []byte{0xc5, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xc5, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xc5, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		// - bin 32:
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x00}, decoded: []byte{}},
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x01, 0x00}, decoded: []byte{0}},
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
		{encoded: append([]byte{0xc6, 0x00, 0x01, 0x86, 0xa0}, fillerBytes(100000)...), decoded: fillerBytes(100000)},
		{encoded: []byte{0xc6}, err: io.EOF},
		{encoded: []byte{0xc6, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		// array:
		// - fixarray:
		{encoded: []byte{0x90}, decoded: []any{}},
		{encoded: append([]byte{0x91}, genArrayData(1)...), decoded: []any{"0"}},
		{encoded: append([]byte{0x92}, genArrayData(2)...), decoded: []any{"0", "1"}},
		{encoded: append([]byte{0x9f}, genArrayData(15)...), decoded: genArray(15)},
		{encoded: []byte{0x91}, err: io.EOF},
		{encoded: []byte{0x91, 0xa1}, err: io.EOF},
		{encoded: []byte{0x91, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0x9f}, genArrayData(14)...), err: io.EOF},
		// - array 16:
		{encoded: []byte{0xdc, 0x00, 0x00}, decoded: []any{}},
		{encoded: append([]byte{0xdc, 0x00, 0x01}, genArrayData(1)...), decoded: []any{"0"}},
		{encoded: append([]byte{0xdc, 0x00, 0x02}, genArrayData(2)...), decoded: []any{"0", "1"}},
		{encoded: append([]byte{0xdc, 0xff, 0xff}, genArrayData(65535)...), decoded: genArray(65535)},
		{encoded: []byte{0xdc}, err: io.EOF},
		{encoded: []byte{0xdc, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdc, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xdc, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0xdc, 0xff, 0xff}, genArrayData(65534)...), err: io.EOF},
		// - array 32:
		{encoded: []byte{0xdd, 0x00, 0x00, 0x00, 0x00}, decoded: []any{}},
		{encoded: append([]byte{0xdd, 0x00, 0x00, 0x00, 0x01}, genArrayData(1)...), decoded: []any{"0"}},
		{encoded: append([]byte{0xdd, 0x00, 0x00, 0x00, 0x02}, genArrayData(2)...), decoded: []any{"0", "1"}},
		{encoded: append([]byte{0xdd, 0x00, 0x01, 0x86, 0xa0}, genArrayData(100000)...), decoded: genArray(100000)},
		{encoded: []byte{0xdd}, err: io.EOF},
		{encoded: []byte{0xdd, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdd, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdd, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdd, 0x00, 0x00, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xdd, 0x00, 0x00, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0xdd, 0x00, 0x01, 0x86, 0xa0}, genArrayData(99999)...), err: io.EOF},
		// map:
		// - fixmap:
		{encoded: []byte{0x80}, decoded: map[any]any{}},
		{encoded: append([]byte{0x81}, genMapData(1)...), decoded: map[any]any{"0": int(0)}},
		{encoded: append([]byte{0x82}, genMapData(2)...), decoded: map[any]any{"0": int(0), "1": int(1)}},
		{encoded: append([]byte{0x8f}, genMapData(15)...), decoded: genMap(15)},
		{encoded: []byte{0x81}, err: io.EOF},
		{encoded: []byte{0x81, 0xa1}, err: io.EOF},
		{encoded: []byte{0x81, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0x81, 0xa1, 0x30}, err: io.EOF},
		{encoded: []byte{0x81, 0xa1, 0x30, 0xa1}, err: io.EOF},
		{encoded: []byte{0x81, 0xa1, 0x30, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0x8f}, genMapData(14)...), err: io.EOF},
		// - map 16:
		{encoded: []byte{0xde, 0x00, 0x00}, decoded: map[any]any{}},
		{encoded: append([]byte{0xde, 0x00, 0x01}, genMapData(1)...), decoded: map[any]any{"0": int(0)}},
		{encoded: append([]byte{0xde, 0x00, 0x02}, genMapData(2)...), decoded: map[any]any{"0": int(0), "1": int(1)}},
		{encoded: append([]byte{0xde, 0xff, 0xff}, genMapData(65535)...), decoded: genMap(65535)},
		{encoded: []byte{0xde}, err: io.EOF},
		{encoded: []byte{0xde, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xde, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xde, 0x00, 0x01, 0xa1}, err: io.EOF},
		{encoded: []byte{0xde, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xde, 0x00, 0x01, 0xa1, 0x30}, err: io.EOF},
		{encoded: []byte{0xde, 0x00, 0x01, 0xa1, 0x30, 0xa1}, err: io.EOF},
		{encoded: []byte{0xde, 0x00, 0x01, 0xa1, 0x30, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0xde, 0xff, 0xff}, genMapData(65534)...), err: io.EOF},
		// - map 32:
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x00}, decoded: map[any]any{}},
		{encoded: append([]byte{0xdf, 0x00, 0x00, 0x00, 0x01}, genMapData(1)...), decoded: map[any]any{"0": int(0)}},
		{encoded: append([]byte{0xdf, 0x00, 0x00, 0x00, 0x02}, genMapData(2)...), decoded: map[any]any{"0": int(0), "1": int(1)}},
		{encoded: append([]byte{0xdf, 0x00, 0x01, 0x86, 0xa0}, genMapData(100000)...), decoded: genMap(100000)},
		{encoded: []byte{0xdf}, err: io.EOF},
		{encoded: []byte{0xdf, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdf, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1}, err: io.EOF},
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1, 0x30}, err: io.EOF},
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1, 0x30, 0xa1}, err: io.EOF},
		{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1, 0x30, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0xdf, 0x00, 0x01, 0x86, 0xa0}, genMapData(99999)...), err: io.EOF},
		// extension type (unsupported):
		// - ext 8:
		{encoded: []byte{0xc7, 0x00, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 7, Data: []byte{}}},
		{encoded: []byte{0xc7, 0x01, 0x00, 0x42}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0x42}}},
		{encoded: []byte{0xc7, 0x02, 0x80, 0x42, 0x43}, decoded: &UnresolvedExtensionType{ExtensionType: -128, Data: []byte{0x42, 0x43}}},
		{encoded: append([]byte{0xc7, 0xff, 0x7f}, fillerBytes(255)...), decoded: &UnresolvedExtensionType{ExtensionType: 127, Data: fillerBytes(255)}},
		{encoded: []byte{0xc7}, err: io.EOF},
		{encoded: []byte{0xc7, 0x01}, err: io.EOF},
		{encoded: []byte{0xc7, 0x01, 0x00}, err: io.EOF},
		{encoded: []byte{0xc7, 0x02, 0x00, 0x42}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0xc7, 0xff, 0x00}, fillerBytes(254)...), err: io.ErrUnexpectedEOF},
		// - ext 16:
		{encoded: []byte{0xc8, 0x00, 0x00, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 7, Data: []byte{}}},
		{encoded: []byte{0xc8, 0x00, 0x01, 0x00, 0x42}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0x42}}},
		{encoded: []byte{0xc8, 0x00, 0x02, 0x80, 0x42, 0x43}, decoded: &UnresolvedExtensionType{ExtensionType: -128, Data: []byte{0x42, 0x43}}},
		{encoded: append([]byte{0xc8, 0xff, 0xff, 0x7f}, fillerBytes(65535)...), decoded: &UnresolvedExtensionType{ExtensionType: 127, Data: fillerBytes(65535)}},
		{encoded: []byte{0xc8}, err: io.EOF},
		{encoded: []byte{0xc8, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xc8, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xc8, 0x00, 0x01, 0x00}, err: io.EOF},
		{encoded: []byte{0xc8, 0x00, 0x02, 0x00, 0x42}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0xc8, 0xff, 0xff, 0x00}, fillerBytes(65534)...), err: io.ErrUnexpectedEOF},
		// - ext 32:
		{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x00, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 7, Data: []byte{}}},
		{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x01, 0x00, 0x42}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0x42}}},
		{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x02, 0x80, 0x42, 0x43}, decoded: &UnresolvedExtensionType{ExtensionType: -128, Data: []byte{0x42, 0x43}}},
		{encoded: append([]byte{0xc9, 0x00, 0x01, 0x86, 0xa0, 0x7f}, fillerBytes(100000)...), decoded: &UnresolvedExtensionType{ExtensionType: 127, Data: fillerBytes(100000)}},
		{encoded: []byte{0xc9}, err: io.EOF},
		{encoded: []byte{0xc9, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xc9, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xc9, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x01}, err: io.EOF},
		{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x01, 0x00}, err: io.EOF},
		{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x02, 0x00, 0x42}, err: io.ErrUnexpectedEOF},
		{encoded: append([]byte{0xc9, 0x00, 0x01, 0x86, 0xa0, 0x00}, fillerBytes(99999)...), err: io.ErrUnexpectedEOF},
		// - fixext 1
		{encoded: []byte{0xd4, 0x00, 0x00}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0}}},
		{encoded: []byte{0xd4}, err: io.EOF},
		{encoded: []byte{0xd4, 0x00}, err: io.EOF},
		// - fixext 2
		{encoded: []byte{0xd5, 0x00, 0x00, 0x01}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1}}},
		{encoded: []byte{0xd5}, err: io.EOF},
		{encoded: []byte{0xd5, 0x00}, err: io.EOF},
		{encoded: []byte{0xd5, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		// - fixext 4
		{encoded: []byte{0xd6, 0x00, 0x00, 0x01, 0x02, 0x03}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1, 2, 3}}},
		{encoded: []byte{0xd6}, err: io.EOF},
		{encoded: []byte{0xd6, 0x00}, err: io.EOF},
		{encoded: []byte{0xd6, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xd6, 0x00, 0x00, 0x01, 0x02}, err: io.ErrUnexpectedEOF},
		// - fixext 8
		{encoded: []byte{0xd7, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1, 2, 3, 4, 5, 6, 7}}},
		{encoded: []byte{0xd7}, err: io.EOF},
		{encoded: []byte{0xd7, 0x00}, err: io.EOF},
		{encoded: []byte{0xd7, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xd7, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, err: io.ErrUnexpectedEOF},
		// - fixext 16
		{encoded: []byte{0xd8, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}}},
		{encoded: []byte{0xd8}, err: io.EOF},
		{encoded: []byte{0xd8, 0x00}, err: io.EOF},
		{encoded: []byte{0xd8, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
		{encoded: []byte{0xd8, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e}, err: io.ErrUnexpectedEOF},
		// valid map key types (via fixmap):
		{encoded: []byte{0x81, 0xc0, 0x2a}, decoded: map[any]any{nil: int(42)}},
		{encoded: []byte{0x81, 0xc2, 0x2a}, decoded: map[any]any{false: int(42)}},
		{encoded: []byte{0x81, 0xc3, 0x2a}, decoded: map[any]any{true: int(42)}},
		{encoded: []byte{0x81, 0x0c, 0x2a}, decoded: map[any]any{int(12): int(42)}},
		{encoded: []byte{0x81, 0xf4, 0x2a}, decoded: map[any]any{int(-12): int(42)}},
		{encoded: []byte{0x81, 0xd0, 0x0c, 0x2a}, decoded: map[any]any{int(12): int(42)}},
		{encoded: []byte{0x81, 0xd1, 0xff, 0xf4, 0x2a}, decoded: map[any]any{int(-12): int(42)}},
		{encoded: []byte{0x81, 0xd2, 0x00, 0x00, 0x00, 0x0c, 0x2a}, decoded: map[any]any{int(12): int(42)}},
		{encoded: []byte{0x81, 0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xf4, 0x2a}, decoded: map[any]any{int(-12): int(42)}},
		{encoded: []byte{0x81, 0xcc, 0x0c, 0x2a}, decoded: map[any]any{uint(12): int(42)}},
		{encoded: []byte{0x81, 0xcd, 0x00, 0x0c, 0x2a}, decoded: map[any]any{uint(12): int(42)}},
		{encoded: []byte{0x81, 0xce, 0x00, 0x00, 0x00, 0x0c, 0x2a}, decoded: map[any]any{uint(12): int(42)}},
		{encoded: []byte{0x81, 0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x2a}, decoded: map[any]any{uint(12): int(42)}},
		{encoded: []byte{0x81, 0xca, 0x3f, 0x80, 0x00, 0x00, 0x2a}, decoded: map[any]any{float32(1): int(42)}},
		{encoded: []byte{0x81, 0xcb, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a}, decoded: map[any]any{float64(1): int(42)}},
		{encoded: []byte{0x81, 0xa2, 0x31, 0x32, 0x2a}, decoded: map[any]any{"12": int(42)}},
		{encoded: []byte{0x81, 0xd9, 0x02, 0x31, 0x32, 0x2a}, decoded: map[any]any{"12": int(42)}},
		{encoded: []byte{0x81, 0xda, 0x00, 0x02, 0x31, 0x32, 0x2a}, decoded: map[any]any{"12": int(42)}},
		{encoded: []byte{0x81, 0xdb, 0x00, 0x00, 0x00, 0x02, 0x31, 0x32, 0x2a}, decoded: map[any]any{"12": int(42)}},
		// invalid map key types (via fixmap):
		// TODO: test invalid map key types.
		// TODO: test map DuplicateKeyError.
		// TODO: test timestamp ext.
	}
	runTestCases(t, opts, tCs)
}

// TODO: test extensions, other opts.
