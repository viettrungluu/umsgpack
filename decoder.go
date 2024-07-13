// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains Unmarshal, etc.

package umsgpack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"
)

// Errors ------------------------------------------------------------------------------------------

// DuplicateKeyError is the error returned if Unmarshal encounters data for a map with duplicate
// keys.
//
// This may be suppressed by setting the DisableDuplicateKeyError option.
var DuplicateKeyError = errors.New("Duplicate key")

// UnsupportedKeyTypeError is the error returned if Unmarshal encounters data for a map with a key
// of unsupported type (for a map[any]any key).
//
// This may be suppressed by setting the DisableUnsupportedKeyTypeError option.
var UnsupportedKeyTypeError = errors.New("Unsupported key type")

// InvalidFormatError is the error returned if Unmarshal encounters an invalid format (0xc1).
var InvalidFormatError = errors.New("Invalid format")

// Unmarshal ---------------------------------------------------------------------------------------

// DefaultUnmarshalOptions is the default options used by Unmarshal/UnmarshalBytes if it is passed
// nil options.
var DefaultUnmarshalOptions = &UnmarshalOptions{}

// Unmarshal unmarshals a single MessagePack object from r. It is very simplistic, and produces the
// following types:
//   - nil for nil
//   - bool for true or false
//   - int for any integer serialized as signed
//   - uint for any integer serialized as unsigned
//   - float32 and float64 for 32- and 64-bit floats, respectively
//   - string for (UTF-8) string
//   - []byte for binary
//   - []any for array
//   - map[any]any for map
//   - time.Time for timestamp (extension type -1)
//   - other types per opts.ApplicationUnmarshalExtensions
func Unmarshal(opts *UnmarshalOptions, r io.Reader) (any, error) {
	if opts == nil {
		opts = DefaultUnmarshalOptions
	}
	u := &unmarshaller{opts: opts, r: r}
	rv, _, err := u.unmarshalObject()
	return rv, err
}

// UnmarshalBytes is like Unmarshal, except taking byte data instead of an io.Reader.
func UnmarshalBytes(opts *UnmarshalOptions, data []byte) (any, error) {
	return Unmarshal(opts, bytes.NewBuffer(data))
}

// UnmarshalOptions specifies options for Unmarshal.
type UnmarshalOptions struct {
	// If DisableDuplicateKeyError is set, then DuplicateKeyErrors will not be returned.
	// Instead, the *first* key-value pair for the given key will "win".
	//
	// The default is to return such errors since inconsistencies across unmarshallers could
	// lead to security problems.
	DisableDuplicateKeyError bool

	// If DisableUnsupportedKeyTypeError is set, then UnsupportedKeyTypeErrors will not be
	// returned. Instead, the key-value pair will be "dropped".
	//
	// The default is to return such errors since inconsistencies across unmarshallers could
	// lead to security problems.
	DisableUnsupportedKeyTypeError bool

	// If set, then the standard unmarshal transformer will not be run.
	DisableStandardUnmarshalTransformer bool

	// ApplicationUnmarshalTransformer is a marshal transformer run on objects after
	// unmarshalling (and after the standard unmarshal transformer).
	// This is run before the standard marshal transformer.
	ApplicationUnmarshalTransformer UnmarshalTransformerFn
}

// An UnmarshalTransformerFn transforms an object after unmarshalling.
//
// It typically transforms *UnresolvedExtensionType to some more standard/concrete type. (E.g., to
// support the timestamp extension, TimestampExtensionUnmarshalTransformer transforms
// *UnresolvedExtensionType with ExtensionType -1 to time.Time.)
//
// If the transformer does not apply, it should just return its arguments as-is and no error. If it
// applies, it should return the transformed object and whether the transformed object may be a map
// key, but may also return an error if there is some fatal problem. It may determine applicability
// however it wants (e.g., based on type, on reflection, or on nothing at all).
type UnmarshalTransformerFn func(obj any, mapKeySupported bool) (any, bool, error)

// unmarshaller ------------------------------------------------------------------------------------

