package zip

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"io/fs"
	"strings"
)

// Writer creates a zip file.
type Writer struct {
	w    io.WriteSeeker
	dirs []*centralDirectoryHeader
	pre  *fileWriter

	Comment string
}

// NewWriter returns zip.Writer that writes to io.WriteSeeker.
func NewWriter(w io.WriteSeeker) (*Writer, error) {
	return &Writer{
		w:    w,
		dirs: make([]*centralDirectoryHeader, 0),
	}, nil
}

// Create returns io.WriteCloser that creates a file with name.
// If the previous io.WriteCloser has not called Close, it is forced to close.
func (w *Writer) Create(name string) (io.WriteCloser, error) {
	fh := NewFileHeader(name)
	return w.CreateFromHeader(fh)
}

// CreateFromHeader returns io.WriteCloser that creates a file with FileHeader.
// This method updates the FileHeader. FileHeader is completed after Close is called.
// If the previous io.WriteCloser has not called Close, it is forced to close.
func (w *Writer) CreateFromHeader(fh *FileHeader) (io.WriteCloser, error) {
	if err := w.closePreviousFile(); err != nil {
		return nil, err
	}

	// update file header
	fh.MinimumVersion = 20
	fh.GenerateVersion = 20
	fh.GenerateOS = OS_MSDOS
	fh.CRC32 = 0            // update by FileWriter
	fh.CompressedSize = 0   // update by FileWriter
	fh.UncompressedSize = 0 // update by FileWriter

	namesize := len(fh.FileName)
	if strings.HasSuffix(fh.FileName, "/") {
		namesize -= 1

		// directory is only allowed Store method
		fh.Method = &MethodStore{}
	}
	if ok := fs.ValidPath(fh.FileName[:namesize]); !ok {
		return nil, fmt.Errorf("file name is invalid: %q", fh.FileName)
	}

	offset, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	h := &centralDirectoryHeader{}
	if err := h.copyFromHeader(fh); err != nil {
		return nil, err
	}
	h.localHeaderOffset = uint32(offset)

	fw := &fileWriter{
		w: w.w,
		h: h,
	}

	w.dirs = append(w.dirs, h)
	w.pre = fw
	return fw, nil
}

// Copy copies the zip.File to the writer.
// This method does not modify the File argument.
// If the previous io.WriteCloser has not called Close, it is forced to close.
func (w *Writer) Copy(f *File) error {
	r, err := f.OpenRaw()
	if err != nil {
		return err
	}
	defer r.Close()

	return w.CopyFromReader(&f.FileHeader, r)
}

// CopyFromReader copies the io.Reader uncompressed data to the writer.
// This method does not modify the FileHeader argument.
// If the previous io.WriteCloser has not called Close, it is forced to close.
func (w *Writer) CopyFromReader(fh *FileHeader, r io.Reader) error {
	if err := w.closePreviousFile(); err != nil {
		return err
	}

	offset, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	h := &centralDirectoryHeader{}
	if err := h.copyFromHeader(fh); err != nil {
		return err
	}
	h.localHeaderOffset = uint32(offset)

	// write local file header
	lh := &localFileHeader{}
	if err := lh.copyFromHeader(fh); err != nil {
		return err
	}
	if _, err := lh.WriteTo(w.w); err != nil {
		return err
	}

	// write raw payload
	if _, err := io.Copy(w.w, r); err != nil {
		return err
	}

	// write data descriptor
	if fh.Flags.DataDescriptor {
		dd := &dataDescriptor{
			crc32:            fh.CRC32,
			compressedSize:   fh.CompressedSize,
			uncompressedSize: fh.UncompressedSize,
		}
		if _, err := dd.WriteTo(w.w); err != nil {
			return err
		}
	}

	w.dirs = append(w.dirs, h)
	w.pre = nil
	return nil
}

// Close flushes the write data and closes zip.Writer.
// If the previous FileWriter has not called Close, it is forced to close.
func (w *Writer) Close() error {
	if err := w.closePreviousFile(); err != nil {
		return err
	}
	return w.writeCentralDirectories()
}

// closePreviousFile closes the previous FileWriter.
func (w *Writer) closePreviousFile() (err error) {
	if w.pre != nil && !w.pre.IsClosed() {
		err = w.pre.Close()
		w.pre = nil
	}
	return err
}

