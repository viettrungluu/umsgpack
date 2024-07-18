// Copyright 2024 Viet-Trung Luu.
// Use of this source code is governed by the license in the LICENSE file.

package umsgpack

import (
	"io"
)

// ReadViewer --------------------------------------------------------------------------------------

// A ReadViewer is similar in spirit to an io.Reader, except that its methods may return temporary
// views into a buffer.
//
// All its methods should return io.EOF if no bytes are available and io.ErrUnexpectedEOF if some
// but not all bytes are available. In either case, the data returned is not meaningful.
type ReadViewer interface {
	// ReadByte reads exactly one byte, returning it by value.
	ReadByte() (byte, error)

	// ReadView reads exactly n bytes, returning a slice "view" that is valid (at least) until
	// the next operation. (It may do exactly the same thing as ReadCopy, however.)
	ReadView(n uint) ([]byte, error)

	// ReadCopy reads exactly n bytes, returning a slice that the caller may take ownership of
	// (i.e., is valid "forever").
	ReadCopy(n uint) ([]byte, error)
}

// ReadViewerForReader -----------------------------------------------------------------------------

// Internal configuration:
const (
	// readerChunkSize is the maximum single read size from an io.Reader (for a
	// ReadViewerForReader).
	readerChunkSize = 4096
)

// A ReadViewerForReader is a ReadViewer that wraps an io.Reader. (Note that these are typically
// passed by value.)
type ReadViewerForReader struct {
	Reader io.Reader
}

var _ ReadViewer = ReadViewerForReader{}

// ReadByte implements ReadViewer.ReadByte.
func (r ReadViewerForReader) ReadByte() (byte, error) {
	data := make([]byte, 1)
	_, err := io.ReadFull(r.Reader, data)
	return data[0], err
}

// ReadView implements ReadViewer.ReadView.
func (r ReadViewerForReader) ReadView(n uint) ([]byte, error) {
	return r.ReadCopy(n)
}

// ReadCopy implements ReadViewer.ReadCopy.
func (r ReadViewerForReader) ReadCopy(n uint) ([]byte, error) {
	// Fast path:
	if n <= readerChunkSize {
		return r.readCopyAll(n)
	}

	var data []byte
	for n > 0 {
		m := min(n, readerChunkSize)
		// TODO: grow data and read straight into it.
		if chunk, err := r.readCopyAll(m); err != nil {
			if err == io.EOF && len(data) > 0 {
				// Return ErrUnexpectedEOF instead of EOF if we've read any data.
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		} else {
			data = append(data, chunk...)
		}
		n -= m
	}
	return data, nil
}

// readCopyAll is a helper for ReadCopy that reads the data all at once.
func (r ReadViewerForReader) readCopyAll(n uint) ([]byte, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(r.Reader, data); err != nil {
		return nil, err
	}
	return data, nil
}

// ReadViewerForBuffer -----------------------------------------------------------------------------

// A ReadViewerForBuffer is a ReadViewer that wraps a byte slice.
type ReadViewerForBuffer struct {
	Buffer []byte

	pos uint
}

var _ ReadViewer = (*ReadViewerForBuffer)(nil)

// ReadByte implements ReadViewer.ReadByte.
func (r *ReadViewerForBuffer) ReadByte() (byte, error) {
	len := uint(len(r.Buffer))
	if r.pos >= len {
		return 0, io.EOF
	}

	b := r.Buffer[r.pos]
	r.pos += 1
	return b, nil
}

// ReadView implements ReadViewer.ReadView.
func (r *ReadViewerForBuffer) ReadView(n uint) ([]byte, error) {
	if n == 0 {
		return nil, nil
	}
	len := uint(len(r.Buffer))
	if r.pos >= len {
		return nil, io.EOF
	}
	if len-r.pos < n {
		return nil, io.ErrUnexpectedEOF
	}

	rv := r.Buffer[r.pos : r.pos+n]
	r.pos += n
	return rv, nil
}

// ReadCopy implements ReadViewer.ReadCopy.
func (r *ReadViewerForBuffer) ReadCopy(n uint) ([]byte, error) {
	if view, err := r.ReadView(n); err != nil {
		return nil, err
	} else {
		rv := make([]byte, n)
		copy(rv, view)
		return rv, nil
	}
}