// An unmarshaller handles MessagePack unmarshalling for Unmarshal.
type unmarshaller struct {
	opts *UnmarshalOptions
	r    io.Reader
}

// Internal configuration:
// TODO: We could make these options.
const (
	// unmarshalMaxArrayAllocElements is the maximum initial array allocation size in number of
	// elements) when unmarshalling an array. For arrays larger than this, the slice will be
	// grown as needed.
	//
	// (This is less efficient, but prevents bad data from causing huge allocations.)
	unmarshalMaxArrayAllocElements = 1000

	// unmarshalReadChunkSize is the maximum single read size when unmarshalling.
	unmarshalReadChunkSize = 4096
)

// unmarshalObject unmarshals an object. The next byte is expected to be the format.
//
// Note: All internal unmarshal functions are like an UnmarshalExtensionTypeFn and return either an
// error, or on success the object and a boolean indicating if the value is a valid map key (for a
// map[any]any).
func (u *unmarshaller) unmarshalObject() (obj any, mapKeySupported bool, err error) {
	obj, mapKeySupported, err = u.unmarshalStandardObject()
	if err != nil {
		return
	}

	if !u.opts.DisableStandardUnmarshalTransformer {
		obj, mapKeySupported, err = StandardUnmarshalTransformer(obj, mapKeySupported)
		if err != nil {
			return
		}
	}

	if u.opts.ApplicationUnmarshalTransformer != nil {
		obj, mapKeySupported, err = u.opts.ApplicationUnmarshalTransformer(obj, mapKeySupported)
	}

	return
}

// unmarshalStandardObject unmarshals an object to a standard (built-in) object (i.e., without
// applying transformers).
func (u *unmarshaller) unmarshalStandardObject() (any, bool, error) {
	b, err := u.readByte()
	if err != nil {
		return nil, false, err
	}

	switch {
	case b <= 0x7f: // positive fixint: 0xxxxxxx: 0x00 - 0x7f
		return int(b), true, nil
	case b <= 0x8f: // fixmap: 1000xxxx: 0x80 - 0x8f
		return u.unmarshalNMap(uint(b & 0b1111))
	case b <= 0x9f: // fixarray: 1001xxxx: 0x90 - 0x9f
		return u.unmarshalNArray(uint(b & 0b1111))
	case b <= 0xbf: // fixstr: 101xxxxx: 0xa0 - 0xbf
		return u.unmarshalNString(uint(b & 0b11111))
	// Reaches individual range (handled below), until:
	case b >= 0xe0: // negative fixint: 111xxxxx: 0xe0 - 0xff
		// Cast to an int8 first, so that casting to an int will sign-extend.
		return int(int8(b)), true, nil
	}

	switch b {
	case 0xc0: // nil: 11000000: 0xc0
		return nil, true, nil
	case 0xc1: // (never used): 11000001: 0xc1
		return nil, false, InvalidFormatError
	case 0xc2: // false: 11000010: 0xc2
		return false, true, nil
	case 0xc3: // true: 11000011: 0xc3
		return true, true, nil
	case 0xc4: // bin 8: 11000100: 0xc4
		n, _, err := u.unmarshalUint8()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNBytes(n)
	case 0xc5: // bin 16: 11000101: 0xc5
		n, _, err := u.unmarshalUint16()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNBytes(n)
	case 0xc6: // bin 32: 11000110: 0xc6
		n, _, err := u.unmarshalUint32()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNBytes(n)
	case 0xc7: // ext 8: 11000111: 0xc7
		n, _, err := u.unmarshalUint8()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNExt(n)
	case 0xc8: // ext 16: 11001000: 0xc8
		n, _, err := u.unmarshalUint16()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNExt(n)
	case 0xc9: // ext 32: 11001001: 0xc9
		n, _, err := u.unmarshalUint32()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNExt(n)
	case 0xca: // float 32: 11001010: 0xca
		return u.unmarshalFloat32()
	case 0xcb: // float 64: 11001011: 0xcb
		return u.unmarshalFloat64()
	case 0xcc: // uint 8: 11001100: 0xcc
		return u.unmarshalUint8()
	case 0xcd: // uint 16: 11001101: 0xcd
		return u.unmarshalUint16()
	case 0xce: // uint 32: 11001110: 0xce
		return u.unmarshalUint32()
	case 0xcf: // uint 64: 11001111: 0xcf
		return u.unmarshalUint64()
	case 0xd0: // int 8: 11010000: 0xd0
		return u.unmarshalInt8()
	case 0xd1: // int 16: 11010001: 0xd1
		return u.unmarshalInt16()
	case 0xd2: // int 32: 11010010: 0xd2
		return u.unmarshalInt32()
	case 0xd3: // int 64: 11010011: 0xd3
		return u.unmarshalInt64()
	case 0xd4: // fixext 1: 11010100: 0xd4
		return u.unmarshalNExt(1)
	case 0xd5: // fixext 2: 11010101: 0xd5
		return u.unmarshalNExt(2)
	case 0xd6: // fixext 4: 11010110: 0xd6
		return u.unmarshalNExt(4)
	case 0xd7: // fixext 8: 11010111: 0xd7
		return u.unmarshalNExt(8)
	case 0xd8: // fixext 16: 11011000: 0xd8
		return u.unmarshalNExt(16)
	case 0xd9: // str 8: 11011001: 0xd9
		n, _, err := u.unmarshalUint8()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNString(n)
	case 0xda: // str 16: 11011010: 0xda
		n, _, err := u.unmarshalUint16()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNString(n)
	case 0xdb: // str 32: 11011011: 0xdb
		n, _, err := u.unmarshalUint32()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNString(n)
	case 0xdc: // array 16: 11011100: 0xdc
		n, _, err := u.unmarshalUint16()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNArray(n)
	case 0xdd: // array 32: 11011101: 0xdd
		n, _, err := u.unmarshalUint32()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNArray(n)
	case 0xde: // map 16: 11011110: 0xde
		n, _, err := u.unmarshalUint16()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNMap(n)
	case 0xdf: // map 32: 11011111: 0xdf
		n, _, err := u.unmarshalUint32()
		if err != nil {
			return nil, false, err
		}
		return u.unmarshalNMap(n)
	}

	panic("Should be unreachable!")
}

