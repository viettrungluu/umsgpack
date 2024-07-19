// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

package internal_test

import (
	"bytes"
	"io"
	"testing"

	. "github.com/viettrungluu/umsgpack/internal"
)

func makeTestBuf(n int) []byte {
	rv := make([]byte, n)
	for i := 0; i < n; i += 1 {
		rv[i] = byte(i % 256)
	}
	return rv
}

func TestReadViewerForReader_ReadByte(t *testing.T) {
	reader := bytes.NewBuffer([]byte("12"))
	r := ReadViewerForReader{reader}

	if b, err := r.ReadByte(); err != nil || b != '1' {
		t.Errorf("Unexpected result: %v, %v", b, err)
	}
	if b, err := r.ReadByte(); err != nil || b != '2' {
		t.Errorf("Unexpected result: %v, %v", b, err)
	}
	if b, err := r.ReadByte(); err != io.EOF {
		t.Errorf("Unexpected result: %v, %v", b, err)
	}
}

func TestReadViewerForReader_ReadView(t *testing.T) {
	{
		data := []byte("123456")
		reader := bytes.NewBuffer(data)
		r := ReadViewerForReader{reader}

		if buf, err := r.ReadView(0); err != nil {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadView(2); err != nil || bytes.Compare(buf, []byte("12")) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadView(3); err != nil || bytes.Compare(buf, []byte("345")) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadView(2); err != io.ErrUnexpectedEOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadView(1); err != io.EOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
	}

	{
		data := makeTestBuf(ReaderChunkSize)
		reader := bytes.NewBuffer(data)
		r := ReadViewerForReader{reader}

		if buf, err := r.ReadView(ReaderChunkSize); err != nil || bytes.Compare(buf, data) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadView(ReaderChunkSize); err != io.EOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
	}

	{
		data := makeTestBuf(3 * ReaderChunkSize)
		reader := bytes.NewBuffer(data)
		r := ReadViewerForReader{reader}

		if buf, err := r.ReadView(2 * ReaderChunkSize); err != nil || bytes.Compare(buf, data[:2*ReaderChunkSize]) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadView(ReaderChunkSize + 1); err != io.ErrUnexpectedEOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadView(ReaderChunkSize + 1); err != io.EOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
	}
}

func TestReadViewerForReader_ReadCopy(t *testing.T) {
	{
		data := []byte("123456")
		reader := bytes.NewBuffer(data)
		r := ReadViewerForReader{reader}

		if buf, err := r.ReadCopy(0); err != nil {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadCopy(2); err != nil || bytes.Compare(buf, []byte("12")) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		} else {
			// Mutate buffer, and make sure source doesn't change.
			buf[0] = 'x'
			if data[0] != '1' {
				t.Errorf("mutating buffer mutated source data")
			}
		}
		if buf, err := r.ReadCopy(3); err != nil || bytes.Compare(buf, []byte("345")) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadCopy(2); err != io.ErrUnexpectedEOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadCopy(1); err != io.EOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
	}

	{
		data := makeTestBuf(ReaderChunkSize)
		reader := bytes.NewBuffer(data)
		r := ReadViewerForReader{reader}

		if buf, err := r.ReadCopy(ReaderChunkSize); err != nil || bytes.Compare(buf, data) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadCopy(ReaderChunkSize); err != io.EOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
	}

	{
		data := makeTestBuf(3 * ReaderChunkSize)
		reader := bytes.NewBuffer(data)
		r := ReadViewerForReader{reader}

		if buf, err := r.ReadCopy(2 * ReaderChunkSize); err != nil || bytes.Compare(buf, data[:2*ReaderChunkSize]) != 0 {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadCopy(ReaderChunkSize + 1); err != io.ErrUnexpectedEOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
		if buf, err := r.ReadCopy(ReaderChunkSize + 1); err != io.EOF {
			t.Errorf("Unexpected result: %v, %v", buf, err)
		}
	}
}

func TestReadViewerForBuffer_ReadByte(t *testing.T) {
	r := &ReadViewerForBuffer{Buffer: []byte("12")}

	if b, err := r.ReadByte(); err != nil || b != '1' {
		t.Errorf("Unexpected result: %v, %v", b, err)
	}
	if b, err := r.ReadByte(); err != nil || b != '2' {
		t.Errorf("Unexpected result: %v, %v", b, err)
	}
	if b, err := r.ReadByte(); err != io.EOF {
		t.Errorf("Unexpected result: %v, %v", b, err)
	}
}

func TestReadViewerForBuffer_ReadView(t *testing.T) {
	data := []byte("123456")
	r := &ReadViewerForBuffer{Buffer: data}

	if buf, err := r.ReadView(0); err != nil {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
	if buf, err := r.ReadView(2); err != nil || bytes.Compare(buf, []byte("12")) != 0 {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
	if buf, err := r.ReadView(3); err != nil || bytes.Compare(buf, []byte("345")) != 0 {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
	if buf, err := r.ReadView(2); err != io.ErrUnexpectedEOF {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
	if buf, err := r.ReadView(1); err != io.EOF {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
}

func TestReadViewerForBuffer_ReadCopy(t *testing.T) {
	data := []byte("123456")
	r := &ReadViewerForBuffer{Buffer: data}

	if buf, err := r.ReadCopy(0); err != nil {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
	if buf, err := r.ReadCopy(2); err != nil || bytes.Compare(buf, []byte("12")) != 0 {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	} else {
		// Mutate buffer, and make sure source doesn't change.
		buf[0] = 'x'
		if data[0] != '1' {
			t.Errorf("mutating buffer mutated source data")
		}
	}
	if buf, err := r.ReadCopy(3); err != nil || bytes.Compare(buf, []byte("345")) != 0 {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
	if buf, err := r.ReadCopy(2); err != io.ErrUnexpectedEOF {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
	if buf, err := r.ReadCopy(1); err != io.EOF {
		t.Errorf("Unexpected result: %v, %v", buf, err)
	}
}
