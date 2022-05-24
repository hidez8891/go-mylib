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
type Compressor func(dst io.Writer, option uint16) io.WriteCloser

// Decompressor is an interface for implementing decompression schemes.
type Decompressor func(src io.Reader, option uint16) io.ReadCloser

// initMethods initializes a list of compression and decompression methods.
func initMethods() {
	compressors = make(map[uint16]Compressor)
	decompressors = make(map[uint16]Decompressor)

	// Method 0 : Store
	compressors[0] = newStoreWriter
	decompressors[0] = newStoreReader
	// Method 8 : Deflate
	compressors[8] = newDeflateWriter
	decompressors[8] = newDeflateReader
}

// newStoreWrite returns store (raw) compression writer.
func newStoreWriter(w io.Writer, _ uint16) io.WriteCloser {
	return &nopWriteCloser{w}
}

// newStoreWrite returns store (raw) decompression reader.
func newStoreReader(r io.Reader, _ uint16) io.ReadCloser {
	return &nopReadCloser{r}
}

// newStoreWrite returns deflate compression writer.
func newDeflateWriter(w io.Writer, option uint16) io.WriteCloser {
	var level int
	switch option & 0x03 {
	case 0x00:
		level = flate.DefaultCompression
	case 0x01:
		level = flate.BestCompression
	case 0x02:
		level = flate.BestSpeed
	case 0x03:
		level = flate.HuffmanOnly // Super Fast
	default:
		panic("unreachable")
	}

	fw, err := flate.NewWriter(w, level)
	if err != nil {
		panic(err)
	}
	return fw
}

// newStoreWrite returns deflate decompression writer.
func newDeflateReader(r io.Reader, _ uint16) io.ReadCloser {
	return flate.NewReader(r)
}
