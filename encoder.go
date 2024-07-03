// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

package umsgpack

import (
	"bytes"
	"errors"
	"io"
	"math"
	// "time"
)

// Errors ------------------------------------------------------------------------------------------

// UnsupportedTypeForMarshallingError is the error returned if Marshal encounters an object whose
// type is unsupported for marshalling.
var UnsupportedTypeForMarshallingError = errors.New("Unsupported type for marshalling")

// ObjectTooBigForMarshallingError is the error returned if Marshal encounters an object that's too
// big for marshalling (e.g., a string that's 2**32 bytes or longer).
var ObjectTooBigForMarshallingError = errors.New("Object too big for marshalling")

// Marshal -----------------------------------------------------------------------------------------

var DefaultMarshalOptions = &MarshalOptions{}

// Marshal marshals a single object as MessagePack to w.
//
// TODO: more details.
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
	// TODO
}

// Marshaller --------------------------------------------------------------------------------------

// A marshaller handles MessagePack marshalling for Marshal.
type marshaller struct {
	opts *MarshalOptions
	w    io.Writer
}

// marshalObject marshals an object.
func (m *marshaller) marshalObject(obj any) error {
	// TODO: Support custom marshalling: via an interface or via a function (we want to support
	// both; an interface works well for types owned by the caller, while a function works for
	// "third-party" types).

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
		// case map[any]any:
		// TODO: other arrays and maps? and structs?
	}

	// TODO
	return errors.New("Not yet implemented!")
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
	u := len(a)
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
	for _, v := range a {
		if err := m.marshalObject(v); err != nil {
			return err
		}
	}
	return nil
}

// write is a helper for calling the io.Writer's Write.
func (m *marshaller) write(data ...byte) error {
	_, err := m.w.Write(data)
	return err
}
