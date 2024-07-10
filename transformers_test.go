// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests transformers.go.

package umsgpack_test

import (
	"bytes"
	"reflect"
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

// MarshalArrayTransformer -------------------------------------------------------------------------

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
	testMarshalArrayTransformer(t, "map", map[string]int{"hi": 123}, nil, map[string]int{"hi": 123})
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

// MarshalMapTransformer ---------------------------------------------------------------------------

func testMarshalMapTransformer(t *testing.T, name string, obj any, expectedErr error, expectedOut any) {
	if actualOut, actualErr := MarshalMapTransformer(obj); actualErr != expectedErr {
		t.Errorf("%v: incorrect error: actual=%v, expected=%v", name, actualErr, expectedErr)
	} else if expectedErr == nil && !reflect.DeepEqual(actualOut, expectedOut) {
		t.Errorf("%v: incorrect result: actual=%v, expected=%v", name, actualOut, expectedOut)
	}
}

func TestMarshalMapTransformer_notApplicable(t *testing.T) {
	testMarshalMapTransformer(t, "int", int(123), nil, int(123))
	testMarshalMapTransformer(t, "struct", struct{}{}, nil, struct{}{})
	testMarshalMapTransformer(t, "slice", []int{1, 2, 3}, nil, []int{1, 2, 3})
}

func TestMarshalMapTransformer_map(t *testing.T) {
	testMarshalMapTransformer(t, "map[string]int", map[string]int{"hi": 123, "world": 456}, nil, map[any]any{"hi": 123, "world": 456})
	testMarshalMapTransformer(t, "map[string]any", map[string]any{"hi": 123}, nil, map[any]any{"hi": 123})
	testMarshalMapTransformer(t, "map[any]int", map[any]int{"hi": 123}, nil, map[any]any{"hi": 123})
	testMarshalMapTransformer(t, "map[string]struct{}", map[any]struct{}{"hi": struct{}{}}, nil, map[any]any{"hi": struct{}{}})
}

func TestMarshalMapTransformer_MarshalToBytes(t *testing.T) {
	opts := &MarshalOptions{
		ApplicationMarshalObjectTransformers: []MarshalObjectTransformerFn{
			MarshalMapTransformer,
		},
	}

	if encoded, err := MarshalToBytes(opts, map[string]int{"0": 0}); err != nil || bytes.Compare(encoded, append([]byte{0x81}, genMapData(1)...)) != 0 {
		t.Errorf("Unexpected result from MarshalToBytes: %v, %v", encoded, err)
	}
}
