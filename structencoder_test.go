// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests structencoder.go.

package umsgpack_test

import (
	"reflect"
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

func TestDefaultStructMarshalTransformer(t *testing.T) {
	successTestCases := []struct {
		obj      any
		expected any
	}{
		{123, 123},
		{[]int{123, 45}, []int{123, 45}},
		{struct{}{}, map[string]any{}},
		{struct {
			Foo string
			Bar int
			baz int
		}{"hello", 123, 0}, map[string]any{"Foo": "hello", "Bar": 123}},
	}
	for i, tc := range successTestCases {
		if result, err := DefaultStructMarshalTransformer(tc.obj); err != nil {
			t.Errorf("%v: unexpected error: %v", i, err)
		} else if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("%v: unexpected result: %v (expected: %v)", i, result, tc.expected)
		}
	}
}

// TODO: more tests.
