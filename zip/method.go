package zip

import (
	"compress/flate"
	"io"
)

func init() {
	initMethods()
}

var (
	compressors   map[uint16]Compressor
	decompressors map[uint16]Decompressor
)

type Compressor func(io.Writer) io.WriteCloser
type Decompressor func(io.Reader) io.ReadCloser

func initMethods() {
	compressors = make(map[uint16]Compressor)
	decompressors = make(map[uint16]Decompressor)

	// Method 0 : Store
	compressors[0] = newStoreWriter
	decompressors[0] = newStoreReader
	// Method 8 : Flate
	compressors[8] = newFlateWriter
	decompressors[8] = newDeflateReader
}

func newStoreWriter(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

func newStoreReader(r io.Reader) io.ReadCloser {
	return &nopReadCloser{r}
}

func newFlateWriter(w io.Writer) io.WriteCloser {
	fw, err := flate.NewWriter(w, 5)
	if err != nil {
		panic(err)
	}
	return fw
}

func newDeflateReader(r io.Reader) io.ReadCloser {
	return flate.NewReader(r)
}

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
