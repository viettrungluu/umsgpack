// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains TransformerFn, etc.

package umsgpack

// A TransformerFn transforms an object, and is used in different ways for unmarshalling and
// marshalling.
//
// For unmarshalling, it typically transforms an *UnresolvedExtensionType to some other type. (E.g.,
// for the timestamp extension, it transforms *UnresolvedExtensionType with ExtensionType -1 to
// time.Time.) [This is TODO.]
//
// For marshalling, it typically transforms some unsupported (e.g., nonstandard or not built-in)
// type to a marshallable type. (E.g., for the timestamp extension, it transforms time.Time to
// *UnresolvedExtensionType with ExtensionType -1.)
//
// If the transformer does not apply, it should just return the object as-is and no error.
//
// If it applies, it should return the transformed object, but may also return an error if there is
// some fatal problem.
//
// It may determine applicability however it wants (e.g., based on type, on reflection, or on
// nothing at all).
type TransformerFn func(obj any) (any, error)

// ComposeTransformers produces a single transformer from the given transformers (executing them in
// argument order).
func ComposeTransformers(xforms ...TransformerFn) TransformerFn {
	return func(obj any) (any, error) {
		for _, xform := range xforms {
			var err error
			obj, err = xform(obj)
			if err != nil {
				return nil, err
			}
		}
		return obj, nil
	}
}
