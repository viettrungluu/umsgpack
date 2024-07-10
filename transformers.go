// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains standard transformers (for use with Marshal).

package umsgpack

import (
	"reflect"
)

// MarshalArrayTransformer -------------------------------------------------------------------------

// MarshalArrayTransformer is a transformer (MarshalObjectTransformerFn) that transforms any array
// or slice to a []any.
func MarshalArrayTransformer(obj any) (any, error) {
	if kind := reflect.TypeOf(obj).Kind(); kind != reflect.Array && kind != reflect.Slice {
		return obj, nil
	}

	v := reflect.ValueOf(obj)
	vlen := v.Len()
	rv := make([]any, 0, vlen)
	for i := 0; i < vlen; i += 1 {
		rv = append(rv, v.Index(i).Interface())
	}
	return rv, nil
}

var _ MarshalObjectTransformerFn = MarshalArrayTransformer

// MarshalMapTransformer ---------------------------------------------------------------------------

// MarshalMapTransformer is a transformer (MarshalObjectTransformerFn) that transforms any map to a
// map[any]any.
func MarshalMapTransformer(obj any) (any, error) {
	if kind := reflect.TypeOf(obj).Kind(); kind != reflect.Map {
		return obj, nil
	}

	v := reflect.ValueOf(obj)
	rv := map[any]any{}
	for iter := v.MapRange(); iter.Next(); {
		rv[iter.Key().Interface()] = iter.Value().Interface()
	}
	return rv, nil
}
