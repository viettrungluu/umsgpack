// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests decoder.go.

package umsgpack_test

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

func fillerChars(n int) []byte {
	rv := make([]byte, n)
	for i := 0; i < n; i += 1 {
		rv[i] = byte('0' + i%10)
	}
	return rv
}

func fillerBytes(n int) []byte {
	rv := make([]byte, n)
	for i := 0; i < n; i += 1 {
		rv[i] = byte(i % 256)
	}
	return rv
}

func TestUnmarshal(t *testing.T) {
	opts := &UnmarshalOptions{}
	testCases := []struct {
		encoded []byte
		decoded any
		err     error
	}{
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
		// - int 16:
		{encoded: []byte{0xd1, 0x00, 0x00}, decoded: int(0)},
		{encoded: []byte{0xd1, 0x00, 0x01}, decoded: int(1)},
		{encoded: []byte{0xd1, 0x7f, 0xff}, decoded: int(32767)},
		{encoded: []byte{0xd1, 0xff, 0xff}, decoded: int(-1)},
		{encoded: []byte{0xd1, 0xff, 0xfe}, decoded: int(-2)},
		{encoded: []byte{0xd1, 0x80, 0x00}, decoded: int(-32768)},
		// - int 32:
		{encoded: []byte{0xd2, 0x00, 0x00, 0x00, 0x00}, decoded: int(0)},
		{encoded: []byte{0xd2, 0x00, 0x00, 0x00, 0x01}, decoded: int(1)},
		{encoded: []byte{0xd2, 0x7f, 0xff, 0xff, 0xff}, decoded: int(1<<31 - 1)},
		{encoded: []byte{0xd2, 0xff, 0xff, 0xff, 0xff}, decoded: int(-1)},
		{encoded: []byte{0xd2, 0xff, 0xff, 0xff, 0xfe}, decoded: int(-2)},
		{encoded: []byte{0xd2, 0x80, 0x00, 0x00, 0x00}, decoded: int(-(1 << 31))},
		// - int 64:
		{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: int(0)},
		{encoded: []byte{0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: int(1)},
		{encoded: []byte{0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: int(1<<63 - 1)},
		{encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: int(-1)},
		{encoded: []byte{0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}, decoded: int(-2)},
		{encoded: []byte{0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: int(-(1 << 63))},
		// uint:
		// - uint 8:
		{encoded: []byte{0xcc, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xcc, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xcc, 0xfe}, decoded: uint(254)},
		{encoded: []byte{0xcc, 0xff}, decoded: uint(255)},
		// - uint 16:
		{encoded: []byte{0xcd, 0x00, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xcd, 0x00, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xcd, 0xff, 0xfe}, decoded: uint(65534)},
		{encoded: []byte{0xcd, 0xff, 0xff}, decoded: uint(65535)},
		// - uint 32:
		{encoded: []byte{0xce, 0x00, 0x00, 0x00, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xce, 0x00, 0x00, 0x00, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xfe}, decoded: uint(1<<32 - 2)},
		{encoded: []byte{0xce, 0xff, 0xff, 0xff, 0xff}, decoded: uint(1<<32 - 1)},
		// - uint 64:
		{encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, decoded: uint(0)},
		{encoded: []byte{0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, decoded: uint(1)},
		{encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}, decoded: uint(1<<64 - 2)},
		{encoded: []byte{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, decoded: uint(1<<64 - 1)},
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
		// string:
		// - fixstr:
		{encoded: []byte{0xa0}, decoded: ""},
		{encoded: []byte{0xa1, 0x30}, decoded: "0"},
		{encoded: []byte{0xa2, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xbf}, fillerChars(31)...), decoded: "0123456789012345678901234567890"},
		// - str 8:
		{encoded: []byte{0xd9, 0x00}, decoded: ""},
		{encoded: []byte{0xd9, 0x01, 0x30}, decoded: "0"},
		{encoded: []byte{0xd9, 0x02, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xd9, 0xff}, fillerChars(255)...), decoded: string(fillerChars(255))},
		// - str 16:
		{encoded: []byte{0xda, 0x00, 0x00}, decoded: ""},
		{encoded: []byte{0xda, 0x00, 0x01, 0x30}, decoded: "0"},
		{encoded: []byte{0xda, 0x00, 0x02, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xda, 0xff, 0xff}, fillerChars(65535)...), decoded: string(fillerChars(65535))},
		// - str 32:
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x00}, decoded: ""},
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x01, 0x30}, decoded: "0"},
		{encoded: []byte{0xdb, 0x00, 0x00, 0x00, 0x02, 0x30, 0x31}, decoded: "01"},
		{encoded: append([]byte{0xdb, 0x00, 0x01, 0x86, 0xa0}, fillerChars(100000)...), decoded: string(fillerChars(100000))},
		// binary:
		// - bin 8:
		{encoded: []byte{0xc4, 0x00}, decoded: []byte{}},
		{encoded: []byte{0xc4, 0x01, 0x00}, decoded: []byte{0}},
		{encoded: []byte{0xc4, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
		{encoded: append([]byte{0xc4, 0xff}, fillerBytes(255)...), decoded: fillerBytes(255)},
		// - bin 16:
		{encoded: []byte{0xc5, 0x00, 0x00}, decoded: []byte{}},
		{encoded: []byte{0xc5, 0x00, 0x01, 0x00}, decoded: []byte{0}},
		{encoded: []byte{0xc5, 0x00, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
		{encoded: append([]byte{0xc5, 0xff, 0xff}, fillerBytes(65535)...), decoded: fillerBytes(65535)},
		// - bin 32:
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x00}, decoded: []byte{}},
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x01, 0x00}, decoded: []byte{0}},
		{encoded: []byte{0xc6, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}, decoded: []byte{0, 1}},
		{encoded: append([]byte{0xc6, 0x00, 0x01, 0x86, 0xa0}, fillerBytes(100000)...), decoded: fillerBytes(100000)},
		// TODO: array, map, ext, timestamp
	}
	for _, testCase := range testCases {
		buf := bytes.NewBuffer(testCase.encoded)
		if actualDecoded, actualErr := Unmarshal(opts, buf); actualErr != testCase.err {
			t.Errorf("unexected error for encoded=%q (decoded=%#v, err=%v): actualErr=%v", testCase.encoded, testCase.decoded, testCase.err, actualErr)
		} else if testCase.err == nil && !reflect.DeepEqual(actualDecoded, testCase.decoded) {
			t.Errorf("unexected result for encoded=%q (decoded=%#v): actualDecoded=%#v", testCase.encoded, testCase.decoded, actualDecoded)
		}
	}
}

