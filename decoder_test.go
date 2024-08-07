// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests decoder.go.

package umsgpack_test

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"testing"
	"time"

	. "github.com/viettrungluu/umsgpack"
)

// An unmarshalTestCase defines a test case for unmarshalling: the encoded bytes and the expected
// decoded result or the expected error.
type unmarshalTestCase struct {
	encoded []byte
	decoded any
	err     error
}

// testUnmarshal is a helper for testing Unmarshal/UnmarshalBytes with the given options for the
// given test cases.
func testUnmarshal(t *testing.T, opts *UnmarshalOptions, tCs []unmarshalTestCase) {
	for _, tC := range tCs {
		buf := bytes.NewBuffer(tC.encoded)
		if actualDecoded, actualErr := Unmarshal(opts, buf); actualErr != tC.err {
			t.Errorf("unexected error for encoded=%q (decoded=%#v, err=%v): actualErr=%v", tC.encoded, tC.decoded, tC.err, actualErr)
		} else if tC.err == nil && !reflect.DeepEqual(actualDecoded, tC.decoded) {
			t.Errorf("unexected result for encoded=%q (decoded=%#v): actualDecoded=%#v", tC.encoded, tC.decoded, actualDecoded)
		}

		if actualDecoded, actualErr := UnmarshalBytes(opts, tC.encoded); actualErr != tC.err {
			t.Errorf("unexected error for encoded=%q (decoded=%#v, err=%v): actualErr=%v", tC.encoded, tC.decoded, tC.err, actualErr)
		} else if tC.err == nil && !reflect.DeepEqual(actualDecoded, tC.decoded) {
			t.Errorf("unexected result for encoded=%q (decoded=%#v): actualDecoded=%#v", tC.encoded, tC.decoded, actualDecoded)
		}
	}
}

