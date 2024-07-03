// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests encoder.go.

package umsgpack_test

import (
	"bytes"
	// "io"
	// "math"
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
}

// TestMarshal_defaultOpts tests Marshal with the default options (all boolean options are false).
func TestMarshal_defaultOpts(t *testing.T) {
	opts := &MarshalOptions{}
	testMarshal(t, opts, commonMarshalTestCases)
	// TODO: testMarshal(t, opts, defaultOptsMarshalTestCases)
}
