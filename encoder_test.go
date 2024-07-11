// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests encoder.go.

package umsgpack_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"testing"
	"time"

	. "github.com/viettrungluu/umsgpack"
)

// A marshalTestCase defines a test case for marshalling: the original object and the expected
// encoded bytes or the expected error. If prefix is true, then encoded is just a prefix to be
// checked; the actual encoded data is the checked by unmarshalling. (This is to support testing
// map marshalling, since Go's map iteration order is not deterministic.)
type marshalTestCase struct {
	obj     any
	encoded []byte
	err     error
	prefix  bool
}

// testMarshal is a helper for testing Marshal with the given options for the given test cases.
func testMarshal(t *testing.T, opts *MarshalOptions, tCs []marshalTestCase) {
	for _, tC := range tCs {
		buf := &bytes.Buffer{}
		if actualErr := Marshal(opts, buf, tC.obj); actualErr != tC.err {
			t.Errorf("unexected error for obj=%#v (encoded=%q, err=%v): actualErr=%v", tC.obj, tC.encoded, tC.err, actualErr)
		} else if tC.err == nil {
			if tC.prefix {
				if bytes.Compare(buf.Bytes()[:len(tC.encoded)], tC.encoded) != 0 {
					t.Errorf("unexected result for obj=%#v (encoded_prefix=%q): actualEncoded=%q", tC.obj, tC.encoded, buf.Bytes())
				} else {
					if decoded, err := UnmarshalBytes(nil, buf.Bytes()); err != nil {
						t.Errorf("unmarshal failed for obj=%#v (err=%v): actualEncoded=%q", tC.obj, err, buf.Bytes())
					} else if !reflect.DeepEqual(decoded, tC.obj) {
						t.Errorf("unmarshal output did not match for obj=%#v: decodedObj=%#v", tC.obj, decoded)
					}
				}
			} else {
				if bytes.Compare(buf.Bytes(), tC.encoded) != 0 {
					t.Errorf("unexected result for obj=%#v (encoded=%q): actualEncoded=%q", tC.obj, tC.encoded, buf.Bytes())
				}
			}
		}
	}
}

