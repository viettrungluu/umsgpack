// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests decoder.go.

package umsgpack_test

import (
	"bytes"
	"reflect"
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

func TestUnmarshal(t *testing.T) {
	opts := UnmarshalOptions{}
	testCases := []struct{
		encoded []byte
		decoded any
	}{
		// Positive fixint:
		{ encoded: []byte{0x00}, decoded: int(0) },
		{ encoded: []byte{0x01}, decoded: int(1) },
		{ encoded: []byte{0x02}, decoded: int(2) },
		{ encoded: []byte{0x7f}, decoded: int(127) },
	}
	for _, testCase := range testCases {
		buf := bytes.NewBuffer(testCase.encoded)
		if actualDecoded, err := Unmarshal(opts, buf); err != nil || !reflect.DeepEqual(actualDecoded, testCase.decoded) {
			t.Errorf("unexected result for encoded=%q (decoded=%#v): err=%v, actualDecoded=%#v", testCase.encoded, testCase.decoded, err, actualDecoded)
		}
	}
}