// commonUnmarshalTestCases contains unmarshalTestCases that should pass regardless of options.
var commonUnmarshalTestCases = []unmarshalTestCase{
	// no data:
	{encoded: []byte{}, err: io.EOF},
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
	{encoded: []byte{0xd0}, err: io.ErrUnexpectedEOF},
	// - int 16:
	{encoded: []byte{0xd1, 0x00, 0x00}, decoded: int(0)},
	{encoded: []byte{0xd1, 0x00, 0x01}, decoded: int(1)},
	{encoded: []byte{0xd1, 0x7f, 0xff}, decoded: int(32767)},
	{encoded: []byte{0xd1, 0xff, 0xff}, decoded: int(-1)},
	{encoded: []byte{0xd1, 0xff, 0xfe}, decoded: int(-2)},
	{encoded: []byte{0xd1, 0x80, 0x00}, decoded: int(-32768)},
	{encoded: []byte{0xd1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd1, 0x00}, err: io.ErrUnexpectedEOF},
	// - int 32:
	{encoded: []byte{0xd2, 0x00, 0x00, 0x00, 0x00}, decoded: int(0)},
	{encoded: []byte{0xd2, 0x00, 0x00, 0x00, 0x01}, decoded: int(1)},
	{encoded: []byte{0xd2, 0x7f, 0xff, 0xff, 0xff}, decoded: int(1<<31 - 1)},
	{encoded: []byte{0xd2, 0xff, 0xff, 0xff, 0xff}, decoded: int(-1)},
	{encoded: []byte{0xd2, 0xff, 0xff, 0xff, 0xfe}, decoded: int(-2)},
	{encoded: []byte{0xd2, 0x80, 0x00, 0x00, 0x00}, decoded: int(-(1 << 31))},
	{encoded: []byte{0xd2}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd2, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	// - int 64:
	{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: int(0)},
	{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: int(1)},
	{encoded: []byte{0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: int(1<<63 - 1)},
	{encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: int(-1)},
	{encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}, decoded: int(-2)},
	{encoded: []byte{0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: int(-(1 << 63))},
	{encoded: []byte{0xd3}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	// uint:
	// - uint 8:
	{encoded: []byte{0xcc, 0x00}, decoded: uint(0)},
	{encoded: []byte{0xcc, 0x01}, decoded: uint(1)},
	{encoded: []byte{0xcc, 0xfe}, decoded: uint(254)},
	{encoded: []byte{0xcc, 0xff}, decoded: uint(255)},
	{encoded: []byte{0xcc}, err: io.ErrUnexpectedEOF},
	// - uint 16:
	{encoded: []byte{0xcd, 0x00, 0x00}, decoded: uint(0)},
	{encoded: []byte{0xcd, 0x00, 0x01}, decoded: uint(1)},
	{encoded: []byte{0xcd, 0xff, 0xfe}, decoded: uint(65534)},
	{encoded: []byte{0xcd, 0xff, 0xff}, decoded: uint(65535)},
	{encoded: []byte{0xcd}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xcd, 0x00}, err: io.ErrUnexpectedEOF},
	// - uint 32:
	{encoded: []byte{0xce, 0x00, 0x00, 0x00, 0x00}, decoded: uint(0)},
	{encoded: []byte{0xce, 0x00, 0x00, 0x00, 0x01}, decoded: uint(1)},
	{encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xfe}, decoded: uint(1<<32 - 2)},
	{encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xff}, decoded: uint(1<<32 - 1)},
	{encoded: []byte{0xce}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xce, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	// - uint 64:
	{encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: uint(0)},
	{encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: uint(1)},
	{encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}, decoded: uint(1<<64 - 2)},
	{encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: uint(1<<64 - 1)},
	{encoded: []byte{0xcf}, err: io.ErrUnexpectedEOF},
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
	{encoded: []byte{0xca}, err: io.ErrUnexpectedEOF},
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
	{encoded: []byte{0xcb}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xcb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	// string:
	// - fixstr:
	{encoded: []byte{0xa0}, decoded: ""},
	{encoded: []byte{0xa1, 0x30}, decoded: "0"},
	{encoded: []byte{0xa2, 0x30, 0x31}, decoded: "01"},
	{encoded: append([]byte{0xbf}, fillerChars(31)...), decoded: "0123456789012345678901234567890"},
	{encoded: []byte{0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xbf}, fillerChars(30)...), err: io.ErrUnexpectedEOF},
	// - str 8:
	{encoded: []byte{0xd9, 0x00}, decoded: ""},
	{encoded: []byte{0xd9, 0x01, 0x30}, decoded: "0"},
	{encoded: []byte{0xd9, 0x02, 0x30, 0x31}, decoded: "01"},
	{encoded: append([]byte{0xd9, 0xff}, fillerChars(255)...), decoded: string(fillerChars(255))},
	{encoded: []byte{0xd9}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd9, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd9, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	// - str 16:
	{encoded: []byte{0xda, 0x00, 0x00}, decoded: ""},
	{encoded: []byte{0xda, 0x00, 0x01, 0x30}, decoded: "0"},
	{encoded: []byte{0xda, 0x00, 0x02, 0x30, 0x31}, decoded: "01"},
	{encoded: append([]byte{0xda, 0xff, 0xff}, fillerChars(65535)...), decoded: string(fillerChars(65535))},
	{encoded: []byte{0xda}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xda, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xda, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xda, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	// - str 32:
	{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x00}, decoded: ""},
	{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x01, 0x30}, decoded: "0"},
	{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x02, 0x30, 0x31}, decoded: "01"},
	{encoded: append([]byte{0xdb, 0x00, 0x01, 0x86, 0xa0}, fillerChars(100000)...), decoded: string(fillerChars(100000))},
	{encoded: []byte{0xdb}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdb, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	// binary:
	// - bin 8:
	{encoded: []byte{0xc4, 0x00}, decoded: []byte{}},
	{encoded: []byte{0xc4, 0x01, 0x00}, decoded: []byte{0}},
	{encoded: []byte{0xc4, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
	{encoded: append([]byte{0xc4, 0xff}, fillerBytes(255)...), decoded: fillerBytes(255)},
	{encoded: []byte{0xc4}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc4, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	// - bin 16:
	{encoded: []byte{0xc5, 0x00, 0x00}, decoded: []byte{}},
	{encoded: []byte{0xc5, 0x00, 0x01, 0x00}, decoded: []byte{0}},
	{encoded: []byte{0xc5, 0x00, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
	{encoded: append([]byte{0xc5, 0xff, 0xff}, fillerBytes(65535)...), decoded: fillerBytes(65535)},
	{encoded: []byte{0xc5}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc5, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc5, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc5, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	// - bin 32:
	{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x00}, decoded: []byte{}},
	{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x01, 0x00}, decoded: []byte{0}},
	{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
	{encoded: append([]byte{0xc6, 0x00, 0x01, 0x86, 0xa0}, fillerBytes(100000)...), decoded: fillerBytes(100000)},
	{encoded: []byte{0xc6}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc6, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	// array:
	// - fixarray:
	{encoded: []byte{0x90}, decoded: []any{}},
	{encoded: append([]byte{0x91}, genArrayData(1)...), decoded: []any{"0"}},
	{encoded: append([]byte{0x92}, genArrayData(2)...), decoded: []any{"0", "1"}},
	{encoded: append([]byte{0x9f}, genArrayData(15)...), decoded: genArray(15)},
	{encoded: []byte{0x91}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0x91, 0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0x91, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0x9f}, genArrayData(14)...), err: io.ErrUnexpectedEOF},
	// - array 16:
	{encoded: []byte{0xdc, 0x00, 0x00}, decoded: []any{}},
	{encoded: append([]byte{0xdc, 0x00, 0x01}, genArrayData(1)...), decoded: []any{"0"}},
	{encoded: append([]byte{0xdc, 0x00, 0x02}, genArrayData(2)...), decoded: []any{"0", "1"}},
	{encoded: append([]byte{0xdc, 0xff, 0xff}, genArrayData(65535)...), decoded: genArray(65535)},
	{encoded: []byte{0xdc}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdc, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdc, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdc, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xdc, 0xff, 0xff}, genArrayData(65534)...), err: io.ErrUnexpectedEOF},
	// - array 32:
	{encoded: []byte{0xdd, 0x00, 0x00, 0x00, 0x00}, decoded: []any{}},
	{encoded: append([]byte{0xdd, 0x00, 0x00, 0x00, 0x01}, genArrayData(1)...), decoded: []any{"0"}},
	{encoded: append([]byte{0xdd, 0x00, 0x00, 0x00, 0x02}, genArrayData(2)...), decoded: []any{"0", "1"}},
	{encoded: append([]byte{0xdd, 0x00, 0x01, 0x86, 0xa0}, genArrayData(100000)...), decoded: genArray(100000)},
	{encoded: []byte{0xdd}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdd, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdd, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdd, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdd, 0x00, 0x00, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdd, 0x00, 0x00, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xdd, 0x00, 0x01, 0x86, 0xa0}, genArrayData(99999)...), err: io.ErrUnexpectedEOF},
	// map:
	// - fixmap:
	{encoded: []byte{0x80}, decoded: map[any]any{}},
	{encoded: append([]byte{0x81}, genMapData(1)...), decoded: map[any]any{"0": int(0)}},
	{encoded: append([]byte{0x82}, genMapData(2)...), decoded: map[any]any{"0": int(0), "1": int(1)}},
	{encoded: append([]byte{0x8f}, genMapData(15)...), decoded: genMap(15)},
	{encoded: []byte{0x81}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0x81, 0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0x81, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0x81, 0xa1, 0x30}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0x81, 0xa1, 0x30, 0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0x81, 0xa1, 0x30, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0x8f}, genMapData(14)...), err: io.ErrUnexpectedEOF},
	// - map 16:
	{encoded: []byte{0xde, 0x00, 0x00}, decoded: map[any]any{}},
	{encoded: append([]byte{0xde, 0x00, 0x01}, genMapData(1)...), decoded: map[any]any{"0": int(0)}},
	{encoded: append([]byte{0xde, 0x00, 0x02}, genMapData(2)...), decoded: map[any]any{"0": int(0), "1": int(1)}},
	{encoded: append([]byte{0xde, 0xff, 0xff}, genMapData(65535)...), decoded: genMap(65535)},
	{encoded: []byte{0xde}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xde, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xde, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xde, 0x00, 0x01, 0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xde, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xde, 0x00, 0x01, 0xa1, 0x30}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xde, 0x00, 0x01, 0xa1, 0x30, 0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xde, 0x00, 0x01, 0xa1, 0x30, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xde, 0xff, 0xff}, genMapData(65534)...), err: io.ErrUnexpectedEOF},
	// - map 32:
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x00}, decoded: map[any]any{}},
	{encoded: append([]byte{0xdf, 0x00, 0x00, 0x00, 0x01}, genMapData(1)...), decoded: map[any]any{"0": int(0)}},
	{encoded: append([]byte{0xdf, 0x00, 0x00, 0x00, 0x02}, genMapData(2)...), decoded: map[any]any{"0": int(0), "1": int(1)}},
	{encoded: append([]byte{0xdf, 0x00, 0x01, 0x86, 0xa0}, genMapData(100000)...), decoded: genMap(100000)},
	{encoded: []byte{0xdf}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1, 0x30}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1, 0x30, 0xa1}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1, 0x30, 0xc4, 0x02, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xdf, 0x00, 0x01, 0x86, 0xa0}, genMapData(99999)...), err: io.ErrUnexpectedEOF},
	// extension type (base errors):
	// - ext 8:
	{encoded: []byte{0xc7}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc7, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc7, 0x01, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc7, 0x02, 0x00, 0x42}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xc7, 0xff, 0x00}, fillerBytes(254)...), err: io.ErrUnexpectedEOF},
	// - ext 16:
	{encoded: []byte{0xc8}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc8, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc8, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc8, 0x00, 0x01, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc8, 0x00, 0x02, 0x00, 0x42}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xc8, 0xff, 0xff, 0x00}, fillerBytes(65534)...), err: io.ErrUnexpectedEOF},
	// - ext 32:
	{encoded: []byte{0xc9}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc9, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc9, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x01}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x01, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x02, 0x00, 0x42}, err: io.ErrUnexpectedEOF},
	{encoded: append([]byte{0xc9, 0x00, 0x01, 0x86, 0xa0, 0x00}, fillerBytes(99999)...), err: io.ErrUnexpectedEOF},
	// - fixext 1
	{encoded: []byte{0xd4}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd4, 0x00}, err: io.ErrUnexpectedEOF},
	// - fixext 2
	{encoded: []byte{0xd5}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd5, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd5, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	// - fixext 4
	{encoded: []byte{0xd6}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd6, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd6, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd6, 0x00, 0x00, 0x01, 0x02}, err: io.ErrUnexpectedEOF},
	// - fixext 8
	{encoded: []byte{0xd7}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd7, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd7, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd7, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, err: io.ErrUnexpectedEOF},
	// - fixext 16
	{encoded: []byte{0xd8}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd8, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd8, 0x00, 0x00}, err: io.ErrUnexpectedEOF},
	{encoded: []byte{0xd8, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e}, err: io.ErrUnexpectedEOF},
	// supported map key types (via fixmap):
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
}