// commonMarshalTestCases contains marshalTestCases that should pass regardless of options.
var commonMarshalTestCases = []marshalTestCase{
	// nil: 11000000: 0xc0
	{obj: nil, encoded: []byte{0xc0}},
	// *** bool
	// false: 11000010: 0xc2
	{obj: false, encoded: []byte{0xc2}},
	// true: 11000011: 0xc3
	{obj: true, encoded: []byte{0xc3}},
	// *** int (which we require to be 64-bit)
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int(0), encoded: []byte{0x00}},
	{obj: int(0x42), encoded: []byte{0x42}},
	{obj: int(0x7f), encoded: []byte{0x7f}},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int(-1), encoded: []byte{0xff}},
	{obj: int(-32), encoded: []byte{0xe0}},
	// int 8: 11010000: 0xd0
	{obj: int(-33), encoded: []byte{0xd0, 0xdf}},
	{obj: int(math.MinInt8), encoded: []byte{0xd0, 0x80}},
	// int 16: 11010001: 0xd1
	{obj: int(math.MaxInt8 + 1), encoded: []byte{0xd1, 0x00, 0x80}},
	{obj: int(math.MaxInt16), encoded: []byte{0xd1, 0x7f, 0xff}},
	{obj: int(math.MinInt8 - 1), encoded: []byte{0xd1, 0xff, 0x7f}},
	{obj: int(math.MinInt16), encoded: []byte{0xd1, 0x80, 0x00}},
	// int 32: 11010010: 0xd2
	{obj: int(math.MaxInt16 + 1), encoded: []byte{0xd2, 0x00, 0x00, 0x80, 0x00}},
	{obj: int(math.MaxInt32), encoded: []byte{0xd2, 0x7f, 0xff, 0xff, 0xff}},
	{obj: int(math.MinInt16 - 1), encoded: []byte{0xd2, 0xff, 0xff, 0x7f, 0xff}},
	{obj: int(math.MinInt32), encoded: []byte{0xd2, 0x80, 0x00, 0x00, 0x00}},
	// int 64: 11010011: 0xd3
	{obj: int(math.MaxInt32 + 1), encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00}},
	{obj: int(math.MaxInt64), encoded: []byte{0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	{obj: int(math.MinInt32 - 1), encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff}},
	{obj: int(math.MinInt64), encoded: []byte{0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	// *** int8
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int8(0), encoded: []byte{0x00}},
	{obj: int8(0x42), encoded: []byte{0x42}},
	{obj: int8(0x7f), encoded: []byte{0x7f}},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int8(-1), encoded: []byte{0xff}},
	{obj: int8(-32), encoded: []byte{0xe0}},
	// int 8: 11010000: 0xd0
	{obj: int8(-33), encoded: []byte{0xd0, 0xdf}},
	{obj: int8(math.MinInt8), encoded: []byte{0xd0, 0x80}},
	// *** int16
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int16(0), encoded: []byte{0x00}},
	{obj: int16(0x42), encoded: []byte{0x42}},
	{obj: int16(0x7f), encoded: []byte{0x7f}},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int16(-1), encoded: []byte{0xff}},
	{obj: int16(-32), encoded: []byte{0xe0}},
	// int 8: 11010000: 0xd0
	{obj: int16(-33), encoded: []byte{0xd0, 0xdf}},
	{obj: int16(math.MinInt8), encoded: []byte{0xd0, 0x80}},
	// int 16: 11010001: 0xd1
	{obj: int16(math.MaxInt8 + 1), encoded: []byte{0xd1, 0x00, 0x80}},
	{obj: int16(math.MaxInt16), encoded: []byte{0xd1, 0x7f, 0xff}},
	{obj: int16(math.MinInt8 - 1), encoded: []byte{0xd1, 0xff, 0x7f}},
	{obj: int16(math.MinInt16), encoded: []byte{0xd1, 0x80, 0x00}},
	// *** int32
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int32(0), encoded: []byte{0x00}},
	{obj: int32(0x42), encoded: []byte{0x42}},
	{obj: int32(0x7f), encoded: []byte{0x7f}},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int32(-1), encoded: []byte{0xff}},
	{obj: int32(-32), encoded: []byte{0xe0}},
	// int 8: 11010000: 0xd0
	{obj: int32(-33), encoded: []byte{0xd0, 0xdf}},
	{obj: int32(math.MinInt8), encoded: []byte{0xd0, 0x80}},
	// int 16: 11010001: 0xd1
	{obj: int32(math.MaxInt8 + 1), encoded: []byte{0xd1, 0x00, 0x80}},
	{obj: int32(math.MaxInt16), encoded: []byte{0xd1, 0x7f, 0xff}},
	{obj: int32(math.MinInt8 - 1), encoded: []byte{0xd1, 0xff, 0x7f}},
	{obj: int32(math.MinInt16), encoded: []byte{0xd1, 0x80, 0x00}},
	// int 32: 11010010: 0xd2
	{obj: int32(math.MaxInt16 + 1), encoded: []byte{0xd2, 0x00, 0x00, 0x80, 0x00}},
	{obj: int32(math.MaxInt32), encoded: []byte{0xd2, 0x7f, 0xff, 0xff, 0xff}},
	{obj: int32(math.MinInt16 - 1), encoded: []byte{0xd2, 0xff, 0xff, 0x7f, 0xff}},
	{obj: int32(math.MinInt32), encoded: []byte{0xd2, 0x80, 0x00, 0x00, 0x00}},
	// *** int64
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int64(0), encoded: []byte{0x00}},
	{obj: int64(0x42), encoded: []byte{0x42}},
	{obj: int64(0x7f), encoded: []byte{0x7f}},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int64(-1), encoded: []byte{0xff}},
	{obj: int64(-32), encoded: []byte{0xe0}},
	// int 8: 11010000: 0xd0
	{obj: int64(-33), encoded: []byte{0xd0, 0xdf}},
	{obj: int64(math.MinInt8), encoded: []byte{0xd0, 0x80}},
	// int 16: 11010001: 0xd1
	{obj: int64(math.MaxInt8 + 1), encoded: []byte{0xd1, 0x00, 0x80}},
	{obj: int64(math.MaxInt16), encoded: []byte{0xd1, 0x7f, 0xff}},
	{obj: int64(math.MinInt8 - 1), encoded: []byte{0xd1, 0xff, 0x7f}},
	{obj: int64(math.MinInt16), encoded: []byte{0xd1, 0x80, 0x00}},
	// int 32: 11010010: 0xd2
	{obj: int64(math.MaxInt16 + 1), encoded: []byte{0xd2, 0x00, 0x00, 0x80, 0x00}},
	{obj: int64(math.MaxInt32), encoded: []byte{0xd2, 0x7f, 0xff, 0xff, 0xff}},
	{obj: int64(math.MinInt16 - 1), encoded: []byte{0xd2, 0xff, 0xff, 0x7f, 0xff}},
	{obj: int64(math.MinInt32), encoded: []byte{0xd2, 0x80, 0x00, 0x00, 0x00}},
	// int 64: 11010011: 0xd3
	{obj: int64(math.MaxInt32 + 1), encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00}},
	{obj: int64(math.MaxInt64), encoded: []byte{0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	{obj: int64(math.MinInt32 - 1), encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff}},
	{obj: int64(math.MinInt64), encoded: []byte{0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	// *** uint
	// uint 8: 11001100: 0xcc
	{obj: uint(0), encoded: []byte{0xcc, 0x00}},
	{obj: uint(0x42), encoded: []byte{0xcc, 0x42}},
	{obj: uint(math.MaxUint8), encoded: []byte{0xcc, 0xff}},
	// uint 16: 11001101: 0xcd
	{obj: uint(math.MaxUint8 + 1), encoded: []byte{0xcd, 0x01, 0x00}},
	{obj: uint(math.MaxUint16), encoded: []byte{0xcd, 0xff, 0xff}},
	// uint 32: 11001110: 0xce
	{obj: uint(math.MaxUint16 + 1), encoded: []byte{0xce, 0x00, 0x01, 0x00, 0x00}},
	{obj: uint(math.MaxUint32), encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xff}},
	// uint 64: 11001111: 0xcf
	{obj: uint(math.MaxUint32 + 1), encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}},
	{obj: uint(math.MaxUint64), encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	// *** uint8
	// uint 8: 11001100: 0xcc
	{obj: uint8(0), encoded: []byte{0xcc, 0x00}},
	{obj: uint8(0x42), encoded: []byte{0xcc, 0x42}},
	{obj: uint8(math.MaxUint8), encoded: []byte{0xcc, 0xff}},
	// *** uint16
	// uint 8: 11001100: 0xcc
	{obj: uint16(0), encoded: []byte{0xcc, 0x00}},
	{obj: uint16(0x42), encoded: []byte{0xcc, 0x42}},
	{obj: uint16(math.MaxUint8), encoded: []byte{0xcc, 0xff}},
	// uint 16: 11001101: 0xcd
	{obj: uint16(math.MaxUint8 + 1), encoded: []byte{0xcd, 0x01, 0x00}},
	{obj: uint16(math.MaxUint16), encoded: []byte{0xcd, 0xff, 0xff}},
	// *** uint32
	// uint 8: 11001100: 0xcc
	{obj: uint32(0), encoded: []byte{0xcc, 0x00}},
	{obj: uint32(0x42), encoded: []byte{0xcc, 0x42}},
	{obj: uint32(math.MaxUint8), encoded: []byte{0xcc, 0xff}},
	// uint 16: 11001101: 0xcd
	{obj: uint32(math.MaxUint8 + 1), encoded: []byte{0xcd, 0x01, 0x00}},
	{obj: uint32(math.MaxUint16), encoded: []byte{0xcd, 0xff, 0xff}},
	// uint 32: 11001110: 0xce
	{obj: uint32(math.MaxUint16 + 1), encoded: []byte{0xce, 0x00, 0x01, 0x00, 0x00}},
	{obj: uint32(math.MaxUint32), encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xff}},
	// *** uint64
	// uint 8: 11001100: 0xcc
	{obj: uint64(0), encoded: []byte{0xcc, 0x00}},
	{obj: uint64(0x42), encoded: []byte{0xcc, 0x42}},
	{obj: uint64(math.MaxUint8), encoded: []byte{0xcc, 0xff}},
	// uint 16: 11001101: 0xcd
	{obj: uint64(math.MaxUint8 + 1), encoded: []byte{0xcd, 0x01, 0x00}},
	{obj: uint64(math.MaxUint16), encoded: []byte{0xcd, 0xff, 0xff}},
	// uint 32: 11001110: 0xce
	{obj: uint64(math.MaxUint16 + 1), encoded: []byte{0xce, 0x00, 0x01, 0x00, 0x00}},
	{obj: uint64(math.MaxUint32), encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xff}},
	// uint 64: 11001111: 0xcf
	{obj: uint64(math.MaxUint32 + 1), encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}},
	{obj: uint64(math.MaxUint64), encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	// *** uintptr (assumed to be 64-bit here)
	// uint 8: 11001100: 0xcc
	{obj: uintptr(0), encoded: []byte{0xcc, 0x00}},
	{obj: uintptr(0x42), encoded: []byte{0xcc, 0x42}},
	{obj: uintptr(math.MaxUint8), encoded: []byte{0xcc, 0xff}},
	// uint 16: 11001101: 0xcd
	{obj: uintptr(math.MaxUint8 + 1), encoded: []byte{0xcd, 0x01, 0x00}},
	{obj: uintptr(math.MaxUint16), encoded: []byte{0xcd, 0xff, 0xff}},
	// uint 32: 11001110: 0xce
	{obj: uintptr(math.MaxUint16 + 1), encoded: []byte{0xce, 0x00, 0x01, 0x00, 0x00}},
	{obj: uintptr(math.MaxUint32), encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xff}},
	// uint 64: 11001111: 0xcf
	{obj: uintptr(math.MaxUint32 + 1), encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}},
	{obj: uintptr(math.MaxUint64), encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	// *** float32
	// float 32: 11001010: 0xca
	{obj: float32(0), encoded: []byte{0xca, 0x00, 0x00, 0x00, 0x00}},
	{obj: float32(math.SmallestNonzeroFloat32), encoded: []byte{0xca, 0x00, 0x00, 0x00, 0x01}},
	{obj: float32(0.75), encoded: []byte{0xca, 0x3f, 0x40, 0x00, 0x00}},
	{obj: float32(1), encoded: []byte{0xca, 0x3f, 0x80, 0x00, 0x00}},
	{obj: math.Nextafter32(1, 2), encoded: []byte{0xca, 0x3f, 0x80, 0x00, 0x01}},
	{obj: float32(math.MaxFloat32), encoded: []byte{0xca, 0x7f, 0x7f, 0xff, 0xff}},
	{obj: float32(math.Inf(1)), encoded: []byte{0xca, 0x7f, 0x80, 0x00, 0x00}},
	{obj: float32(-1), encoded: []byte{0xca, 0xbf, 0x80, 0x00, 0x00}},
	{obj: float32(math.Inf(-1)), encoded: []byte{0xca, 0xff, 0x80, 0x00, 0x00}},
	// *** float64
	// float 64: 11001011: 0xcb
	{obj: float64(0), encoded: []byte{0xcb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{obj: float64(math.SmallestNonzeroFloat64), encoded: []byte{0xcb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
	{obj: float64(0.75), encoded: []byte{0xcb, 0x3f, 0xe8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{obj: float64(1), encoded: []byte{0xcb, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{obj: math.Nextafter(1, 2), encoded: []byte{0xcb, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
	{obj: float64(math.MaxFloat64), encoded: []byte{0xcb, 0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	{obj: math.Inf(1), encoded: []byte{0xcb, 0x7f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{obj: float64(-1), encoded: []byte{0xcb, 0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{obj: math.Inf(-1), encoded: []byte{0xcb, 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	// *** string
	// fixstr: 101xxxxx: 0xa0 - 0xbf
	{obj: "", encoded: []byte{0xa0}},
	{obj: "h", encoded: []byte{0xa1, 0x68}},
	{obj: "hi", encoded: []byte{0xa2, 0x68, 0x69}},
	{obj: string(fillerChars(31)), encoded: append([]byte{0xbf}, fillerChars(31)...)},
	// str 8: 11011001: 0xd9
	{obj: string(fillerChars(32)), encoded: append([]byte{0xd9, 0x20}, fillerChars(32)...)},
	{obj: string(fillerChars(0xff)), encoded: append([]byte{0xd9, 0xff}, fillerChars(0xff)...)},
	// str 16: 11011010: 0xda
	{obj: string(fillerChars(0x100)), encoded: append([]byte{0xda, 0x01, 0x00}, fillerChars(0x100)...)},
	{obj: string(fillerChars(0xffff)), encoded: append([]byte{0xda, 0xff, 0xff}, fillerChars(0xffff)...)},
	// str 32: 11011011: 0xdb
	{obj: string(fillerChars(0x10000)), encoded: append([]byte{0xdb, 0x00, 0x01, 0x00, 0x00}, fillerChars(0x10000)...)},
	{obj: string(fillerChars(99999)), encoded: append([]byte{0xdb, 0x00, 0x01, 0x86, 0x9f}, fillerChars(99999)...)},
	// *** []byte
	// bin 8: 11000100: 0xc4
	{obj: []byte{}, encoded: []byte{0xc4, 0x00}},
	{obj: []byte{0x00}, encoded: []byte{0xc4, 0x01, 0x00}},
	{obj: []byte{0x00, 0x01}, encoded: []byte{0xc4, 0x02, 0x00, 0x01}},
	{obj: fillerBytes(0xff), encoded: append([]byte{0xc4, 0xff}, fillerBytes(0xff)...)},
	// bin 16: 11000101: 0xc5
	{obj: fillerBytes(0x100), encoded: append([]byte{0xc5, 0x01, 0x00}, fillerBytes(0x100)...)},
	{obj: fillerBytes(0xffff), encoded: append([]byte{0xc5, 0xff, 0xff}, fillerBytes(0xffff)...)},
	// bin 32: 11000110: 0xc6
	{obj: fillerBytes(0x10000), encoded: append([]byte{0xc6, 0x00, 0x01, 0x00, 0x00}, fillerBytes(0x10000)...)},
	{obj: fillerBytes(99999), encoded: append([]byte{0xc6, 0x00, 0x01, 0x86, 0x9f}, fillerBytes(99999)...)},
	// *** []any
	// fixarray: 1001xxxx: 0x90 - 0x9f
	{obj: []any{}, encoded: []byte{0x90}},
	{obj: genArray(1), encoded: append([]byte{0x91}, genArrayData(1)...)},
	{obj: genArray(2), encoded: append([]byte{0x92}, genArrayData(2)...)},
	{obj: genArray(0xf), encoded: append([]byte{0x9f}, genArrayData(0xf)...)},
	// array 16: 11011100: 0xdc
	{obj: genArray(0x10), encoded: append([]byte{0xdc, 0x00, 0x10}, genArrayData(0x10)...)},
	{obj: genArray(0xffff), encoded: append([]byte{0xdc, 0xff, 0xff}, genArrayData(0xffff)...)},
	// array 32: 11011101: 0xdd
	{obj: genArray(0x10000), encoded: append([]byte{0xdd, 0x00, 0x01, 0x00, 0x00}, genArrayData(0x10000)...)},
	{obj: genArray(99999), encoded: append([]byte{0xdd, 0x00, 0x01, 0x86, 0x9f}, genArrayData(99999)...)},
	// *** map[any]any
	// fixmap: 1000xxxx: 0x80 - 0x8f
	{obj: map[any]any{}, encoded: []byte{0x80}},
	{obj: genMap(1), encoded: append([]byte{0x81}, genMapData(1)...)},
	{obj: genMap(2), encoded: []byte{0x82}, prefix: true},
	{obj: genMap(0xf), encoded: []byte{0x8f}, prefix: true},
	// map 16: 11011110: 0xde
	{obj: genMap(0x10), encoded: []byte{0xde, 0x00, 0x10}, prefix: true},
	{obj: genMap(0xffff), encoded: []byte{0xde, 0xff, 0xff}, prefix: true},
	// map 32: 11011111: 0xdf
	{obj: genMap(0x10000), encoded: []byte{0xdf, 0x00, 0x01, 0x00, 0x00}, prefix: true},
	{obj: genMap(99999), encoded: []byte{0xdf, 0x00, 0x01, 0x86, 0x9f}, prefix: true},
	// *** *UnresolvedExtensionType
	// fixext 1: 11010100: 0xd4
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00}}, encoded: []byte{0xd4, 0x12, 0x00}},
	// fixext 2: 11010101: 0xd5
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01}}, encoded: []byte{0xd5, 0x12, 0x00, 0x01}},
	// fixext 4: 11010110: 0xd6
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01, 0x02, 0x03}}, encoded: []byte{0xd6, 0x12, 0x00, 0x01, 0x02, 0x03}},
	// fixext 8: 11010111: 0xd7
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(8)}, encoded: append([]byte{0xd7, 0x12}, fillerBytes(8)...)},
	// fixext 16: 11011000: 0xd8
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(16)}, encoded: append([]byte{0xd8, 0x12}, fillerBytes(16)...)},
	// ext 8: 11000111: 0xc7
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{}}, encoded: []byte{0xc7, 0x00, 0x12}},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01, 0x02}}, encoded: []byte{0xc7, 0x03, 0x12, 0x00, 0x01, 0x02}},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(5)}, encoded: append([]byte{0xc7, 0x05, 0x12}, fillerBytes(5)...)},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(7)}, encoded: append([]byte{0xc7, 0x07, 0x12}, fillerBytes(7)...)},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(9)}, encoded: append([]byte{0xc7, 0x09, 0x12}, fillerBytes(9)...)},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(15)}, encoded: append([]byte{0xc7, 0x0f, 0x12}, fillerBytes(15)...)},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(17)}, encoded: append([]byte{0xc7, 0x11, 0x12}, fillerBytes(17)...)},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(math.MaxUint8)}, encoded: append([]byte{0xc7, 0xff, 0x12}, fillerBytes(math.MaxUint8)...)},
	// ext 16: 11001000: 0xc8
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(math.MaxUint8 + 1)}, encoded: append([]byte{0xc8, 0x01, 0x00, 0x12}, fillerBytes(math.MaxUint8+1)...)},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(math.MaxUint16)}, encoded: append([]byte{0xc8, 0xff, 0xff, 0x12}, fillerBytes(math.MaxUint16)...)},
	// ext 32: 11001001: 0xc9
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(math.MaxUint16 + 1)}, encoded: append([]byte{0xc9, 0x00, 0x01, 0x00, 0x00, 0x12}, fillerBytes(math.MaxUint16+1)...)},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(99999)}, encoded: append([]byte{0xc9, 0x00, 0x01, 0x86, 0x9f, 0x12}, fillerBytes(99999)...)},
	// *** []string
	// fixarray: 1001xxxx: 0x90 - 0x9f
	{obj: []string{}, encoded: []byte{0x90}},
	{obj: genTypedArray(1), encoded: append([]byte{0x91}, genArrayData(1)...)},
	{obj: genTypedArray(2), encoded: append([]byte{0x92}, genArrayData(2)...)},
	{obj: genTypedArray(0xf), encoded: append([]byte{0x9f}, genArrayData(0xf)...)},
	// array 16: 11011100: 0xdc
	{obj: genTypedArray(0x10), encoded: append([]byte{0xdc, 0x00, 0x10}, genArrayData(0x10)...)},
	{obj: genTypedArray(0xffff), encoded: append([]byte{0xdc, 0xff, 0xff}, genArrayData(0xffff)...)},
	// array 32: 11011101: 0xdd
	{obj: genTypedArray(0x10000), encoded: append([]byte{0xdd, 0x00, 0x01, 0x00, 0x00}, genArrayData(0x10000)...)},
	{obj: genTypedArray(99999), encoded: append([]byte{0xdd, 0x00, 0x01, 0x86, 0x9f}, genArrayData(99999)...)},
	// *** [n]string
	// fixarray: 1001xxxx: 0x90 - 0x9f
	{obj: [0]string{}, encoded: []byte{0x90}},
	{obj: [4]string{"0", "1", "2", "3"}, encoded: append([]byte{0x94}, genArrayData(4)...)},
	// (skip testing other formats; should be handled like slices)
	// *** map[string]int
	// fixmap: 1000xxxx: 0x80 - 0x8f
	{obj: map[string]int{}, encoded: []byte{0x80}},
	{obj: genTypedMap(1), encoded: append([]byte{0x81}, genMapData(1)...)},
	// TODO: comparing decoded doesn't work!
	// {obj: genTypedMap(2), encoded: []byte{0x82}, prefix: true},
	// {obj: genTypedMap(0xf), encoded: []byte{0x8f}, prefix: true},
	// map 16: 11011110: 0xde
	// {obj: genTypedMap(0x10), encoded: []byte{0xde, 0x00, 0x10}, prefix: true},
	// {obj: genTypedMap(0xffff), encoded: []byte{0xde, 0xff, 0xff}, prefix: true},
	// map 32: 11011111: 0xdf
	// {obj: genTypedMap(0x10000), encoded: []byte{0xdf, 0x00, 0x01, 0x00, 0x00}, prefix: true},
	// {obj: genTypedMap(99999), encoded: []byte{0xdf, 0x00, 0x01, 0x86, 0x9f}, prefix: true},
	// *** time.Time
	// timestamp 32
	{obj: time.Unix(0, 0), encoded: []byte{0xd6, 0xff, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(math.MaxUint32, 0), encoded: []byte{0xd6, 0xff, 0xff, 0xff, 0xff, 0xff}},
	// timestamp 64
	{obj: time.Unix(math.MaxUint32+1, 0), encoded: []byte{0xd7, 0xff, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(1<<34-1, 0), encoded: []byte{0xd7, 0xff, 0x00, 0x00, 0x00, 0x03, 0xff, 0xff, 0xff, 0xff}},
	{obj: time.Unix(0, 1), encoded: []byte{0xd7, 0xff, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(0, 999999999), encoded: []byte{0xd7, 0xff, 0xee, 0x6b, 0x27, 0xfc, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(1<<34-1, 999999999), encoded: []byte{0xd7, 0xff, 0xee, 0x6b, 0x27, 0xff, 0xff, 0xff, 0xff, 0xff}},
	// timestamp 96
	{obj: time.Unix(1<<34, 0), encoded: []byte{0xc7, 0x0c, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(math.MaxInt64, 0), encoded: []byte{0xc7, 0x0c, 0xff, 0x00, 0x00, 0x00, 0x00, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	{obj: time.Unix(1<<34, 1), encoded: []byte{0xc7, 0x0c, 0xff, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(1<<34, 999999999), encoded: []byte{0xc7, 0x0c, 0xff, 0x3b, 0x9a, 0xc9, 0xff, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(-1, 0), encoded: []byte{0xc7, 0x0c, 0xff, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	{obj: time.Unix(math.MinInt64, 0), encoded: []byte{0xc7, 0x0c, 0xff, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(-1, 1), encoded: []byte{0xc7, 0x0c, 0xff, 0x00, 0x00, 0x00, 0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	{obj: time.Unix(-1, 999999999), encoded: []byte{0xc7, 0x0c, 0xff, 0x3b, 0x9a, 0xc9, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	{obj: time.Unix(math.MinInt64, 1), encoded: []byte{0xc7, 0x0c, 0xff, 0x00, 0x00, 0x00, 0x01, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{obj: time.Unix(math.MinInt64, 999999999), encoded: []byte{0xc7, 0x0c, 0xff, 0x3b, 0x9a, 0xc9, 0xff, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	// *** Errors
	{obj: chan int(nil), err: UnsupportedTypeForMarshallingError},
}

type testMarshalType1 string

type testMarshalType2 struct{}

type testMarshalType3 struct {
	value bool
}

type testMarshalType4 int

type testMarshalType5 int

var defaultOptsMarshalTestCases = []marshalTestCase{
	{obj: testMarshalType1(""), err: UnsupportedTypeForMarshallingError},
	{obj: &testMarshalType2{}, err: UnsupportedTypeForMarshallingError},
	{obj: &testMarshalType3{}, err: UnsupportedTypeForMarshallingError},
	{obj: testMarshalType4(0), err: UnsupportedTypeForMarshallingError},
	{obj: testMarshalType5(0), err: UnsupportedTypeForMarshallingError},
}

var applicationTransformersMarshalTestCases = []marshalTestCase{
	// *** testMarshalType1
	{obj: testMarshalType1("oops"), err: testError},
	// fixext 1: 11010100: 0xd4
	{obj: testMarshalType1("h"), encoded: []byte{0xd4, 0x0c, 0x68}},
	// fixext 2: 11010101: 0xd5
	{obj: testMarshalType1("hi"), encoded: []byte{0xd5, 0x0c, 0x68, 0x69}},
	// fixext 4: 11010110: 0xd6
	{obj: testMarshalType1("hihi"), encoded: []byte{0xd6, 0x0c, 0x68, 0x69, 0x68, 0x69}},
	// fixext 8: 11010111: 0xd7
	{obj: testMarshalType1(string(fillerChars(8))), encoded: append([]byte{0xd7, 0x0c}, fillerChars(8)...)},
	// fixext 16: 11011000: 0xd8
	{obj: testMarshalType1(string(fillerChars(16))), encoded: append([]byte{0xd8, 0x0c}, fillerChars(16)...)},
	// ext 8: 11000111: 0xc7
	{obj: testMarshalType1(""), encoded: []byte{0xc7, 0x00, 0x0c}},
	{obj: testMarshalType1(string(fillerChars(42))), encoded: append([]byte{0xc7, 0x2a, 0x0c}, fillerChars(42)...)},
	// ext 16: 11001000: 0xc8
	{obj: testMarshalType1(string(fillerChars(256))), encoded: append([]byte{0xc8, 0x01, 0x00, 0x0c}, fillerChars(256)...)},
	// ext 32: 11001001: 0xc9
	{obj: testMarshalType1(string(fillerChars(99999))), encoded: append([]byte{0xc9, 0x00, 0x01, 0x86, 0x9f, 0x0c}, fillerChars(99999)...)},
	// *** testMarshalType2
	{obj: &testMarshalType2{}, encoded: []byte{0xd5, 0x2a, 0x68, 0x69}},
	// *** testMarshalType3
	{obj: &testMarshalType3{}, err: UnsupportedTypeForMarshallingError},
	// *** testMarshalType4
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: testMarshalType4(0x12), encoded: []byte{0x12}},
	// *** testMarshalType5
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: testMarshalType5(0x12), encoded: []byte{0x12}},
	// *** other
	{obj: map[any]any{testMarshalType1("hi"): &testMarshalType2{}}, encoded: []byte{0x81, 0xd5, 0x0c, 0x68, 0x69, 0xd5, 0x2a, 0x68, 0x69}},
}

// A marshalWriteErrorTestCase defines a test case for marshalling write errors: the original object
// and where the write error will occur.
type marshalWriteErrorTestCase struct {
	obj   any
	errAt int
}

// A *limitedDiscardWriter is an io.Writer that discards up to a given number of bytes. It'll return
// io.ErrShortWrite if too much is written.
type limitedDiscardWriter struct {
	left int
}

var _ io.Writer = (*limitedDiscardWriter)(nil)

func (w *limitedDiscardWriter) Write(p []byte) (n int, err error) {
	w.left -= len(p)
	if w.left < 0 {
		return len(p) + w.left, io.ErrShortWrite
	} else {
		return len(p), nil
	}
}

// testMarshalWriteError is a helper for testing Marshal with the given options for the given
// write-error test cases.
func testMarshalWriteError(t *testing.T, opts *MarshalOptions, tCs []marshalWriteErrorTestCase) {
	for _, tC := range tCs {
		writer := &limitedDiscardWriter{tC.errAt}
		if err := Marshal(opts, writer, tC.obj); err != io.ErrShortWrite {
			t.Errorf("unexected error for obj=%#v (errAt=%v): err=%v", tC.obj, tC.errAt, err)
		}
	}
}

var commonMarshalWriteErrorTestCases = []marshalWriteErrorTestCase{
	// nil: 11000000: 0xc0
	{obj: nil, errAt: 0},
	// *** bool
	// false: 11000010: 0xc2
	{obj: false, errAt: 0},
	// true: 11000011: 0xc3
	{obj: true, errAt: 0},
	// *** int (which we require to be 64-bit)
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int(42), errAt: 0},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int(-12), errAt: 0},
	// int 8: 11010000: 0xd0
	{obj: int(-42), errAt: 0},
	{obj: int(-42), errAt: 1},
	// int 16: 11010001: 0xd1
	{obj: int(1234), errAt: 0},
	{obj: int(1234), errAt: 1},
	{obj: int(1234), errAt: 2},
	// int 32: 11010010: 0xd2
	{obj: int(123456), errAt: 0},
	{obj: int(123456), errAt: 1},
	{obj: int(123456), errAt: 4},
	// int 64: 11010011: 0xd3
	{obj: int(9876543210), errAt: 0},
	{obj: int(9876543210), errAt: 1},
	{obj: int(9876543210), errAt: 8},
	// *** int8
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int8(42), errAt: 0},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int8(-12), errAt: 0},
	// int 8: 11010000: 0xd0
	{obj: int8(-42), errAt: 0},
	{obj: int8(-42), errAt: 1},
	// *** int16
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int16(42), errAt: 0},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int16(-12), errAt: 0},
	// int 8: 11010000: 0xd0
	{obj: int16(-42), errAt: 0},
	{obj: int16(-42), errAt: 1},
	// int 16: 11010001: 0xd1
	{obj: int16(1234), errAt: 0},
	{obj: int16(1234), errAt: 1},
	{obj: int16(1234), errAt: 2},
	// *** int32
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int32(42), errAt: 0},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int32(-12), errAt: 0},
	// int 8: 11010000: 0xd0
	{obj: int32(-42), errAt: 0},
	{obj: int32(-42), errAt: 1},
	// int 16: 11010001: 0xd1
	{obj: int32(1234), errAt: 0},
	{obj: int32(1234), errAt: 1},
	{obj: int32(1234), errAt: 2},
	// int 32: 11010010: 0xd2
	{obj: int32(123456), errAt: 0},
	{obj: int32(123456), errAt: 1},
	{obj: int32(123456), errAt: 4},
	// *** int64
	// positive fixint: 0xxxxxxx: 0x00 - 0x7f
	{obj: int64(42), errAt: 0},
	// negative fixint: 111xxxxx: 0xe0 - 0xff
	{obj: int64(-12), errAt: 0},
	// int 8: 11010000: 0xd0
	{obj: int64(-42), errAt: 0},
	{obj: int64(-42), errAt: 1},
	// int 16: 11010001: 0xd1
	{obj: int64(1234), errAt: 0},
	{obj: int64(1234), errAt: 1},
	{obj: int64(1234), errAt: 2},
	// int 32: 11010010: 0xd2
	{obj: int64(123456), errAt: 0},
	{obj: int64(123456), errAt: 1},
	{obj: int64(123456), errAt: 4},
	// int 64: 11010011: 0xd3
	{obj: int64(9876543210), errAt: 0},
	{obj: int64(9876543210), errAt: 1},
	{obj: int64(9876543210), errAt: 8},
	// *** uint
	// uint 8: 11001100: 0xcc
	{obj: uint(42), errAt: 0},
	{obj: uint(42), errAt: 1},
	// uint 16: 11001101: 0xcd
	{obj: uint(1234), errAt: 0},
	{obj: uint(1234), errAt: 1},
	{obj: uint(1234), errAt: 2},
	// uint 32: 11001110: 0xce
	{obj: uint(123456), errAt: 0},
	{obj: uint(123456), errAt: 1},
	{obj: uint(123456), errAt: 4},
	// uint 64: 11001111: 0xcf
	{obj: uint(9876543210), errAt: 0},
	{obj: uint(9876543210), errAt: 1},
	{obj: uint(9876543210), errAt: 8},
	// *** uint8
	// uint 8: 11001100: 0xcc
	{obj: uint8(42), errAt: 0},
	{obj: uint8(42), errAt: 1},
	// *** uint16
	// uint 8: 11001100: 0xcc
	{obj: uint16(42), errAt: 0},
	{obj: uint16(42), errAt: 1},
	// uint 16: 11001101: 0xcd
	{obj: uint16(1234), errAt: 0},
	{obj: uint16(1234), errAt: 1},
	{obj: uint16(1234), errAt: 2},
	// *** uint32
	// uint 8: 11001100: 0xcc
	{obj: uint32(42), errAt: 0},
	{obj: uint32(42), errAt: 1},
	// uint 16: 11001101: 0xcd
	{obj: uint32(1234), errAt: 0},
	{obj: uint32(1234), errAt: 1},
	{obj: uint32(1234), errAt: 2},
	// uint 32: 11001110: 0xce
	{obj: uint32(123456), errAt: 0},
	{obj: uint32(123456), errAt: 1},
	{obj: uint32(123456), errAt: 4},
	// *** uint64
	// uint 8: 11001100: 0xcc
	{obj: uint64(42), errAt: 0},
	{obj: uint64(42), errAt: 1},
	// uint 16: 11001101: 0xcd
	{obj: uint64(1234), errAt: 0},
	{obj: uint64(1234), errAt: 1},
	{obj: uint64(1234), errAt: 2},
	// uint 32: 11001110: 0xce
	{obj: uint64(123456), errAt: 0},
	{obj: uint64(123456), errAt: 1},
	{obj: uint64(123456), errAt: 4},
	// uint 64: 11001111: 0xcf
	{obj: uint64(9876543210), errAt: 0},
	{obj: uint64(9876543210), errAt: 1},
	{obj: uint64(9876543210), errAt: 8},
	// *** uintptr (assumed to be 64-bit here)
	// uint 8: 11001100: 0xcc
	{obj: uintptr(42), errAt: 0},
	{obj: uintptr(42), errAt: 1},
	// uint 16: 11001101: 0xcd
	{obj: uintptr(1234), errAt: 0},
	{obj: uintptr(1234), errAt: 1},
	{obj: uintptr(1234), errAt: 2},
	// uint 32: 11001110: 0xce
	{obj: uintptr(123456), errAt: 0},
	{obj: uintptr(123456), errAt: 1},
	{obj: uintptr(123456), errAt: 4},
	// uint 64: 11001111: 0xcf
	{obj: uintptr(9876543210), errAt: 0},
	{obj: uintptr(9876543210), errAt: 1},
	{obj: uintptr(9876543210), errAt: 8},
	// *** float32
	// float 32: 11001010: 0xca
	{obj: float32(12.34), errAt: 0},
	{obj: float32(12.34), errAt: 1},
	{obj: float32(12.34), errAt: 4},
	// *** float64
	// float 64: 11001011: 0xcb
	{obj: float64(12.34), errAt: 0},
	{obj: float64(12.34), errAt: 1},
	{obj: float64(12.34), errAt: 8},
	// *** string
	// fixstr: 101xxxxx: 0xa0 - 0xbf
	{obj: "hi", errAt: 0},
	{obj: "hi", errAt: 1},
	{obj: "hi", errAt: 2},
	// str 8: 11011001: 0xd9
	{obj: string(fillerChars(56)), errAt: 0},
	{obj: string(fillerChars(56)), errAt: 1},
	{obj: string(fillerChars(56)), errAt: 2},
	{obj: string(fillerChars(56)), errAt: 57},
	// str 16: 11011010: 0xda
	{obj: string(fillerChars(1234)), errAt: 0},
	{obj: string(fillerChars(1234)), errAt: 1},
	{obj: string(fillerChars(1234)), errAt: 2},
	{obj: string(fillerChars(1234)), errAt: 3},
	{obj: string(fillerChars(1234)), errAt: 1236},
	// str 32: 11011011: 0xdb
	{obj: string(fillerChars(123456)), errAt: 0},
	{obj: string(fillerChars(123456)), errAt: 1},
	{obj: string(fillerChars(123456)), errAt: 5},
	{obj: string(fillerChars(123456)), errAt: 6},
	{obj: string(fillerChars(123456)), errAt: 123460},
	// *** []byte
	// bin 8: 11000100: 0xc4
	{obj: fillerBytes(56), errAt: 0},
	{obj: fillerBytes(56), errAt: 1},
	{obj: fillerBytes(56), errAt: 2},
	{obj: fillerBytes(56), errAt: 57},
	// bin 16: 11000101: 0xc5
	{obj: fillerBytes(1234), errAt: 0},
	{obj: fillerBytes(1234), errAt: 1},
	{obj: fillerBytes(1234), errAt: 2},
	{obj: fillerBytes(1234), errAt: 3},
	{obj: fillerBytes(1234), errAt: 1236},
	// bin 32: 11000110: 0xc6
	{obj: fillerBytes(123456), errAt: 0},
	{obj: fillerBytes(123456), errAt: 1},
	{obj: fillerBytes(123456), errAt: 5},
	{obj: fillerBytes(123456), errAt: 6},
	{obj: fillerBytes(123456), errAt: 123460},
	// *** []any
	// fixarray: 1001xxxx: 0x90 - 0x9f
	{obj: genArray(12), errAt: 0},
	{obj: genArray(12), errAt: 1},
	// array 16: 11011100: 0xdc
	{obj: genArray(42), errAt: 0},
	{obj: genArray(42), errAt: 1},
	{obj: genArray(42), errAt: 2},
	{obj: genArray(42), errAt: 3},
	// array 32: 11011101: 0xdd
	{obj: genArray(123456), errAt: 0},
	{obj: genArray(123456), errAt: 1},
	{obj: genArray(123456), errAt: 4},
	{obj: genArray(123456), errAt: 5},
	// *** map[any]any
	// fixmap: 1000xxxx: 0x80 - 0x8f
	{obj: genMap(12), errAt: 0},
	{obj: genMap(12), errAt: 1},
	{obj: genMap(12), errAt: 2},
	{obj: genMap(12), errAt: 3},
	// map 16: 11011110: 0xde
	{obj: genMap(42), errAt: 0},
	{obj: genMap(42), errAt: 1},
	{obj: genMap(42), errAt: 2},
	{obj: genMap(42), errAt: 3},
	{obj: genMap(42), errAt: 4},
	{obj: genMap(42), errAt: 5},
	// map 32: 11011111: 0xdf
	{obj: genMap(123456), errAt: 0},
	{obj: genMap(123456), errAt: 1},
	{obj: genMap(123456), errAt: 4},
	{obj: genMap(123456), errAt: 5},
	{obj: genMap(123456), errAt: 6},
	{obj: genMap(123456), errAt: 7},
	// *** *UnresolvedExtensionType
	// fixext 1: 11010100: 0xd4
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00}}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00}}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00}}, errAt: 2},
	// fixext 2: 11010101: 0xd5
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01}}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01}}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01}}, errAt: 2},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01}}, errAt: 3},
	// fixext 4: 11010110: 0xd6
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01, 0x02, 0x03}}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01, 0x02, 0x03}}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01, 0x02, 0x03}}, errAt: 2},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: []byte{0x00, 0x01, 0x02, 0x03}}, errAt: 5},
	// fixext 8: 11010111: 0xd7
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(8)}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(8)}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(8)}, errAt: 2},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(8)}, errAt: 9},
	// fixext 16: 11011000: 0xd8
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(16)}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(16)}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(16)}, errAt: 2},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(16)}, errAt: 17},
	// ext 8: 11000111: 0xc7
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(42)}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(42)}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(42)}, errAt: 2},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(42)}, errAt: 3},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(42)}, errAt: 44},
	// ext 16: 11001000: 0xc8
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(1234)}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(1234)}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(1234)}, errAt: 2},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(1234)}, errAt: 3},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(1234)}, errAt: 4},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(1234)}, errAt: 1237},
	// ext 32: 11001001: 0xc9
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(123456)}, errAt: 0},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(123456)}, errAt: 1},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(123456)}, errAt: 4},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(123456)}, errAt: 5},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(123456)}, errAt: 6},
	{obj: &UnresolvedExtensionType{ExtensionType: 0x12, Data: fillerBytes(123456)}, errAt: 123461},
	// *** []string
	// fixarray: 1001xxxx: 0x90 - 0x9f
	{obj: genTypedArray(12), errAt: 0},
	{obj: genTypedArray(12), errAt: 1},
	// array 16: 11011100: 0xdc
	{obj: genTypedArray(42), errAt: 0},
	{obj: genTypedArray(42), errAt: 1},
	{obj: genTypedArray(42), errAt: 2},
	{obj: genTypedArray(42), errAt: 3},
	// array 32: 11011101: 0xdd
	{obj: genTypedArray(123456), errAt: 0},
	{obj: genTypedArray(123456), errAt: 1},
	{obj: genTypedArray(123456), errAt: 4},
	{obj: genTypedArray(123456), errAt: 5},
	// *** map[string]int
	// fixmap: 1000xxxx: 0x80 - 0x8f
	{obj: genTypedMap(12), errAt: 0},
	{obj: genTypedMap(12), errAt: 1},
	{obj: genTypedMap(12), errAt: 2},
	{obj: genTypedMap(12), errAt: 3},
	// map 16: 11011110: 0xde
	{obj: genTypedMap(42), errAt: 0},
	{obj: genTypedMap(42), errAt: 1},
	{obj: genTypedMap(42), errAt: 2},
	{obj: genTypedMap(42), errAt: 3},
	{obj: genTypedMap(42), errAt: 4},
	{obj: genTypedMap(42), errAt: 5},
	// map 32: 11011111: 0xdf
	{obj: genTypedMap(123456), errAt: 0},
	{obj: genTypedMap(123456), errAt: 1},
	{obj: genTypedMap(123456), errAt: 4},
	{obj: genTypedMap(123456), errAt: 5},
	{obj: genTypedMap(123456), errAt: 6},
	{obj: genTypedMap(123456), errAt: 7},
	// *** time.Time
	// timestamp 32
	{obj: time.Unix(0, 0), errAt: 0},
	{obj: time.Unix(0, 0), errAt: 1},
	{obj: time.Unix(0, 0), errAt: 2},
	{obj: time.Unix(0, 0), errAt: 5},
	// timestamp 64
	{obj: time.Unix(0, 1), errAt: 0},
	{obj: time.Unix(0, 1), errAt: 1},
	{obj: time.Unix(0, 1), errAt: 2},
	{obj: time.Unix(0, 1), errAt: 9},
	// timestamp 96
	{obj: time.Unix(1<<34, 0), errAt: 0},
	{obj: time.Unix(1<<34, 0), errAt: 1},
	{obj: time.Unix(1<<34, 0), errAt: 2},
	{obj: time.Unix(1<<34, 0), errAt: 3},
	{obj: time.Unix(1<<34, 0), errAt: 14},
}