/*
	switch {
	case b <= 0x8f: // fixmap: 1000xxxx: 0x80 - 0x8f
		return u.unmarshalNMap(uint(b & 0b1111))
	case b <= 0x9f: // fixarray: 1001xxxx: 0x90 - 0x9f
		return u.unmarshalNArray(uint(b & 0b1111))
	}

	switch b {
	case 0xc7: // ext 8: 11000111: 0xc7
		n, err := u.unmarshalUint8()
		if err != nil {
			return nil, err
		}
		return u.unmarshalNExt(n)
	case 0xc8: // ext 16: 11001000: 0xc8
		n, err := u.unmarshalUint16()
		if err != nil {
			return nil, err
		}
		return u.unmarshalNExt(n)
	case 0xc9: // ext 32: 11001001: 0xc9
		n, err := u.unmarshalUint32()
		if err != nil {
			return nil, err
		}
		return u.unmarshalNExt(n)
	case 0xd4: // fixext 1: 11010100: 0xd4
		return u.unmarshalNExt(1)
	case 0xd5: // fixext 2: 11010101: 0xd5
		return u.unmarshalNExt(2)
	case 0xd6: // fixext 4: 11010110: 0xd6
		return u.unmarshalNExt(4)
	case 0xd7: // fixext 8: 11010111: 0xd7
		return u.unmarshalNExt(8)
	case 0xd8: // fixext 16: 11011000: 0xd8
		return u.unmarshalNExt(16)
	case 0xdc: // array 16: 11011100: 0xdc
		n, err := u.unmarshalUint16()
		if err != nil {
			return nil, err
		}
		return u.unmarshalNArray(n)
	case 0xdd: // array 32: 11011101: 0xdd
		n, err := u.unmarshalUint32()
		if err != nil {
			return nil, err
		}
		return u.unmarshalNArray(n)
	case 0xde: // map 16: 11011110: 0xde
		n, err := u.unmarshalUint16()
		if err != nil {
			return nil, err
		}
		return u.unmarshalNMap(n)
	case 0xdf: // map 32: 11011111: 0xdf
		n, err := u.unmarshalUint32()
		if err != nil {
			return nil, err
		}
		return u.unmarshalNMap(n)
	}
*/
