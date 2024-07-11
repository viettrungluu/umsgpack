// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains standard transformers (for use with Marshal).

package umsgpack

import (
	"reflect"
)

// MarshalMapTransformer ---------------------------------------------------------------------------

// MarshalMapTransformer is a transformer (MarshalTransformerFn) that transforms any map to a
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

var _ MarshalTransformerFn = MarshalMapTransformer
