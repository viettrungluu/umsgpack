// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

// This file contains Marshal, etc.

package umsgpack

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"time"
)

// Errors ------------------------------------------------------------------------------------------

// UnsupportedTypeForMarshallingError is the error returned if Marshal encounters an object whose
// type is unsupported for marshalling.
var UnsupportedTypeForMarshallingError = errors.New("Unsupported type for marshalling")

// ObjectTooBigForMarshallingError is the error returned if Marshal encounters an object that's too
// big for marshalling (e.g., a string that's 2**32 bytes or longer).
var ObjectTooBigForMarshallingError = errors.New("Object too big for marshalling")

// Marshal -----------------------------------------------------------------------------------------

// DefaultMarshalOptions is the default options used by Marshal/MarshalToBytes if it is passed nil
// options.
var DefaultMarshalOptions = &MarshalOptions{}

// Marshal marshals a single object as MessagePack to w.
//
// It marshals:
//   - nil to nil
//   - bool to true/false
//   - signed integer types (int, int{8,16,32,64}) to the most compact signed int format
//     (positive/negative fixint, int {8,16,32,64}) possible for the given value; note that it never
//     marshals a signed integer type to a MessagePack uint format, even though MessagePack's type
//     system permits this
//   - unsigned integer types (uint, uint{8,16,32,64}, uintptr) to the most compact uint format
//     (uint {8,16,32,64}) possible; note that it never marshals an unsigned integer to a
//     MessagePack int or fixint format
//   - float32 to float 32
//   - float64 to float 64; note that it will never marshals a float64 to a MessagePack float 32,
//     even when the representation would be exact
//   - string to the most compact str format (fixstr, str {8,16,32}) possible
//   - []byte to the most compact bin format (bin {8,16,32}) possible
//   - []any to the most compact array format (fixarray, array {16,32}) possible
//   - map[any]any to the most compact map format (fixmap, map {16,32}) possible
//   - *UnresolvedExtensionType to the most compact extension format (fixext {1,2,4,8,16}, ext
//     {8,16,32}) possible
//   - types transformed by the standard transformer to the above (unless
//     opts.DisableStandardTransformer is set); currently, this just effectively marshals time.Time
//     to the timestamp extension (type -1), using the most compact format possible (timestamp
//     {32,64,96}, as fixext {4,8}/ext 8, respectively)
//   - types transformed by the application transformer (opts.ApplicationTransformer) to the above
func Marshal(opts *MarshalOptions, w io.Writer, obj any) error {
	if opts == nil {
		opts = DefaultMarshalOptions
	}
	m := &marshaller{opts: opts, w: w}
	return m.marshalObject(obj)
}

