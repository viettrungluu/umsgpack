// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests arrayxform.go.

package umsgpack_test

import (
	"bytes"
	"reflect"
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

func testMarshalArrayTransformer(t *testing.T, name string, obj any, expectedErr error, expectedOut any) {
	if actualOut, actualErr := MarshalArrayTransformer(obj); actualErr != expectedErr {
		t.Errorf("%v: incorrect error: actual=%v, expected=%v", name, actualErr, expectedErr)
	} else if expectedErr == nil && !reflect.DeepEqual(actualOut, expectedOut) {
		t.Errorf("%v: incorrect result: actual=%v, expected=%v", name, actualOut, expectedOut)
	}
}

func TestMarshalArrayTransformer_notApplicable(t *testing.T) {
	testMarshalArrayTransformer(t, "int", int(123), nil, int(123))
	testMarshalArrayTransformer(t, "string", "hi", nil, "hi")
	testMarshalArrayTransformer(t, "struct", struct{}{}, nil, struct{}{})
}

func TestMarshalArrayTransformer_array(t *testing.T) {
	testMarshalArrayTransformer(t, "int array", [3]int{1, 2, 3}, nil, []any{1, 2, 3})
	testMarshalArrayTransformer(t, "string array", [2]string{"hello", "world"}, nil, []any{"hello", "world"})
	testMarshalArrayTransformer(t, "empty array", [0]struct{}{}, nil, []any{})
}

func TestMarshalArrayTransformer_slice(t *testing.T) {
	testMarshalArrayTransformer(t, "int slice", []int{1, 2, 3}, nil, []any{1, 2, 3})
	testMarshalArrayTransformer(t, "string slice", []string{"hello", "world"}, nil, []any{"hello", "world"})
	testMarshalArrayTransformer(t, "empty slice", []struct{}{}, nil, []any{})
}

func TestMarshalArrayTransformer_MarshalToBytes(t *testing.T) {
	opts := &MarshalOptions{
		ApplicationMarshalObjectTransformers: []MarshalObjectTransformerFn{
			MarshalArrayTransformer,
		},
	}

	if encoded, err := MarshalToBytes(opts, []string{"0", "1", "2"}); err != nil || bytes.Compare(encoded, append([]byte{0x93}, genArrayData(3)...)) != 0 {
		t.Errorf("Unexpected result from MarshalToBytes: %v, %v", encoded, err)
	}
}

// TODO: add test of MarshalArrayTransformer + Marshal.
