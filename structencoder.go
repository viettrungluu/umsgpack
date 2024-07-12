// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains a simple MarshalTransformerFn for marshalling structs.

package umsgpack

import (
	"reflect"
)

// StructMarshalTransformerOptions are options for MakeStructMarshalTransformer.
type StructMarshalTransformerOptions struct {
	// FilterFn filters fields: it should return true if the field should be included. If nil,
	// all (exported) fields will be included.
	FilterFn func(field reflect.StructField) bool

	// KeyNameFn transforms the field name to the map key name. If nil, the field name will be
	// used verbatim.
	KeyNameFn func(s string) string
}

// MakeStructMarshalTransformer makes a MarshalTransformerFn for transforming structs to a
// marshallable map[string]any.
func MakeStructMarshalTransformer(opts *StructMarshalTransformerOptions) MarshalTransformerFn {
	if opts == nil {
		opts = &StructMarshalTransformerOptions{}
	}

	filterFn := opts.FilterFn
	if filterFn == nil {
		filterFn = func(field reflect.StructField) bool { return true }
	}

	keyNameFn := opts.KeyNameFn
	if keyNameFn == nil {
		keyNameFn = func(s string) string { return s }
	}

	return func(obj any) (any, error) {
		t := reflect.TypeOf(obj)
		if t.Kind() != reflect.Struct {
			return obj, nil
		}

		fields := reflect.VisibleFields(t)
		v := reflect.ValueOf(obj)
		rv := map[string]any{}
		for _, field := range fields {
			if !field.IsExported() || !filterFn(field) {
				continue
			}

			key := keyNameFn(field.Name)
			value := v.FieldByIndex(field.Index).Interface()
			rv[key] = value
		}

		return rv, nil
	}
}

var DefaultStructMarshalTransformer = MakeStructMarshalTransformer(nil)
