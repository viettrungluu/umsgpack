// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// umsgpack is a tiny (micro), simple implementation of MessagePack.
//
// Unlike other Go implementations of MessagePack, it more closely adheres to MessagePack's weak
// type system.
//
// # Usage
//
// Unmarshalling is as simple as:
//
//	output, err := umsgpack.UnmarshalBytes(nil, input)
//
// Or if you have an io.Reader:
//
//	output, err := umsgpack.Unmarshal(nil, reader)
//
// Both take optional options, which can for example add support for extension types:
//
//	// Unmarshals a time.Duration represented as a 64-bit value, big-endian.
//	unmarshalDuration := func(data []byte) (any, bool, error) {
//		if len(data) != 8 {
//			return nil, false, errors.New("invalid")
//		} else {
//			return time.Duration(binary.BigEndian.Uint64(data)), true, nil
//		}
//	}
//	opts := &umsgpack.UnmarshalOptions{
//		ApplicationUnmarshalTransformer: umsgpack.MakeExtensionTypeUnmarshalTransformer(
//			map[int8]umsgpack.UnmarshalExtensionTypeFn{
//				42: unmarshalDuration, // Extension type 42.
//			},
//		),
//	}
//	output, err := umsgpack.UnmarshalBytes(opts, input)
//
// (One would typically prebuild opts, perhaps as a global variable.)
//
// Marshalling is just as simple:
//
//	output, err := umsgpack.MarshalToBytes(nil, input)
//
// Or if you have an io.Writer:
//
//	err := umsgpack.Marshal(nil, writer, input)
//
// To support extensions:
//
//	// Marshals a time.Duration to a extension type 42, containing a 64-bit value, big-endian.
//	marshalDuration := func(obj any) (any, error) {
//		if duration, ok := obj.(time.Duration); !ok {
//			return obj, nil
//		} else {
//			data := make([]byte, 8)
//			binary.BigEndian.PutUint64(data, uint64(duration))
//			return &umsgpack.UnresolvedExtensionType{ExtensionType: 42, Data: data}, nil
//		}
//	}
//	opts := &umsgpack.MarshalOptions{
//		ApplicationMarshalTransformer: marshalDuration,
//	}
//	output, err := umsgpack.MarshalToBytes(opts, input)
package umsgpack

import (
	"strconv"
)

func init() {
	if strconv.IntSize < 64 {
		panic("umsgpack requires at least 64-bit int!")
	}
}
