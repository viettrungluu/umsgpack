// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains a simple MarshalTransformerFn for marshalling structs.

package umsgpack

import (
	"reflect"
)

// StructMarshalTransformerOptions are options for MakeStructMarshalTransformer.
type StructMarshalTransformerOptions struct {
	// FieldFn "handles" a field: it decides whether it should be included and if so the map key
	// to use. If nil, the default is to include all (expored) fields and use the field name
	// (field.Name) verbatim as the key.
	FieldFn func(field reflect.StructField) (includeField bool, mapKey string)
}

// MakeStructMarshalTransformer makes a MarshalTransformerFn for transforming structs to a
// marshallable map[string]any.
func MakeStructMarshalTransformer(opts *StructMarshalTransformerOptions) MarshalTransformerFn {
	if opts == nil {
		opts = &StructMarshalTransformerOptions{}
	}

	fieldFn := opts.FieldFn
	if fieldFn == nil {
		fieldFn = func(field reflect.StructField) (bool, string) {
			return true, field.Name
		}
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
			if !field.IsExported() {
				continue
			}

			includeField, key := fieldFn(field)
			if !includeField {
				continue
			}

			value := v.FieldByIndex(field.Index).Interface()
			rv[key] = value
		}

		return rv, nil
	}
}

var DefaultStructMarshalTransformer = MakeStructMarshalTransformer(nil)
