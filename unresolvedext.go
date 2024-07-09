// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

package umsgpack

// An *UnresolvedExtensionType represents data from an unresolved/unsupported extension type.
type UnresolvedExtensionType struct {
	ExtensionType int8
	Data          []byte
}
