package buffer

import (
	"errors"
	"io"
)

// Reader supports Buffer reading.
type Reader struct {
	buf *Buffer
	pos int
}

// NewReader returns buf's Reader.
func NewReader(buf *Buffer) *Reader {
	return &Reader{
		buf: buf,
		pos: 0,
	}
}

// Read reads []byte from Buffer.
// This function updates the read position.
func (r *Reader) Read(b []byte) (n int, err error) {
	if r.buf.Len() <= r.pos {
		return 0, io.EOF
	}

	cn := copy(b, r.buf.data[r.pos:])
	r.pos += cn
	return cn, nil
}

// ReadAt writes []byte from Buffer at offset off.
// This function does not update the read position.
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	if r.buf.Len() <= int(off) {
		return 0, io.EOF
	}

	cn := copy(b, r.buf.data[off:])
	return cn, nil
}

// Seek sets the next read position.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		offset += 0
	case 1:
		offset += int64(r.pos)
	case 2:
		offset += int64(r.buf.Len())
	default:
		return 0, errors.New("invalid whence")
	}
	if offset < 0 {
		return 0, errors.New("negative position")
	}

	if r.buf.Len() <= int(offset) {
		return 0, io.EOF
	}
	r.pos = int(offset)

	return int64(r.pos), nil
}