// TestMarshal_defaultOpts tests Marshal with the default options (all boolean options are false).
func TestMarshal_defaultOpts(t *testing.T) {
	opts := &MarshalOptions{}
	testMarshal(t, opts, commonMarshalTestCases)
	testMarshal(t, opts, defaultOptsMarshalTestCases)
	testMarshalWriteError(t, opts, commonMarshalWriteErrorTestCases)
}

var testError = errors.New("test error")

func TestMarshal_applicationTransformers(t *testing.T) {
	opts := &MarshalOptions{
		LateTransformers: []TransformerFn{
			func(obj any) (any, error) {
				if t, ok := obj.(testMarshalType1); ok {
					if t == "oops" {
						return nil, testError
					}
					return &UnresolvedExtensionType{
						ExtensionType: 12,
						Data:          []byte(t),
					}, nil
				} else {
					return obj, nil
				}
			},
			func(obj any) (any, error) {
				if _, ok := obj.(*testMarshalType2); ok {
					return &UnresolvedExtensionType{
						ExtensionType: 42,
						Data:          []byte("hi"),
					}, nil
				} else {
					return obj, nil
				}
			},
			func(obj any) (any, error) {
				if t, ok := obj.(testMarshalType4); ok {
					return testMarshalType5(t), nil
				} else {
					return obj, nil
				}
			},
			func(obj any) (any, error) {
				if t, ok := obj.(testMarshalType5); ok {
					return int(t), nil
				} else {
					return obj, nil
				}
			},
		},
	}
	testMarshal(t, opts, commonMarshalTestCases)
	testMarshal(t, opts, applicationTransformersMarshalTestCases)
	testMarshalWriteError(t, opts, commonMarshalWriteErrorTestCases)
}

func TestMarshalToBytes(t *testing.T) {
	opts := &MarshalOptions{
		LateTransformers: []TransformerFn{
			func(obj any) (any, error) {
				if t, ok := obj.(testMarshalType1); ok {
					return &UnresolvedExtensionType{
						ExtensionType: 12,
						Data:          []byte(t),
					}, nil
				} else {
					return obj, nil
				}
			},
		},
	}

	if encoded, err := MarshalToBytes(opts, testMarshalType1("hi")); err != nil || bytes.Compare(encoded, []byte{0xd5, 0x0c, 0x68, 0x69}) != 0 {
		t.Errorf("Unexpected result from MarshalToBytes: %v, %v", encoded, err)
	}

	if encoded, err := MarshalToBytes(opts, &testMarshalType2{}); err != UnsupportedTypeForMarshallingError {
		t.Errorf("Unexpected result from MarshalToBytes: %v, %v", encoded, err)
	}
}
