package zip

import (
	"compress/flate"
	"errors"
	"fmt"
	"io"
)

// MethodType represents a compression method.
type MethodType interface {
	ID() uint16
	set(flag uint16) error
	get() uint16
	newCompressor(w io.Writer) (io.WriteCloser, error)
	newDecompressor(r io.Reader) (io.ReadCloser, error)
}

const (
	methodStoreID    uint16 = 0x00 // method ID for storing original file
	methodDeflatedID uint16 = 0x08 // method ID for deflating
)

// methodFactory returns a MethodType from a method ID.
func methodFactory(method uint16) (MethodType, error) {
	switch method {
	case methodStoreID:
		return &MethodStore{}, nil
	case methodDeflatedID:
		return &MethodDeflated{DefaultCompression}, nil
	}
	return nil, errors.New("unsupport compression method")
}

// CompressionType represents a compression level.
type CompressionType int

const (
	DefaultCompression   CompressionType = iota // default level compression
	MaximumCompression                          // maximum compression
	FastCompression                             // fast compression
	SuperFastCompression                        // super fast compression
)

// MethodStore is a compression method for storing data.
type MethodStore struct {
}

// ID returns a compression method's ID.
func (MethodStore) ID() uint16 {
	return methodStoreID
}

// set sets method options from a zip header's flags.
func (MethodStore) set(flag uint16) error {
	return nil
}

// get returns method options in zip header's flag format.
func (MethodStore) get() uint16 {
	return 0x00
}

// newCompressor returns a compressor.
func (m MethodStore) newCompressor(w io.Writer) (io.WriteCloser, error) {
	return &nopWriteCloser{w}, nil
}

// newDecompressor returns a decompressor.
func (m MethodStore) newDecompressor(r io.Reader) (io.ReadCloser, error) {
	return &nopReadCloser{r}, nil
}

// MethodDeflated is a compression method for deflate.
type MethodDeflated struct {
	Compression CompressionType
}

// ID returns a compression method's ID.
func (MethodDeflated) ID() uint16 {
	return methodDeflatedID
}

// set sets method options from a zip header's flags.
func (m *MethodDeflated) set(flag uint16) error {
	switch (flag >> 1) & 0x03 {
	case 0x00:
		m.Compression = DefaultCompression
	case 0x01:
		m.Compression = MaximumCompression
	case 0x02:
		m.Compression = FastCompression
	case 0x03:
		m.Compression = SuperFastCompression
	}
	return nil
}

// get returns method options in zip header's flag format.
func (m MethodDeflated) get() uint16 {
	switch m.Compression {
	case DefaultCompression:
		return 0x00 << 1
	case MaximumCompression:
		return 0x01 << 1
	case FastCompression:
		return 0x02 << 1
	case SuperFastCompression:
		return 0x03 << 1
	}
	return 0 // never reach
}

// newCompressor returns a compressor.
func (m MethodDeflated) newCompressor(w io.Writer) (io.WriteCloser, error) {
	var level int
	switch m.Compression {
	case DefaultCompression:
		level = flate.DefaultCompression
	case MaximumCompression:
		level = flate.BestCompression
	case FastCompression:
		level = flate.BestSpeed
	case SuperFastCompression:
		level = flate.HuffmanOnly // Super Fast
	default:
		return nil, fmt.Errorf("unsupport compression level: %#v", m.Compression)
	}

	return flate.NewWriter(w, level)
}

// newDecompressor returns a decompressor.
func (m MethodDeflated) newDecompressor(r io.Reader) (io.ReadCloser, error) {
	return flate.NewReader(r), nil
}