// timestampUnmarshalTestCases contains test cases for the built-in support for the timestamp
// extension (-1).
var timestampUnmarshalTestCases = []unmarshalTestCase{
	// Timestamp extension type (-1):
	// - timestamp 32
	//   (as fixext 4, which is canonical/minimal)
	{encoded: []byte{0xd6, 0xff, 0x00, 0x00, 0x00, 0x00}, decoded: time.Unix(0, 0)},
	{encoded: []byte{0xd6, 0xff, 0x12, 0x34, 0x56, 0x78}, decoded: time.Unix(0x12345678, 0)},
	//   (as ext 8/16/32)
	{encoded: []byte{0xc7, 0x04, 0xff, 0x12, 0x34, 0x56, 0x78}, decoded: time.Unix(0x12345678, 0)},
	{encoded: []byte{0xc8, 0x00, 0x04, 0xff, 0x12, 0x34, 0x56, 0x78}, decoded: time.Unix(0x12345678, 0)},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x04, 0xff, 0x12, 0x34, 0x56, 0x78}, decoded: time.Unix(0x12345678, 0)},
	// - timestamp 64
	//   (as fixext 8, which is canonical/minimal)
	{encoded: []byte{0xd7, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: time.Unix(0, 0)},
	{encoded: []byte{0xd7, 0xff, 0x1d, 0x6f, 0x34, 0x56, 0x34, 0x56, 0x78, 0x9a}, decoded: time.Unix(0x23456789a, 123456789)},
	{encoded: []byte{0xd7, 0xff, 0xee, 0x6b, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00}, err: InvalidTimestampError},
	//   (as ext 8/16/32)
	{encoded: []byte{0xc7, 0x08, 0xff, 0x1d, 0x6f, 0x34, 0x56, 0x34, 0x56, 0x78, 0x9a}, decoded: time.Unix(0x23456789a, 123456789)},
	{encoded: []byte{0xc8, 0x00, 0x08, 0xff, 0x1d, 0x6f, 0x34, 0x56, 0x34, 0x56, 0x78, 0x9a}, decoded: time.Unix(0x23456789a, 123456789)},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x08, 0xff, 0x1d, 0x6f, 0x34, 0x56, 0x34, 0x56, 0x78, 0x9a}, decoded: time.Unix(0x23456789a, 123456789)},
	// - timestamp 96
	//   (as ext 8, which is canonical/minimal)
	{encoded: []byte{0xc7, 0x0c, 0xff, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}, decoded: time.Unix(0x123456789abcdef0, 0x12345678)},
	{encoded: []byte{0xc7, 0x0c, 0xff, 0x3b, 0x9a, 0xca, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, err: InvalidTimestampError},
	//   (as ext 16/32)
	{encoded: []byte{0xc8, 0x00, 0x0c, 0xff, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}, decoded: time.Unix(0x123456789abcdef0, 0x12345678)},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x0c, 0xff, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}, decoded: time.Unix(0x123456789abcdef0, 0x12345678)},
	// - invalid lengths (via ext 8)
	{encoded: []byte{0xc7, 0x00, 0xff}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x01, 0xff, 0x00}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x02, 0xff, 0x00, 0x01}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x03, 0xff, 0x00, 0x01, 0x02}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x05, 0xff, 0x00, 0x01, 0x02, 0x03, 0x04}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x07, 0xff, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x09, 0xff, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x0b, 0xff, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}, err: InvalidTimestampError},
	{encoded: []byte{0xc7, 0x0d, 0xff, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c}, err: InvalidTimestampError},
	// supported map key types (via fixmap):
	{encoded: []byte{0x81, 0xd6, 0xff, 0x12, 0x34, 0x56, 0x78, 0x2a}, decoded: map[any]any{time.Unix(0x12345678, 0): int(42)}},
}