// writeCentralDirectories writes central directory headers.
func (w *Writer) writeCentralDirectories() error {
	startOffset, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	for _, dir := range w.dirs {
		if _, err := dir.WriteTo(w.w); err != nil {
			return err
		}
	}

	endOffset, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	end := &endCentralDirectory{
		numberOfEntriesThisDisk:  uint16(len(w.dirs)),
		numberOfEntries:          uint16(len(w.dirs)),
		sizeOfCentralDirectories: uint32(endOffset - startOffset),
		offsetCentralDirectory:   uint32(startOffset),
		comment:                  []byte(w.Comment),
	}

	if _, err := end.WriteTo(w.w); err != nil {
		return err
	}
	return nil
}

// fileWriter implements a io.WriteCloser that writes compressed data.
type fileWriter struct {
	w io.WriteSeeker          // raw Writer
	h *centralDirectoryHeader // reference to central directory header

	compCounter   *countWriter   // compress size counter
	compWriter    io.WriteCloser // compress Writer
	uncompCounter *countWriter   // uncompress size counter
	crc32         hash.Hash32    // hash calclator
	fw            io.Writer      // file data Writer
	fh            *FileHeader
	initialized   bool
	closed        bool
}

// Write compresses and writes []byte.
func (fw *fileWriter) Write(p []byte) (int, error) {
	if !fw.initialized {
		fw.writeInit()
	}

	return fw.fw.Write(p)
}

// Close flushes the write data and closes zip.FileWriter.
// If FlagDataDescriptor is not set, the file header is rewritten.
func (fw *fileWriter) Close() error {
	if fw.closed {
		return errors.New("already closed")
	}
	fw.closed = true

	if fw.initialized {
		if err := fw.compWriter.Close(); err != nil {
			return err
		}
		fw.fh.CRC32 = fw.crc32.Sum32()
		fw.fh.CompressedSize = uint32(fw.compCounter.Count)
		fw.fh.UncompressedSize = uint32(fw.uncompCounter.Count)
	}
	if err := fw.h.copyFromHeader(fw.fh); err != nil {
		return err
	}

	if fw.fh.Flags.DataDescriptor {
		return fw.writeDataDescriptor()
	} else {
		return fw.rewriteFileHeader()
	}
}

// IsClosed returns whether the fileWriter is closed.
func (fw *fileWriter) IsClosed() bool {
	return fw.closed
}

// writeInit performs the preprocessing for writing and write the file header.
func (fw *fileWriter) writeInit() error {
	var err error
	fw.initialized = true

	fw.fh = &FileHeader{}
	if err := fw.h.copyToHeader(fw.fh); err != nil {
		return err
	}

	fw.compCounter = &countWriter{w: fw.w}
	fw.compWriter, err = fw.fh.Method.newCompressor(fw.compCounter)
	if err != nil {
		return err
	}
	fw.uncompCounter = &countWriter{w: fw.compWriter}
	fw.crc32 = crc32.NewIEEE()

	fw.fw = io.MultiWriter(
		fw.uncompCounter,
		fw.crc32,
	)

	return fw.writeFileHeader()
}

// writeFileHeader writes the file header.
func (fw *fileWriter) writeFileHeader() error {
	h := new(localFileHeader)
	if err := h.copyFromHeader(fw.fh); err != nil {
		return err
	}
	_, err := h.WriteTo(fw.w)
	return err
}

// writeDataDescriptor writes the data descriptor.
func (fw *fileWriter) writeDataDescriptor() error {
	dd := new(dataDescriptor)
	dd.crc32 = fw.fh.CRC32
	dd.compressedSize = fw.fh.CompressedSize
	dd.uncompressedSize = fw.fh.UncompressedSize

	_, err := dd.WriteTo(fw.w)
	return err
}

// rewriteFileHeader rewrites the file header.
func (fw *fileWriter) rewriteFileHeader() error {
	current, err := fw.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if _, err := fw.w.Seek(int64(fw.h.localHeaderOffset), io.SeekStart); err != nil {
		return err
	}
	if err := fw.writeFileHeader(); err != nil {
		return err
	}

	_, err = fw.w.Seek(current, io.SeekStart)
	return err
}
