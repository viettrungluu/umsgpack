// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

package umsgpack

import (
	"bytes"
	// "encoding/binary"
	"errors"
	"io"
	// "math"
	// "time"
)

// Errors ------------------------------------------------------------------------------------------

// TODO

// Marshal -----------------------------------------------------------------------------------------

var DefaultMarshalOptions = &MarshalOptions{}

// Marshal marshals a single object as MessagePack to w.
//
// TODO: more details.
func Marshal(opts *MarshalOptions, w io.Writer, obj any) error {
	if opts == nil {
		opts = DefaultMarshalOptions
	}
	return nil
	m := &marshaller{opts: opts, w: w}
	return m.marshalObject(obj)
}

// MarshalToBytes is like Marshal, except that it returns byte data instead of using an io.Writer.
func MarshalToBytes(opts *MarshalOptions, obj any) ([]byte, error) {
	w := &bytes.Buffer{}
	if err := Marshal(opts, w, obj); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
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
	// TODO
	return errors.New("Not yet implemented!")
}
