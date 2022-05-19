package zip

import (
	"io"
)

// nopWriteCloser implements io.WriteCloser by wrapping io.Writer.
type nopWriteCloser struct {
	w io.Writer
}

// Write implements the standard Write interface.
func (w *nopWriteCloser) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

// Close does nothing.
func (w *nopWriteCloser) Close() error {
	return nil
}

// nopReadCloser implements io.ReadCloser by wrapping io.Reader.
type nopReadCloser struct {
	r io.Reader
}

// Read implements the standard Read interface.
func (r *nopReadCloser) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

// Close does nothing.
func (r *nopReadCloser) Close() error {
	return nil
}

// CountWriter implements io.Writer and counts the size of the written data.
type CountWriter struct {
	w     io.Writer
	Count int
}

// Write implements the standard Write interface.
func (w *CountWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.Count += n
	return n, err
}
