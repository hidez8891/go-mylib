package zip

import (
	"compress/flate"
	"io"
)

func init() {
	initMethods()
}

var (
	compressors   map[uint16]Compressor   // map of compression methods
	decompressors map[uint16]Decompressor // map of decompression methods
)

// Compressor is an interface for implementing compression schemes.
type Compressor func(io.Writer) io.WriteCloser

// Decompressor is an interface for implementing decompression schemes.
type Decompressor func(io.Reader) io.ReadCloser

// initMethods initializes a list of compression and decompression methods.
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

// newStoreWrite returns store (raw) compression writer.
func newStoreWriter(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

// newStoreWrite returns store (raw) decompression reader.
func newStoreReader(r io.Reader) io.ReadCloser {
	return &nopReadCloser{r}
}

// newStoreWrite returns deflate compression writer.
func newFlateWriter(w io.Writer) io.WriteCloser {
	fw, err := flate.NewWriter(w, 5)
	if err != nil {
		panic(err)
	}
	return fw
}

// newStoreWrite returns deflate decompression writer.
func newDeflateReader(r io.Reader) io.ReadCloser {
	return flate.NewReader(r)
}