// unmarshalUint8 unmarshals a uint 8 (as a uint).
func (u *unmarshaller) unmarshalUint8() (uint, bool, error) {
	if b, err := u.readByte(); err != nil {
		return 0, false, err
	} else {
		return uint(b), true, err
	}
}

// unmarshalUint16 unmarshals a uint 16 (as a uint).
func (u *unmarshaller) unmarshalUint16() (uint, bool, error) {
	if data, err := u.readBytes(2); err != nil {
		return 0, false, err
	} else {
		return uint(binary.BigEndian.Uint16(data)), true, err
	}
}

// unmarshalUint32 unmarshals a uint 32 (as a uint).
func (u *unmarshaller) unmarshalUint32() (uint, bool, error) {
	if data, err := u.readBytes(4); err != nil {
		return 0, false, err
	} else {
		return uint(binary.BigEndian.Uint32(data)), true, err
	}
}

// unmarshalUint64 unmarshals a uint 64 (as a uint).
func (u *unmarshaller) unmarshalUint64() (uint, bool, error) {
	if data, err := u.readBytes(8); err != nil {
		return 0, false, err
	} else {
		return uint(binary.BigEndian.Uint64(data)), true, err
	}
}

// unmarshalInt8 unmarshals an int 8 (as an int).
func (u *unmarshaller) unmarshalInt8() (int, bool, error) {
	if b, err := u.readByte(); err != nil {
		return 0, false, err
	} else {
		// Cast to an int8 first, so that casting to an int will sign-extend.
		return int(int8(b)), true, err
	}
}

// unmarshalInt16 unmarshals an int 16 (as an int).
func (u *unmarshaller) unmarshalInt16() (int, bool, error) {
	if data, err := u.readBytes(2); err != nil {
		return 0, false, err
	} else {
		// Cast to an int16 first, so that casting to an int will sign-extend.
		return int(int16(binary.BigEndian.Uint16(data))), true, err
	}
}