// defaultOptsUnmarshalTestCases contains unmarshalTestCases that should pass for the default
// options.
//
// NOTE: Avoid testing extensions 34 and 35 here.
var defaultOptsUnmarshalTestCases = []unmarshalTestCase{
	// extension type (unsupported):
	// - ext 8:
	{encoded: []byte{0xc7, 0x00, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 7, Data: []byte{}}},
	{encoded: []byte{0xc7, 0x01, 0x00, 0x42}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0x42}}},
	{encoded: []byte{0xc7, 0x02, 0x80, 0x42, 0x43}, decoded: &UnresolvedExtensionType{ExtensionType: -128, Data: []byte{0x42, 0x43}}},
	{encoded: append([]byte{0xc7, 0xff, 0x7f}, fillerBytes(255)...), decoded: &UnresolvedExtensionType{ExtensionType: 127, Data: fillerBytes(255)}},
	// - ext 16:
	{encoded: []byte{0xc8, 0x00, 0x00, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 7, Data: []byte{}}},
	{encoded: []byte{0xc8, 0x00, 0x01, 0x00, 0x42}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0x42}}},
	{encoded: []byte{0xc8, 0x00, 0x02, 0x80, 0x42, 0x43}, decoded: &UnresolvedExtensionType{ExtensionType: -128, Data: []byte{0x42, 0x43}}},
	{encoded: append([]byte{0xc8, 0xff, 0xff, 0x7f}, fillerBytes(65535)...), decoded: &UnresolvedExtensionType{ExtensionType: 127, Data: fillerBytes(65535)}},
	// - ext 32:
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x00, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 7, Data: []byte{}}},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x01, 0x00, 0x42}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0x42}}},
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x02, 0x80, 0x42, 0x43}, decoded: &UnresolvedExtensionType{ExtensionType: -128, Data: []byte{0x42, 0x43}}},
	{encoded: append([]byte{0xc9, 0x00, 0x01, 0x86, 0xa0, 0x7f}, fillerBytes(100000)...), decoded: &UnresolvedExtensionType{ExtensionType: 127, Data: fillerBytes(100000)}},
	// - fixext 1
	{encoded: []byte{0xd4, 0x00, 0x00}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0}}},
	// - fixext 2
	{encoded: []byte{0xd5, 0x00, 0x00, 0x01}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1}}},
	// - fixext 4
	{encoded: []byte{0xd6, 0x00, 0x00, 0x01, 0x02, 0x03}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1, 2, 3}}},
	// - fixext 8
	{encoded: []byte{0xd7, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1, 2, 3, 4, 5, 6, 7}}},
	// - fixext 16
	{encoded: []byte{0xd8, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}, decoded: &UnresolvedExtensionType{ExtensionType: 0, Data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}}},
	// unsupported map key types (via fixmap):
	{encoded: []byte{0x81, 0xc4, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xc5, 0x00, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xc6, 0x00, 0x00, 0x00, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0x90, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xdc, 0x00, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xdd, 0x00, 0x00, 0x00, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0x80, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xde, 0x00, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xdf, 0x00, 0x00, 0x00, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xc7, 0x00, 0x07, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xc8, 0x00, 0x00, 0x07, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xc9, 0x00, 0x00, 0x00, 0x00, 0x07, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xd4, 0x00, 0x00, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xd5, 0x00, 0x00, 0x01, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xd6, 0x00, 0x00, 0x01, 0x02, 0x03, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xd7, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x2a}, err: UnsupportedKeyTypeError},
	{encoded: []byte{0x81, 0xd8, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x2a}, err: UnsupportedKeyTypeError},
	// duplicate keys (via fixmap):
	{encoded: []byte{0x82, 0x0c, 0x2a, 0x0c, 0x2b}, err: DuplicateKeyError},
	// TODO: test more key types for duplicate keys?
}

