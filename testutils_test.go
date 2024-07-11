// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains test utils that are used in multiple test files.

package umsgpack_test

import (
	"strconv"
)

// fillerChars generates n filler characters in the pattern 012345678901234....
func fillerChars(n int) []byte {
	rv := make([]byte, n)
	for i := 0; i < n; i += 1 {
		rv[i] = byte('0' + i%10)
	}
	return rv
}

// fillerBytes generates n filler bytes in the pattern 0, 1, 2, ..., 255, 0, 1, ....
func fillerBytes(n int) []byte {
	rv := make([]byte, n)
	for i := 0; i < n; i += 1 {
		rv[i] = byte(i % 256)
	}
	return rv
}

// genArrayData generates test array (encoded) data with n entries; matches genArray.
func genArrayData(n int) []byte {
	rv := []byte{}
	for i := 0; i < n; i += 1 {
		s := strconv.Itoa(i)
		rv = append(rv, byte(0xa0+len(s)))
		rv = append(rv, []byte(s)...)
	}
	return rv
}

// genArray generates a test array with n entries; matches genArrayData.
func genArray(n int) []any {
	rv := []any{}
	for i := 0; i < n; i += 1 {
		rv = append(rv, strconv.Itoa(i))
	}
	return rv
}

// genTypedArray generates a strongly-typed test array (slice) with n entries; matches genArrayData.
func genTypedArray(n int) []string {
	rv := []string{}
	for i := 0; i < n; i += 1 {
		rv = append(rv, strconv.Itoa(i))
	}
	return rv
}

// genMapData generates test map (encoded) data with n key-value pairs; matches genMap.
func genMapData(n int) []byte {
	rv := []byte{}
	for i := 0; i < n; i += 1 {
		s := strconv.Itoa(i)
		rv = append(rv, byte(0xa0+len(s)))
		rv = append(rv, []byte(s)...)
		j := i % 10000
		if j <= 0x7f {
			rv = append(rv, byte(j)) // positive fixint
		} else {
			rv = append(rv, 0xd1, byte(j>>8), byte(j)) // int 16
		}
	}
	return rv
}

// genMap generates test map with n key-value pairs; matches genMapData.
func genMap(n int) map[any]any {
	rv := map[any]any{}
	for i := 0; i < n; i += 1 {
		rv[strconv.Itoa(i)] = i % 10000
	}
	return rv
}

// genTypedMap generates strongly-typed test map with n key-value pairs; matches genMapData.
func genTypedMap(n int) map[string]int {
	rv := map[string]int{}
	for i := 0; i < n; i += 1 {
		rv[strconv.Itoa(i)] = i % 10000
	}
	return rv
}
