// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file tests structencoder.go.

package umsgpack_test

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

func TestDefaultStructMarshalTransformer(t *testing.T) {
	testCases := []struct {
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
	for i, tc := range testCases {
		if result, err := DefaultStructMarshalTransformer(tc.obj); err != nil {
			t.Errorf("%v: unexpected error: %v", i, err)
		} else if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("%v: unexpected result: %v (expected: %v)", i, result, tc.expected)
		}
	}
}

func TestMakeStructMarshalTransformer(t *testing.T) {
	opts := &StructMarshalTransformerOptions{
		FieldFn: func(field reflect.StructField) (bool, string) {
			// Only include fields with even name lengths.
			if len(field.Name)%2 == 0 {
				return true, strings.ToUpper(field.Name)
			} else {
				return false, ""
			}
		},
	}
	transformer := MakeStructMarshalTransformer(opts)

	testCases := []struct {
		obj      any
		expected any
	}{
		{123, 123},
		{[]int{123, 45}, []int{123, 45}},
		{struct{}{}, map[string]any{}},
		{struct {
			Hi     string
			World  int
			Frob   bool
			secret int
		}{"world", 123, true, 456}, map[string]any{"HI": "world", "FROB": true}},
	}
	for i, tc := range testCases {
		if result, err := transformer(tc.obj); err != nil {
			t.Errorf("%v: unexpected error: %v", i, err)
		} else if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("%v: unexpected result: %v (expected: %v)", i, result, tc.expected)
		}
	}
}