// unmarshalInt32 unmarshals an int 32 (as an int).
func (u *unmarshaller) unmarshalInt32() (int, bool, error) {
	if data, err := u.readBytes(4); err != nil {
		return 0, false, err
	} else {
		// Cast to an int32 first, so that casting to an int will sign-extend.
		return int(int32(binary.BigEndian.Uint32(data))), true, err
	}
}

// unmarshalInt64 unmarshals an int 64 (as an int).
func (u *unmarshaller) unmarshalInt64() (int, bool, error) {
	if data, err := u.readBytes(8); err != nil {
		return 0, false, err
	} else {
		// Cast to an int64 first, so that casting to an int will sign-extend.
		return int(int64(binary.BigEndian.Uint64(data))), true, err
	}
}

// unmarshalFloat32 unmarshals a float 32 (as a float32).
func (u *unmarshaller) unmarshalFloat32() (float32, bool, error) {
	if data, err := u.readBytes(4); err != nil {
		return 0, false, err
	} else {
		return math.Float32frombits(binary.BigEndian.Uint32(data)), true, err
	}
}

// unmarshalFloat64 unmarshals a float 64 (as a float64).
func (u *unmarshaller) unmarshalFloat64() (float64, bool, error) {
	if data, err := u.readBytes(8); err != nil {
		return 0, false, err
	} else {
		return math.Float64frombits(binary.BigEndian.Uint64(data)), true, err
	}
}

// unmarshalNMap unmarshals a map with n entries.
func (u *unmarshaller) unmarshalNMap(n uint) (map[any]any, bool, error) {
	rv := map[any]any{}
	for i := uint(0); i < n; i += 1 {
		// Always try to unmarshal both the key and value even if we're going to return a
		// higher-level error (duplicate key or unsupported key type) -- because if we
		// ignore the error, then we need to "advance" our position properly.
		key, mapKeySupported, err := u.unmarshalObject()
		if err != nil {
			return nil, false, err
		}

		value, _, err := u.unmarshalObject()
		if err != nil {
			return nil, false, err
		}

		if !mapKeySupported {
			if !u.opts.DisableUnsupportedKeyTypeError {
				return nil, false, UnsupportedKeyTypeError
			}
			// Else ignore this key-value pair.
		} else if _, alreadyPresent := rv[key]; alreadyPresent {
			if !u.opts.DisableDuplicateKeyError {
				return nil, false, DuplicateKeyError
			}
			// Else let the first key-value pair with the same key win.
		} else {
			rv[key] = value
		}
	}
	return rv, false, nil
}

// unmarshalNArray unmarshals an array with n entries.
func (u *unmarshaller) unmarshalNArray(n uint) ([]any, bool, error) {
	rv := make([]any, 0, min(n, unmarshalMaxArrayAllocElements))
	for i := uint(0); i < n; i += 1 {
		element, _, err := u.unmarshalObject()
		if err != nil {
			return nil, false, err
		}
		rv = append(rv, element)
	}
	return rv, false, nil
}

// unmarshalNString unmarshals a string of length n (bytes).
// Note that it does not validate that it is valid UTF-8.
// TODO: Should it be an option?
func (u *unmarshaller) unmarshalNString(n uint) (string, bool, error) {
	if data, err := u.readBytesChunked(n); err != nil {
		return "", false, err
	} else {
		return string(data), true, nil
	}
}

// unmarshalNBytes unmarshals a byte array of length n (bytes).
func (u *unmarshaller) unmarshalNBytes(n uint) ([]byte, bool, error) {
	if data, err := u.readBytesChunked(n); err != nil {
		return nil, false, err
	} else {
		return data, false, nil
	}
}

// unmarshalNExt unmarshals an extension with data of length n (bytes).
func (u *unmarshaller) unmarshalNExt(n uint) (any, bool, error) {
	if extensionType, _, err := u.unmarshalInt8(); err != nil {
		return nil, false, err
	} else {
		if data, err := u.readBytesChunked(n); err != nil {
			return nil, false, err
		} else {
			return &UnresolvedExtensionType{ExtensionType: int8(extensionType), Data: data}, false, nil
		}
	}
}