// nonDefaultOptsUnmarshalTestCases contains unmarshalTestCases that should pass for when all
// boolean options are true.
var nonDefaultOptsUnmarshalTestCases = []unmarshalTestCase{
	// unsupported map key types (via fixmap):
	{encoded: []byte{0x81, 0xc4, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xc5, 0x00, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xc6, 0x00, 0x00, 0x00, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0x90, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xdc, 0x00, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xdd, 0x00, 0x00, 0x00, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0x80, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xde, 0x00, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xdf, 0x00, 0x00, 0x00, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xc7, 0x00, 0x07, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xc8, 0x00, 0x00, 0x07, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xc9, 0x00, 0x00, 0x00, 0x00, 0x07, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xd4, 0x00, 0x00, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xd5, 0x00, 0x00, 0x01, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xd6, 0x00, 0x00, 0x01, 0x02, 0x03, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xd7, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x2a}, decoded: map[any]any{}},
	{encoded: []byte{0x81, 0xd8, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x2a}, decoded: map[any]any{}},
	// duplicate keys (via fixmap):
	{encoded: []byte{0x82, 0x0c, 0x2a, 0x0c, 0x2b}, decoded: map[any]any{int(12): int(42)}},
	// TODO: test more key types for duplicate keys?
}

// TestUnmarshal_defaultOpts tests Unmarshal/UnmarshalBytes with the default options (all boolean
// options are false).
func TestUnmarshal_defaultOpts(t *testing.T) {
	opts := &UnmarshalOptions{}
	testUnmarshal(t, opts, commonUnmarshalTestCases)
	testUnmarshal(t, opts, timestampUnmarshalTestCases)
	testUnmarshal(t, opts, defaultOptsUnmarshalTestCases)
}

