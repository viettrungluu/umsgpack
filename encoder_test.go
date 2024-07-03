// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests encoder.go.

package umsgpack_test

import (
	"bytes"
	// "io"
	"math"
	// "reflect"
	// "strconv"
	"testing"
	// "time"

	. "github.com/viettrungluu/umsgpack"
)

// A marshalTestCase defines a test case for marshalling: the original object and the expected
// encoded bytes or the expected error.
type marshalTestCase struct {
	obj     any
	encoded []byte
	err     error
}

// testMarshal is a helper for testing Marshal with the given options for the given test cases.
func testMarshal(t *testing.T, opts *MarshalOptions, tCs []marshalTestCase) {
	for _, tC := range tCs {
		buf := &bytes.Buffer{}
		if actualErr := Marshal(opts, buf, tC.obj); actualErr != tC.err {
			t.Errorf("unexected error for obj=%#v (encoded=%q, err=%v): actualErr=%v", tC.obj, tC.encoded, tC.err, actualErr)
		} else if tC.err == nil && bytes.Compare(buf.Bytes(), tC.encoded) != 0 {
			t.Errorf("unexected result for obj=%#v (encoded=%q): actualEncoded=%q", tC.obj, tC.encoded, buf.Bytes())
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
	// TODO: uint
	// TODO: uint8
	// TODO: uint16
	// TODO: uint32
	// TODO: uint64
	// TODO: uintptr
}

// TestMarshal_defaultOpts tests Marshal with the default options (all boolean options are false).
func TestMarshal_defaultOpts(t *testing.T) {
	opts := &MarshalOptions{}
	testMarshal(t, opts, commonMarshalTestCases)
	// TODO: testMarshal(t, opts, defaultOptsMarshalTestCases)
}
