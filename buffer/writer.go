package buffer

import "errors"

// Writer supports Buffer writing.
type Writer struct {
	buf *Buffer
	pos int
}

// NewWriter returns buf's Writer.
func NewWriter(buf *Buffer) *Writer {
	return &Writer{
		buf: buf,
		pos: 0,
	}
}

// Write writes []byte to Buffer.
// This function updates the write position.
func (w *Writer) Write(b []byte) (n int, err error) {
	size := w.pos + len(b)
	if w.buf.Len() < size {
		w.buf.grow(size - w.buf.Len())
	}

	cn := copy(w.buf.data[w.pos:], b)
	w.pos += cn
	return cn, nil
}

// WriteAt writes []byte to Buffer at offset off.
// This function does not update the write position.
func (w *Writer) WriteAt(b []byte, off int64) (n int, err error) {
	size := int(off) + len(b)
	if w.buf.Len() < size {
		w.buf.grow(size - w.buf.Len())
	}

	cn := copy(w.buf.data[off:], b)
	return cn, nil
}

// Seek sets the next write position.
func (w *Writer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		offset += 0
	case 1:
		offset += int64(w.pos)
	case 2:
		offset += int64(w.buf.Len())
	default:
		return 0, errors.New("invalid whence")
	}
	if offset < 0 {
		return 0, errors.New("negative position")
	}

	size := int(offset)
	if w.buf.Len() < size {
		w.buf.grow(size - w.buf.Len())
	}
	w.pos = int(offset)

	return int64(w.pos), nil
}