// TestUnmarshal_nonDefaultOpts tests Unmarshal/UnmarshalBytes with all boolean options set to true.
func TestUnmarshal_nonDefaultOpts(t *testing.T) {
	opts := &UnmarshalOptions{
		DisableDuplicateKeyError:       true,
		DisableUnsupportedKeyTypeError: true,
		// TODO: DisableStandardUnmarshalTransformer
	}
	testUnmarshal(t, opts, commonUnmarshalTestCases)
	testUnmarshal(t, opts, timestampUnmarshalTestCases)
	testUnmarshal(t, opts, nonDefaultOptsUnmarshalTestCases)
}

// Used by TestUnmarshal_applicationExtensions below.
type testExtensionType struct {
	data []byte
}

var applicationExtensionsUnmarshalTestCases = []unmarshalTestCase{
	// extension type 34 (supported):
	// - ext 8:
	{encoded: []byte{0xc7, 0x00, 0x22}, decoded: ""},
	{encoded: []byte{0xc7, 0x02, 0x22, 0x68, 0x69}, decoded: "hi"},
	// - ext 16:
	{encoded: []byte{0xc8, 0x00, 0x02, 0x22, 0x68, 0x69}, decoded: "hi"},
	// - ext 32:
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x02, 0x22, 0x68, 0x69}, decoded: "hi"},
	// - fixext 1
	{encoded: []byte{0xd4, 0x22, 0x68}, decoded: "h"},
	// - fixext 2
	{encoded: []byte{0xd5, 0x22, 0x68, 0x69}, decoded: "hi"},
	// - fixext 4
	{encoded: []byte{0xd6, 0x22, 0x68, 0x69, 0x68, 0x69}, decoded: "hihi"},
	// - fixext 8
	{encoded: []byte{0xd7, 0x22, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69}, decoded: "hihihihi"},
	// - fixext 16
	{encoded: []byte{0xd8, 0x22, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69}, decoded: "hihihihihihihihi"},
	// - as map key types (via fixmap)
	{encoded: []byte{0x81, 0xd5, 0x22, 0x68, 0x69, 0x2a}, decoded: map[any]any{"hi": int(42)}},
	// extension type 35 (supported):
	// - ext 8:
	{encoded: []byte{0xc7, 0x00, 0x23}, decoded: &testExtensionType{data: []byte{}}},
	{encoded: []byte{0xc7, 0x02, 0x23, 0x68, 0x69}, decoded: &testExtensionType{data: []byte("hi")}},
	// - ext 16:
	{encoded: []byte{0xc8, 0x00, 0x02, 0x23, 0x68, 0x69}, decoded: &testExtensionType{data: []byte("hi")}},
	// - ext 32:
	{encoded: []byte{0xc9, 0x00, 0x00, 0x00, 0x02, 0x23, 0x68, 0x69}, decoded: &testExtensionType{data: []byte("hi")}},
	// - fixext 1
	{encoded: []byte{0xd4, 0x23, 0x68}, decoded: &testExtensionType{data: []byte("h")}},
	// - fixext 2
	{encoded: []byte{0xd5, 0x23, 0x68, 0x69}, decoded: &testExtensionType{data: []byte("hi")}},
	// - fixext 4
	{encoded: []byte{0xd6, 0x23, 0x68, 0x69, 0x68, 0x69}, decoded: &testExtensionType{data: []byte("hihi")}},
	// - fixext 8
	{encoded: []byte{0xd7, 0x23, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69}, decoded: &testExtensionType{data: []byte("hihihihi")}},
	// - fixext 16
	{encoded: []byte{0xd8, 0x23, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69, 0x68, 0x69}, decoded: &testExtensionType{data: []byte("hihihihihihihihi")}},
	// - as map key types (via fixmap)
	{encoded: []byte{0x81, 0xd5, 0x23, 0x68, 0x69, 0x2a}, err: UnsupportedKeyTypeError},
}

