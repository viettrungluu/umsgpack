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

type transformerTestCase struct {
	name        string
	obj         any
	expectedErr error
	expectedOut any
}

func testTransformer(t *testing.T, xform MarshalTransformerFn, tCs []transformerTestCase) {
	for _, tC := range tCs {
		if actualOut, actualErr := xform(tC.obj); actualErr != tC.expectedErr {
			t.Errorf("%v: incorrect error: actual=%v, expected=%v", tC.name, actualErr, tC.expectedErr)
		} else if tC.expectedErr == nil && !reflect.DeepEqual(actualOut, tC.expectedOut) {
			t.Errorf("%v: incorrect result: actual=%v, expected=%v", tC.name, actualOut, tC.expectedOut)
		}
	}
}

// MarshalMapTransformer ---------------------------------------------------------------------------

func TestMarshalMapTransformer_notApplicable(t *testing.T) {
	testTransformer(t, MarshalMapTransformer, []transformerTestCase{
		{"int", int(123), nil, int(123)},
		{"struct", struct{}{}, nil, struct{}{}},
		{"slice", []int{1, 2, 3}, nil, []int{1, 2, 3}},
	})
}

func TestMarshalMapTransformer_map(t *testing.T) {
	testTransformer(t, MarshalMapTransformer, []transformerTestCase{
		{"map[string]int", map[string]int{"hi": 123, "world": 456}, nil, map[any]any{"hi": 123, "world": 456}},
		{"map[string]any", map[string]any{"hi": 123}, nil, map[any]any{"hi": 123}},
		{"map[any]int", map[any]int{"hi": 123}, nil, map[any]any{"hi": 123}},
		{"map[string]struct{}", map[any]struct{}{"hi": struct{}{}}, nil, map[any]any{"hi": struct{}{}}},
	})
}

func TestMarshalMapTransformer_MarshalToBytes(t *testing.T) {
	opts := &MarshalOptions{
		LateMarshalTransformers: []MarshalTransformerFn{
			MarshalMapTransformer,
		},
	}

	if encoded, err := MarshalToBytes(opts, map[string]int{"0": 0}); err != nil || bytes.Compare(encoded, append([]byte{0x81}, genMapData(1)...)) != 0 {
		t.Errorf("Unexpected result from MarshalToBytes: %v, %v", encoded, err)
	}
}
