// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains UnresolvedExtensionType, used by both Unmarshal and Marshal.

package umsgpack

// An *UnresolvedExtensionType represents data from an unresolved/unsupported extension type.
type UnresolvedExtensionType struct {
	ExtensionType int8
	Data          []byte
}