// MarshalToBytes is like Marshal, except that it returns byte data instead of using an io.Writer.
func MarshalToBytes(opts *MarshalOptions, obj any) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := Marshal(opts, buf, obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalOptions specifies options for Marshal.
type MarshalOptions struct {
	// If set, then the standard transformer is not run.
	DisableStandardTransformer bool

	// ApplicationTransformer is a transformer run on the object before marshalling. This
	// transformer is run before the standard transformer.
	ApplicationTransformer TransformerFn
}

// Marshaller --------------------------------------------------------------------------------------

// A marshaller handles MessagePack marshalling for Marshal.
type marshaller struct {
	opts *MarshalOptions
	w    io.Writer
}

// marshalObject marshals an object.
func (m *marshaller) marshalObject(obj any) error {
	if m.opts.ApplicationTransformer != nil {
		var err error
		obj, err = m.opts.ApplicationTransformer(obj)
		if err != nil {
			return err
		}
	}

	if !m.opts.DisableStandardTransformer {
		var err error
		obj, err = StandardMarshalTransformer(obj)
		if err != nil {
			return err
		}
	}

	if obj == nil {
		return m.marshalNil()
	}

	switch v := obj.(type) {
	case bool:
		return m.marshalBool(v)
	case int:
		return m.marshalInt64(int64(v))
	case int8:
		return m.marshalInt64(int64(v))
	case int16:
		return m.marshalInt64(int64(v))
	case int32:
		return m.marshalInt64(int64(v))
	case int64:
		return m.marshalInt64(v)
	case uint:
		return m.marshalUint64(uint64(v))
	case uint8:
		return m.marshalUint64(uint64(v))
	case uint16:
		return m.marshalUint64(uint64(v))
	case uint32:
		return m.marshalUint64(uint64(v))
	case uint64:
		return m.marshalUint64(v)
	case uintptr:
		return m.marshalUint64(uint64(v))
	case float32:
		return m.marshalFloat32(v)
	case float64:
		return m.marshalFloat64(v)
	case string:
		return m.marshalString(v)
	case []byte:
		return m.marshalBytes(v)
	case []any:
		return m.marshalArray(v)
	case map[any]any:
		return m.marshalMap(v)
	case *UnresolvedExtensionType:
		return m.marshalExtensionType(int(v.ExtensionType), v.Data)
	}

	switch reflect.TypeOf(obj).Kind() {
	case reflect.Array, reflect.Slice:
		return m.marshalGenericArrayOrSlice(obj)
	case reflect.Map:
		return m.marshalGenericMap(obj)
	}

	return UnsupportedTypeForMarshallingError
}

// marshalNil marshals a nil.
func (m *marshaller) marshalNil() error {
	return m.write(0xc0) // nil: 11000000: 0xc0
}

// marshalBool marshals a bool.
func (m *marshaller) marshalBool(b bool) error {
	if b { // true: 11000011: 0xc3
		return m.write(0xc3)
	} else { // false: 11000010: 0xc2
		return m.write(0xc2)
	}
}

// marshalInt64 marshals an int64 (in a minimal way, though never as a MessagePack uint type).
func (m *marshaller) marshalInt64(i int64) error {
	switch {
	case i >= 0 && i <= 0x7f: // positive fixint: 0xxxxxxx: 0x00 - 0x7f
		return m.write(byte(i & 0xff))
	case i >= -(0x100-0xe0) && i < 0: // negative fixint: 111xxxxx: 0xe0 - 0xff
		return m.write(byte(i & 0xff))
	case i >= math.MinInt8 && i <= math.MaxInt8: // int 8: 11010000: 0xd0
		return m.write(0xd0, byte(i&0xff))
	case i >= math.MinInt16 && i <= math.MaxInt16: // int 16: 11010001: 0xd1
		return m.write(0xd1, byte((i>>8)&0xff), byte(i&0xff))
	case i >= math.MinInt32 && i <= math.MaxInt32: // int 32: 11010010: 0xd2
		return m.write(0xd2, byte((i>>24)&0xff), byte((i>>16)&0xff), byte((i>>8)&0xff), byte(i&0xff))
	default: // int 64: 11010011: 0xd3
		return m.write(0xd3, byte((i>>56)&0xff), byte((i>>48)&0xff), byte((i>>40)&0xff), byte((i>>32)&0xff), byte((i>>24)&0xff), byte((i>>16)&0xff), byte((i>>8)&0xff), byte(i&0xff))
	}
}

// marshalUint64 marshals a uint64 (in a minimal way, though only as a MessagePack uint type and
// never as a fixint).
func (m *marshaller) marshalUint64(u uint64) error {
	switch {
	case u <= math.MaxUint8: // uint 8: 11001100: 0xcc
		return m.write(0xcc, byte(u&0xff))
	case u <= math.MaxUint16: // uint 16: 11001101: 0xcd
		return m.write(0xcd, byte((u>>8)&0xff), byte(u&0xff))
	case u <= math.MaxUint32: // uint 32: 11001110: 0xce
		return m.write(0xce, byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff))
	default: // uint 64: 11001111: 0xcf
		return m.write(0xcf, byte((u>>56)&0xff), byte((u>>48)&0xff), byte((u>>40)&0xff), byte((u>>32)&0xff), byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff))
	}
}

// marshalFloat32 marshals a float32.
func (m *marshaller) marshalFloat32(f float32) error {
	u := math.Float32bits(f)
	// float 32: 11001010: 0xca
	return m.write(0xca, byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff))
}

