package zip

import (
	"errors"
	"io"
)

type nopWriteCloser struct {
	w io.Writer
}

func (w *nopWriteCloser) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *nopWriteCloser) Close() error {
	return nil
}

type nopReadCloser struct {
	r io.Reader
}

func (r *nopReadCloser) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *nopReadCloser) Close() error {
	return nil
}

type CountWriter struct {
	w     io.Writer
	Count int
}

func (w *CountWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.Count += n
	return n, err
}

type MultiWriter struct {
	ws []io.Writer
}

func (mw *MultiWriter) Write(p []byte) (int, error) {
	wsize := len(p)
	for _, w := range mw.ws {
		n, err := w.Write(p)
		if err != nil {
			return n, err
		}
		if n != wsize {
			return n, errors.New("write small data")
		}
	}
	return wsize, nil
}
