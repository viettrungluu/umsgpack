// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file fuzz tests decoder.go.

package umsgpack_test

import (
	"testing"

	. "github.com/viettrungluu/umsgpack"
)

func FuzzUnmarshalBytes(f *testing.F) {
	for _, tCs := range [][]unmarshalTestCase{
		commonUnmarshalTestCases,
		timestampUnmarshalTestCases,
		defaultOptsUnmarshalTestCases,
		nonDefaultOptsUnmarshalTestCases,
	} {
		for _, tC := range tCs {
			// Skip really large test cases.
			if len(tC.encoded) > 5000 {
				continue
			}
			f.Add(tC.encoded)
		}
	}

	f.Fuzz(func(t *testing.T, encoded []byte) {
		// We just don't want it to panic.
		UnmarshalBytes(nil, encoded)
	})
}