func TestUnmarshal_applicationExtensions(t *testing.T) {
	opts := &UnmarshalOptions{
		ApplicationUnmarshalTransformer: MakeExtensionTypeUnmarshalTransformer(
			map[int8]UnmarshalExtensionTypeFn{
				// 34: just unmarshals to a string.
				34: func(data []byte) (any, bool, error) {
					return string(data), true, nil
				},
				// 35: unmarshals to a *testExtensionType.
				35: func(data []byte) (any, bool, error) {
					return &testExtensionType{data: data}, false, nil
				},
			},
		),
	}
	testUnmarshal(t, opts, commonUnmarshalTestCases)
	testUnmarshal(t, opts, timestampUnmarshalTestCases)
	testUnmarshal(t, opts, defaultOptsUnmarshalTestCases)
	testUnmarshal(t, opts, applicationExtensionsUnmarshalTestCases)
}

var timestampExtensionOverrideUnmarshalTestCases = []unmarshalTestCase{
	// Timestamp extension type (-1):
	// - timestamp 32
	//   (as fixext 4, which is canonical/minimal)
	// - timestamp 64
	//   (as fixext 8, which is canonical/minimal)
	{encoded: []byte{0xd7, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: "tomorrow"},
	// - timestamp 96
	//   (as ext 8, which is canonical/minimal)
	{encoded: []byte{0xc7, 0x0c, 0xff, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}, decoded: "tomorrow"},
	// - invalid lengths (via ext 8)
	{encoded: []byte{0xc7, 0x00, 0xff}, decoded: "tomorrow"},
}

func TestUnmarshal_timestampExtensionOverride(t *testing.T) {
	opts := &UnmarshalOptions{
		DisableStandardUnmarshalTransformer: true,
		ApplicationUnmarshalTransformer: MakeExtensionTypeUnmarshalTransformer(
			map[int8]UnmarshalExtensionTypeFn{
				// -1: just unmarshals to a string.
				-1: func(data []byte) (any, bool, error) {
					return string("tomorrow"), true, nil
				},
			},
		),
	}
	testUnmarshal(t, opts, commonUnmarshalTestCases)
	testUnmarshal(t, opts, defaultOptsUnmarshalTestCases)
	testUnmarshal(t, opts, timestampExtensionOverrideUnmarshalTestCases)
}

// TODO: test MakeExtensionTypeUnmarshalTransformer.
