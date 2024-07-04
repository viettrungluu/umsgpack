// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests encoder.go.

package umsgpack_test

import (
	"bytes"
	// "io"
	"math"
	"reflect"
	"testing"
	// "time"

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
	// TODO: test error cases (mostly write failing).
}

// TestMarshal_defaultOpts tests Marshal with the default options (all boolean options are false).
func TestMarshal_defaultOpts(t *testing.T) {
	opts := &MarshalOptions{}
	testMarshal(t, opts, commonMarshalTestCases)
	// TODO: testMarshal(t, opts, defaultOptsMarshalTestCases)
}

// TODO: test application extension types.
