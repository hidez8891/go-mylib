package zip

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"io/fs"
	"strings"
	"time"
)

// Writer creates a zip file.
type Writer struct {
	w    io.WriteSeeker
	dirs []*centralDirectoryHeader

	Comment string
}

// NewWriter returns zip.Writer that writes to io.WriteSeeker.
func NewWriter(w io.WriteSeeker) (*Writer, error) {
	return &Writer{
		w:    w,
		dirs: make([]*centralDirectoryHeader, 0),
	}, nil
}

// Create returns zip.FileWriter that creates a file with name.
func (w *Writer) Create(name string) (*FileWriter, error) {
	namesize := len(name)
	if strings.HasSuffix(name, "/") {
		namesize -= 1
	}
	if ok := fs.ValidPath(name[:namesize]); !ok {
		return nil, fmt.Errorf("file name is invalid: %q", name)
	}

	offset, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	h := &centralDirectoryHeader{
		localFileHeader: localFileHeader{
			RequireVersion: 20, // require deflate compression
			Flags:          0,
			Method:         MethodDeflated,
			ModifiedTime:   time.Now(),
			FileName:       name,
		},
		GenerateVersion:   MadeByMSDOS | 20,
		LocalHeaderOffset: uint32(offset),
	}

	if strings.HasSuffix(name, "/") {
		// directory is only allowed Store method
		h.Method = MethodStored
	}

	w.dirs = append(w.dirs, h)

	return &FileWriter{
		w: w.w,
		h: h,
	}, nil
}

// Close flushes the write data and closes zip.Writer.
func (w *Writer) Close() error {
	return w.writeCentralDirectories()
}

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
		numberOfEntries:          uint16(len(w.dirs)),
		sizeOfCentralDirectories: uint32(endOffset - startOffset),
		offsetCentralDirectory:   uint32(startOffset),
		Comment:                  w.Comment,
	}

	if _, err := end.WriteTo(w.w); err != nil {
		return err
	}
	return nil
}

// FileWriter implements a io.Writer that compresses and writes data.
type FileWriter struct {
	w io.WriteSeeker // raw Writer
	h *centralDirectoryHeader

	compCounter   *CountWriter   // compress size counter
	compWriter    io.WriteCloser // compress Writer
	uncompCounter *CountWriter   // uncompress size counter
	crc32         hash.Hash32    // hash calclator
	fw            io.Writer      // file data Writer
	initialized   bool

	Comment string
}

// newFileWriter returns zip.FileWriter that writes to io.WriteSeeker.
func newFileWriter(w io.WriteSeeker, h *centralDirectoryHeader) *FileWriter {
	return &FileWriter{
		w: w,
		h: h,
	}
}

// Flags returns the current flags.
func (fw *FileWriter) Flags() uint16 {
	return fw.h.Flags
}

// SetFlag sets additional flags.
// This function should not be called after writing data.
func (fw *FileWriter) SetFlags(flags uint16) error {
	if fw.initialized {
		return errors.New("operation is invalid after writing")
	}
	fw.h.Flags |= flags
	return nil
}

// UnsetFlag clears the specified flags.
// This function should not be called after writing data.
func (fw *FileWriter) UnsetFlags(flags uint16) error {
	if fw.initialized {
		return errors.New("operation is invalid after writing")
	}
	fw.h.Flags &^= flags
	return nil
}

// Name returns the file name.
func (fw *FileWriter) Name() string {
	return fw.h.FileName
}

// Method returns the method ID of the data compression mode.
func (fw *FileWriter) Method() uint16 {
	return fw.h.Method
}

// SetMethod updates the method ID of the data compression mode.
// This function should not be called after writing data.
func (fw *FileWriter) SetMethod(methodID uint16) error {
	if strings.HasSuffix(fw.h.FileName, "/") {
		return errors.New("directory path does not allow Method update")
	}
	if fw.initialized {
		return errors.New("operation is invalid after writing")
	}
	fw.h.Method = methodID
	return nil
}

// ModifiedTime returns the modified time.
func (fw *FileWriter) ModifiedTime() time.Time {
	return fw.h.ModifiedTime
}

// SetModifiedTime updates the modified time.
// This function should not be called after writing data.
func (fw *FileWriter) SetModifiedTime(t time.Time) error {
	if fw.initialized {
		return errors.New("operation is invalid after writing")
	}
	fw.h.ModifiedTime = t
	return nil
}

// Write compresses and writes []byte.
func (fw *FileWriter) Write(p []byte) (int, error) {
	if !fw.initialized {
		fw.writeInit()
	}

	return fw.fw.Write(p)
}

// Close flushes the write data and closes zip.FileWriter.
// If FlagDataDescriptor is not set, the file header is rewritten.
func (fw *FileWriter) Close() error {
	if fw.initialized {
		if err := fw.compWriter.Close(); err != nil {
			return err
		}
		fw.h.CRC32 = fw.crc32.Sum32()
		fw.h.CompressedSize = uint32(fw.compCounter.Count)
		fw.h.UncompressedSize = uint32(fw.uncompCounter.Count)
		fw.h.Comment = fw.Comment
	}

	if fw.h.Flags&FlagDataDescriptor != 0 {
		return fw.writeDataDescriptor()
	} else {
		return fw.rewriteFileHeader()
	}
}

// writeInit performs the preprocessing for writing and write the file header.
func (fw *FileWriter) writeInit() error {
	fw.initialized = true

	comp, ok := compressors[fw.h.Method]
	if !ok {
		return errors.New("Unsupport compress method")
	}

	fw.compCounter = &CountWriter{w: fw.w}
	fw.compWriter = comp(fw.compCounter, fw.h.Flags)
	fw.uncompCounter = &CountWriter{w: fw.compWriter}
	fw.crc32 = crc32.NewIEEE()

	fw.fw = io.MultiWriter(
		fw.uncompCounter,
		fw.crc32,
	)

	return fw.writeFileHeader()
}

// writeFileHeader writes the file header.
func (fw *FileWriter) writeFileHeader() error {
	_, err := fw.h.localFileHeader.WriteTo(fw.w)
	return err
}

// writeDataDescriptor writes the data descriptor.
func (fw *FileWriter) writeDataDescriptor() error {
	dd := new(dataDescriptor)
	dd.CRC32 = fw.h.CRC32
	dd.CompressedSize = fw.h.CompressedSize
	dd.UncompressedSize = fw.h.UncompressedSize

	_, err := dd.WriteTo(fw.w)
	return err
}

// rewriteFileHeader rewrites the file header.
func (fw *FileWriter) rewriteFileHeader() error {
	current, err := fw.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	offset := fw.h.LocalHeaderOffset
	if _, err := fw.w.Seek(int64(offset), io.SeekStart); err != nil {
		return err
	}
	if err := fw.writeFileHeader(); err != nil {
		return err
	}

	_, err = fw.w.Seek(current, io.SeekStart)
	return err
}