// readByte is a helper that reads exactly one byte.
func (u *unmarshaller) readByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(u.r, buf)
	return buf[0], err
}

// readBytes is a helper that reads exactly n bytes, all at once. It should only be used to read a
// controlled number of bytes (e.g., a fixed number, or a bounded number).
func (u *unmarshaller) readBytes(n uint) ([]byte, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(u.r, data); err != nil {
		return nil, err
	}
	return data, nil
}

// readBytesChunked is like readBytes, but it doesn't try to allocate all n bytes immediately.
func (u *unmarshaller) readBytesChunked(n uint) ([]byte, error) {
	// Fast path:
	if n <= unmarshalReadChunkSize {
		return u.readBytes(n)
	}

	var rv []byte
	for n > 0 {
		m := min(n, unmarshalReadChunkSize)
		// TODO: grow rv and read straight into it.
		if data, err := u.readBytes(m); err != nil {
			return nil, err
		} else {
			rv = append(rv, data...)
		}
		n -= m
	}
	return rv, nil
}

// Unmarshal transformers --------------------------------------------------------------------------

// TODO: compose unmarshal transformers?

// An UnmarshalExtensionTypeFn unmarshals the given data for a (fixed, known) extension type. It may
// keep a reference to the data (i.e., it need not copy it).
//
// It either returns an error, or on success the object and a boolean indicating if the value is a
// valid map key (for a map[any]any).
type UnmarshalExtensionTypeFn func(data []byte) (obj any, mapKeySupported bool, err error)

// MakeExtensionTypeUnmarshalTransformer makes an unmarshal transformer for the given extensions,
// specified as a map from an extension type (-128 to 127) to an UnmarshalExtensionTypeFn.
func MakeExtensionTypeUnmarshalTransformer(unmarshalExtensions map[int8]UnmarshalExtensionTypeFn) UnmarshalTransformerFn {
	return func(obj any, mapKeySupported bool) (any, bool, error) {
		ext, ok := obj.(*UnresolvedExtensionType)
		if !ok {
			return obj, mapKeySupported, nil
		}

		// Shouldn't normally happen, but could happen if other transformers play tricks.
		if ext == nil {
			return obj, mapKeySupported, nil
		}

		unmarshalFn, ok := unmarshalExtensions[ext.ExtensionType]
		if !ok {
			return obj, mapKeySupported, nil
		}

		return unmarshalFn(ext.Data)
	}
}

// StandardUnmarshalTransformer is the standard unmarshal transformer run by Unmarshal (before the
// application unmarshal transformer, if any).
var StandardUnmarshalTransformer UnmarshalTransformerFn = MakeExtensionTypeUnmarshalTransformer(
	map[int8]UnmarshalExtensionTypeFn{
		-1: UnmarshalTimestampExtensionType,
	},
)

// InvalidTimestampError is the error returned by TimestampExtensionUnmarshalTransformer for an
// invalid timestamp.
var InvalidTimestampError = errors.New("Invalid timestamp")

// UnmarshalTimestampExtensionType is an UnmarshalExtensionTypeFn that unmarshals the standard (-1)
// timestamp extension type.
func UnmarshalTimestampExtensionType(data []byte) (any, bool, error) {
	switch len(data) {
	case 4: // timestamp 32
		sec := int64(binary.BigEndian.Uint32(data))
		return time.Unix(sec, 0), true, nil
	case 8: // timestamp 64
		data64 := binary.BigEndian.Uint64(data)
		nsec := int64(data64 >> 34)
		sec := int64(data64 & 0x00000003ffffffff)
		if nsec >= 1_000_000_000 {
			return nil, false, InvalidTimestampError
		}
		return time.Unix(sec, nsec), true, nil
	case 12: // timestamp 96
		nsec := int64(binary.BigEndian.Uint32(data[0:4]))
		sec := int64(binary.BigEndian.Uint64(data[4:12]))
		if nsec >= 1_000_000_000 {
			return nil, false, InvalidTimestampError
		}
		return time.Unix(sec, nsec), true, nil
	default:
		return nil, false, InvalidTimestampError
	}
}
