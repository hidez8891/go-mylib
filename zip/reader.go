package zip

import (
	"bytes"
	"errors"
	"io"
	"math"
)

type Reader struct {
	r io.ReadSeeker

	Files []*File
}

func NewReader(r io.ReadSeeker) (*Reader, error) {
	zr := &Reader{
		r: r,
	}

	if err := zr.init(); err != nil {
		return nil, err
	}
	return zr, nil
}

func (r *Reader) init() error {
	offset, err := findEndCentralDirectory(r.r)
	if err != nil {
		return err
	}
	if _, err := r.r.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	enddir := new(endCentralDirectory)
	if _, err := enddir.ReadFrom(r.r); err != nil {
		return err
	}

	r.Files = make([]*File, enddir.numberOfEntries)
	if _, err := r.r.Seek(int64(enddir.offsetCentralDirectory), io.SeekStart); err != nil {
		return err
	}
	for i := 0; i < int(enddir.numberOfEntries); i++ {
		cdir := new(centralDirectoryHeader)
		if _, err := cdir.ReadFrom(r.r); err != nil {
			return err
		}
		r.Files[i] = &File{
			r:    r.r,
			cdir: cdir,
		}
	}

	return nil
}

type File struct {
	r    io.ReadSeeker
	cdir *centralDirectoryHeader
}

func (f *File) Open() (io.ReadCloser, error) {
	r, err := f.OpenRaw()
	if err != nil {
		return nil, err
	}

	decomp, ok := decompressors[f.cdir.Method]
	if !ok {
		return nil, errors.New("Unsupport decompress method")
	}

	return decomp(r), nil
}

func (f *File) OpenRaw() (io.ReadCloser, error) {
	if _, err := f.r.Seek(int64(f.cdir.LocalHeaderOffset), io.SeekStart); err != nil {
		return nil, err
	}

	h := new(localFileHeader)
	if _, err := h.ReadFrom(f.r); err != nil {
		return nil, err
	}
	// simple name check
	if f.cdir.FileName != h.FileName {
		return nil, errors.New("broken zip: file name is different")
	}

	r := io.LimitReader(f.r, int64(f.cdir.CompressedSize))
	return &nopReadCloser{r}, nil
}

func findEndCentralDirectory(r io.ReadSeeker) (offset int64, err error) {
	// get size
	startOffset, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	endOffset, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	filesize := endOffset - startOffset

	// max comment size = max uint16
	size := int64(math.MaxUint16 + sizeEndCentralDirectory)
	if size > filesize {
		size = filesize
	}
	if offset, err = r.Seek(-size, io.SeekEnd); err != nil {
		return 0, err
	}

	buf := make([]byte, size)
	if _, err := r.Read(buf); err != nil {
		return 0, err
	}
	index := bytes.Index(buf, []byte(signEndCentralDirectory))
	if index < 0 {
		return 0, errors.New("invalid zip format: not found end of central directory signature")
	}

	offset += int64(index)
	return offset, nil
}