// marshalFloat64 marshals a float64.
func (m *marshaller) marshalFloat64(f float64) error {
	u := math.Float64bits(f)
	// float 64: 11001011: 0xcb
	return m.write(0xcb, byte((u>>56)&0xff), byte((u>>48)&0xff), byte((u>>40)&0xff), byte((u>>32)&0xff), byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff))
}

// marshalString marshals a string (in a minimal way).
func (m *marshaller) marshalString(s string) error {
	u := len(s)
	switch {
	case u <= (0xbf - 0xa0): // fixstr: 101xxxxx: 0xa0 - 0xbf
		if err := m.write(byte(0xa0 + u)); err != nil {
			return err
		}
	case u <= math.MaxUint8: // str 8: 11011001: 0xd9
		if err := m.write(0xd9, byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint16: // str 16: 11011010: 0xda
		if err := m.write(0xda, byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint32: // str 32: 11011011: 0xdb
		if err := m.write(0xdb, byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	default:
		return ObjectTooBigForMarshallingError
	}
	return m.write([]byte(s)...)
}

// marshalBytes marshals a []byte (in a minimal way).
func (m *marshaller) marshalBytes(b []byte) error {
	u := len(b)
	switch {
	case u <= math.MaxUint8: // bin 8: 11000100: 0xc4
		if err := m.write(0xc4, byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint16: // bin 16: 11000101: 0xc5
		if err := m.write(0xc5, byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint32: // bin 32: 11000110: 0xc6
		if err := m.write(0xc6, byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	default:
		return ObjectTooBigForMarshallingError
	}
	return m.write(b...)
}

// marshalArray marshals a []any (in a minimal way).
func (m *marshaller) marshalArray(a []any) error {
	if err := m.writeArrayPrefix(len(a)); err != nil {
		return err
	}
	for _, v := range a {
		if err := m.marshalObject(v); err != nil {
			return err
		}
	}
	return nil
}

// marshalGenericArrayOrSlice marshals a generic array or slice (i.e., not just []any).
func (m *marshaller) marshalGenericArrayOrSlice(obj any) error {
	v := reflect.ValueOf(obj)
	u := v.Len()
	if err := m.writeArrayPrefix(u); err != nil {
		return err
	}
	for i := 0; i < u; i += 1 {
		if err := m.marshalObject(v.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

// writeArrayPrefix writes the prefix for an array of length u.
func (m *marshaller) writeArrayPrefix(u int) error {
	switch {
	case u <= (0x9f - 0x90): // fixarray: 1001xxxx: 0x90 - 0x9f
		if err := m.write(byte(0x90 + u)); err != nil {
			return err
		}
	case u <= math.MaxUint16: // array 16: 11011100: 0xdc
		if err := m.write(0xdc, byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint32: // array 32: 11011101: 0xdd
		if err := m.write(0xdd, byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	default:
		return ObjectTooBigForMarshallingError
	}
	return nil
}

// marshalMap marshals a map[any]any (in a minimal way).
func (m *marshaller) marshalMap(kvs map[any]any) error {
	if err := m.writeMapPrefix(len(kvs)); err != nil {
		return err
	}
	for k, v := range kvs {
		if err := m.marshalObject(k); err != nil {
			return err
		}
		if err := m.marshalObject(v); err != nil {
			return err
		}
	}
	return nil
}

// marshalGenericMap marshals a generic map (i.e., not just map[any]any).
func (m *marshaller) marshalGenericMap(obj any) error {
	v := reflect.ValueOf(obj)
	if err := m.writeMapPrefix(v.Len()); err != nil {
		return err
	}
	for it := v.MapRange(); it.Next(); {
		if err := m.marshalObject(it.Key().Interface()); err != nil {
			return err
		}
		if err := m.marshalObject(it.Value().Interface()); err != nil {
			return err
		}
	}
	return nil
}

// writeMapPrefix writes the prefix for a map of length u.
func (m *marshaller) writeMapPrefix(u int) error {
	switch {
	case u <= (0x8f - 0x80): // fixmap: 1000xxxx: 0x80 - 0x8f
		if err := m.write(byte(0x80 + u)); err != nil {
			return err
		}
	case u <= math.MaxUint16: // map 16: 11011110: 0xde
		if err := m.write(0xde, byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint32: // map 32: 11011111: 0xdf
		if err := m.write(0xdf, byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	default:
		return ObjectTooBigForMarshallingError
	}
	return nil
}

// marshalExtensionType marshals an extension type (in a minimal way).
func (m *marshaller) marshalExtensionType(extType int, extData []byte) error {
	u := len(extData)
	switch {
	case u == 1: // fixext 1: 11010100: 0xd4
		if err := m.write(0xd4); err != nil {
			return err
		}
	case u == 2: // fixext 2: 11010101: 0xd5
		if err := m.write(0xd5); err != nil {
			return err
		}
	case u == 4: // fixext 4: 11010110: 0xd6
		if err := m.write(0xd6); err != nil {
			return err
		}
	case u == 8: // fixext 8: 11010111: 0xd7
		if err := m.write(0xd7); err != nil {
			return err
		}
	case u == 16: // fixext 16: 11011000: 0xd8
		if err := m.write(0xd8); err != nil {
			return err
		}
	case u <= math.MaxUint8: // ext 8: 11000111: 0xc7
		if err := m.write(0xc7, byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint16: // ext 16: 11001000: 0xc8
		if err := m.write(0xc8, byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	case u <= math.MaxUint32: // ext 32: 11001001: 0xc9
		if err := m.write(0xc9, byte((u>>24)&0xff), byte((u>>16)&0xff), byte((u>>8)&0xff), byte(u&0xff)); err != nil {
			return err
		}
	default:
		return ObjectTooBigForMarshallingError
	}
	if err := m.write(byte(extType)); err != nil {
		return err
	}
	return m.write(extData...)
}

// write is a helper for calling the io.Writer's Write.
func (m *marshaller) write(data ...byte) error {
	_, err := m.w.Write(data)
	return err
}

// Standard marshal transformers -------------------------------------------------------------------

// StandardMarshalTransformer is the standard transformer run by Marshal (after the
var StandardMarshalTransformer = TimestampExtensionMarshalTransformer

// TimestampExtensionMarshalTransformer supports the standard (-1) timestamp extension type by
// transforming time.Time to a minimal *UnresolvedExtensionType.
func TimestampExtensionMarshalTransformer(obj any) (any, error) {
	t, ok := obj.(time.Time)
	if !ok {
		return obj, nil
	}

	sec := t.Unix()
	nsec := t.Nanosecond()
	var data []byte
	if sec >= 0 {
		if nsec == 0 && sec <= math.MaxUint32 {
			// timestamp 32
			data = []byte{byte((sec >> 24) & 0xff), byte((sec >> 16) & 0xff), byte((sec >> 8) & 0xff), byte(sec & 0xff)}
		} else if sec < (1 << 34) {
			// timestamp 64
			u := uint64(sec) | (uint64(nsec) << 34)
			data = []byte{byte((u >> 56) & 0xff), byte((u >> 48) & 0xff), byte((u >> 40) & 0xff), byte((u >> 32) & 0xff), byte((u >> 24) & 0xff), byte((u >> 16) & 0xff), byte((u >> 8) & 0xff), byte(u & 0xff)}
		}
	}

	// timestamp 96
	if data == nil {
		data = []byte{byte((nsec >> 24) & 0xff), byte((nsec >> 16) & 0xff), byte((nsec >> 8) & 0xff), byte(nsec & 0xff), byte((sec >> 56) & 0xff), byte((sec >> 48) & 0xff), byte((sec >> 40) & 0xff), byte((sec >> 32) & 0xff), byte((sec >> 24) & 0xff), byte((sec >> 16) & 0xff), byte((sec >> 8) & 0xff), byte(sec & 0xff)}
	}

	return &UnresolvedExtensionType{ExtensionType: -1, Data: data}, nil
}
