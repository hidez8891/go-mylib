package zip

import (
	"bytes"
	"errors"
	"io"
	"time"

	"go-mylib/byteio"
)

const (
	signLocalFileHeader        string = "PK\x03\x04" // signature of a local file header
	signCentralDirectoryHeader string = "PK\x01\x02" // signature of a central directory header
	signDataDescriptor         string = "PK\x07\x08" // signature of a data descriptor
	signEndCentralDirectory    string = "PK\x05\x06" // signature of an end of central directory record

	sizeLocalFileHeader        int = 30 // size of a local file header
	sizeCentralDirectoryHeader int = 46 // size of a central directory header
	sizeDataDescriptor         int = 16 // size of a data descriptor
	sizeEndCentralDirectory    int = 22 // size of an end of central directory record
)

const (
	MadeByMSDOS uint16 = 0x0000 // version for MS-DOS and OS/2 (FAT, FAT32)
	MadeByUNIX  uint16 = 0x0300 // version for UNIX
	MadeByNEFS  uint16 = 0x0a00 // version for Windows NTFS
	MadeByOSX   uint16 = 0x1300 // version for OS X
)

const (
	flagDataDescriptor uint16 = 0x0008 // flag for data descriptor
	flagUTF8           uint16 = 0x0800 // flag for UTF-8
)

// FlagType represents flags of a zip header.
type FlagType struct {
	DataDescriptor bool
	UTF8           bool
}

// set sets FlagType from a zip header's flags.
func (f *FlagType) set(flag uint16) error {
	f.DataDescriptor = flag&flagDataDescriptor != 0
	f.UTF8 = flag&flagUTF8 != 0
	return nil
}

// get returns a zip header's flags.
func (f *FlagType) get() (flag uint16) {
	if f.DataDescriptor {
		flag |= flagDataDescriptor
	}
	if f.UTF8 {
		flag |= flagUTF8
	}
	return flag
}

// MethodType represents a compression method.
type MethodType interface {
	ID() uint16
	set(flag uint16) error
	get() uint16
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

// localFileHeader represents a local file header in the ZIP specification.
type localFileHeader struct {
	RequireVersion   uint16     // version needed to extract
	Flags            FlagType   // general purpose bit flag
	Method           MethodType // compression method
	ModifiedTime     time.Time  // last modified file date/time
	CRC32            uint32     // CRC-32 for uncompressed data
	CompressedSize   uint32     // compressed data size
	UncompressedSize uint32     // uncompressed data size
	FileName         string     // file name
	ExtraFields      []byte     // extra field data
}

// ReadFrom reads a local file header from io.Reader.
func (h *localFileHeader) ReadFrom(r io.Reader) (int64, error) {
	var sign [4]byte
	if _, err := r.Read(sign[:]); err != nil {
		return 0, err
	}
	if string(sign[:]) != signLocalFileHeader {
		return 0, errors.New("invalid zip format: not found local file header signature")
	}

	var data [sizeLocalFileHeader - 4]byte
	if _, err := io.ReadAtLeast(r, data[:], len(data)); err != nil {
		return 0, err
	}

	var (
		flag      uint16
		method    uint16
		modtime   uint16
		moddate   uint16
		nameSize  uint16
		extraSize uint16
	)
	rr := bytes.NewReader(data[:])
	byteio.GetUint16LE(rr, &h.RequireVersion)
	byteio.GetUint16LE(rr, &flag)
	byteio.GetUint16LE(rr, &method)
	byteio.GetUint16LE(rr, &modtime)
	byteio.GetUint16LE(rr, &moddate)
	byteio.GetUint32LE(rr, &h.CRC32)
	byteio.GetUint32LE(rr, &h.CompressedSize)
	byteio.GetUint32LE(rr, &h.UncompressedSize)
	byteio.GetUint16LE(rr, &nameSize)
	byteio.GetUint16LE(rr, &extraSize)

	if m, err := methodFactory(method); err != nil {
		return 0, err
	} else {
		h.Method = m
	}

	h.Flags.set(flag)
	h.Method.set(flag)
	h.ModifiedTime = uint32ToUTCTime(moddate, modtime)

	if nameSize == 0 {
		return 0, errors.New("invalid file name: name length is 0")
	}
	nameBuf := make([]byte, nameSize)
	if _, err := io.ReadAtLeast(r, nameBuf, len(nameBuf)); err != nil {
		return 0, err
	}
	h.FileName = string(nameBuf)

	if extraSize != 0 {
		extraBuf := make([]byte, extraSize)
		if _, err := io.ReadAtLeast(r, extraBuf, len(extraBuf)); err != nil {
			return 0, err
		}
		h.ExtraFields = extraBuf
	}

	return int64(sizeLocalFileHeader) + int64(nameSize) + int64(extraSize), nil
}

// WriteTo writes a local file header to io.Writer.
func (h *localFileHeader) WriteTo(w io.Writer) (int64, error) {
	flag := h.Flags.get()
	flag |= h.Method.get()
	method := h.Method.ID()
	moddate, modtime := utcTimeToUint32(h.ModifiedTime)

	buf := new(bytes.Buffer)
	buf.Write([]byte(signLocalFileHeader))
	byteio.WriteUint16LE(buf, h.RequireVersion)
	byteio.WriteUint16LE(buf, flag)
	byteio.WriteUint16LE(buf, method)
	byteio.WriteUint16LE(buf, modtime)
	byteio.WriteUint16LE(buf, moddate)
	byteio.WriteUint32LE(buf, h.CRC32)
	byteio.WriteUint32LE(buf, h.CompressedSize)
	byteio.WriteUint32LE(buf, h.UncompressedSize)
	byteio.WriteUint16LE(buf, uint16(len(h.FileName)))
	byteio.WriteUint16LE(buf, uint16(len(h.ExtraFields)))

	if len(h.FileName) == 0 {
		return 0, errors.New("invalid file name: name length is 0")
	}
	if _, err := buf.Write([]byte(h.FileName)); err != nil {
		return 0, err
	}

	if len(h.ExtraFields) != 0 {
		if _, err := buf.Write([]byte(h.ExtraFields)); err != nil {
			return 0, err
		}
	}

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// centralDirectoryHeader represents a central directory header in the ZIP specification.
type centralDirectoryHeader struct {
	localFileHeader
	GenerateVersion   uint16 // version made by
	InternalFileAttr  uint16 // internal file attributes
	ExternalFileAttr  uint32 // external file attributes
	LocalHeaderOffset uint32 // relative offset of local header
	Comment           string // file comment
}

// ReadFrom reads a central directory header from io.Reader.
func (h *centralDirectoryHeader) ReadFrom(r io.Reader) (int64, error) {
	var sign [4]byte
	if _, err := r.Read(sign[:]); err != nil {
		return 0, err
	}
	if string(sign[:]) != signCentralDirectoryHeader {
		return 0, errors.New("invalid zip format: not found central directory header signature")
	}

	var data [sizeCentralDirectoryHeader - 4]byte
	if _, err := io.ReadAtLeast(r, data[:], len(data)); err != nil {
		return 0, err
	}

	var (
		flag        uint16
		method      uint16
		modtime     uint16
		moddate     uint16
		nameSize    uint16
		extraSize   uint16
		commentSize uint16
		diskNumber  uint16
	)
	rr := bytes.NewReader(data[:])
	byteio.GetUint16LE(rr, &h.GenerateVersion)
	byteio.GetUint16LE(rr, &h.RequireVersion)
	byteio.GetUint16LE(rr, &flag)
	byteio.GetUint16LE(rr, &method)
	byteio.GetUint16LE(rr, &modtime)
	byteio.GetUint16LE(rr, &moddate)
	byteio.GetUint32LE(rr, &h.CRC32)
	byteio.GetUint32LE(rr, &h.CompressedSize)
	byteio.GetUint32LE(rr, &h.UncompressedSize)
	byteio.GetUint16LE(rr, &nameSize)
	byteio.GetUint16LE(rr, &extraSize)
	byteio.GetUint16LE(rr, &commentSize)
	byteio.GetUint16LE(rr, &diskNumber)
	byteio.GetUint16LE(rr, &h.InternalFileAttr)
	byteio.GetUint32LE(rr, &h.ExternalFileAttr)
	byteio.GetUint32LE(rr, &h.LocalHeaderOffset)

	if m, err := methodFactory(method); err != nil {
		return 0, err
	} else {
		h.Method = m
	}

	h.Flags.set(flag)
	h.Method.set(flag)
	h.ModifiedTime = uint32ToUTCTime(moddate, modtime)

	if diskNumber != 0 {
		return 0, errors.New("unsupport split zip file")
	}

	if nameSize == 0 {
		return 0, errors.New("invalid file name: name length is 0")
	}
	nameBuf := make([]byte, nameSize)
	if _, err := io.ReadAtLeast(r, nameBuf, len(nameBuf)); err != nil {
		return 0, err
	}
	h.FileName = string(nameBuf)

	if extraSize != 0 {
		extraBuf := make([]byte, extraSize)
		if _, err := io.ReadAtLeast(r, extraBuf, len(extraBuf)); err != nil {
			return 0, err
		}
		h.ExtraFields = extraBuf
	}

	if commentSize != 0 {
		commentBuf := make([]byte, commentSize)
		if _, err := io.ReadAtLeast(r, commentBuf, len(commentBuf)); err != nil {
			return 0, err
		}
		h.Comment = string(commentBuf)
	}

	size := sizeLocalFileHeader
	size += int(nameSize) + int(extraSize) + int(commentSize)
	return int64(size), nil
}

// WriteTo writes a central directory header to io.Writer.
func (h *centralDirectoryHeader) WriteTo(w io.Writer) (int64, error) {
	flag := h.Flags.get()
	flag |= h.Method.get()
	method := h.Method.ID()
	moddate, modtime := utcTimeToUint32(h.ModifiedTime)

	buf := new(bytes.Buffer)
	w.Write([]byte(signCentralDirectoryHeader))
	byteio.WriteUint16LE(buf, h.GenerateVersion)
	byteio.WriteUint16LE(buf, h.RequireVersion)
	byteio.WriteUint16LE(buf, flag)
	byteio.WriteUint16LE(buf, method)
	byteio.WriteUint16LE(buf, modtime)
	byteio.WriteUint16LE(buf, moddate)
	byteio.WriteUint32LE(buf, h.CRC32)
	byteio.WriteUint32LE(buf, h.CompressedSize)
	byteio.WriteUint32LE(buf, h.UncompressedSize)
	byteio.WriteUint16LE(buf, uint16(len(h.FileName)))
	byteio.WriteUint16LE(buf, uint16(len(h.ExtraFields)))
	byteio.WriteUint16LE(buf, uint16(len(h.Comment)))
	byteio.WriteUint16LE(buf, 0)
	byteio.WriteUint16LE(buf, h.InternalFileAttr)
	byteio.WriteUint32LE(buf, h.ExternalFileAttr)
	byteio.WriteUint32LE(buf, h.LocalHeaderOffset)

	if len(h.FileName) == 0 {
		return 0, errors.New("invalid file name: name length is 0")
	}
	if _, err := buf.Write([]byte(h.FileName)); err != nil {
		return 0, err
	}

	if len(h.ExtraFields) != 0 {
		if _, err := buf.Write(h.ExtraFields); err != nil {
			return 0, err
		}
	}

	if len(h.Comment) != 0 {
		if _, err := buf.Write([]byte(h.Comment)); err != nil {
			return 0, err
		}
	}

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// dataDescriptor represents a data descriptor in the ZIP specification.
type dataDescriptor struct {
	CRC32            uint32 // CRC-32 for uncompressed data
	CompressedSize   uint32 // compressed data size
	UncompressedSize uint32 // uncompressed data size
}

// ReadFrom reads a data descriptor from io.Reader.
func (d *dataDescriptor) ReadFrom(r io.Reader) (int64, error) {
	var buf [sizeDataDescriptor]byte
	size := len(buf) - 4
	if _, err := io.ReadAtLeast(r, buf[:size], size); err != nil {
		return 0, err
	}

	var sign [4]byte
	rr := bytes.NewReader(buf[:4])
	if _, err := rr.Read(sign[:]); err != nil {
		return 0, err
	}
	if string(sign[:]) == signDataDescriptor {
		// load additional data
		size += 4
		if _, err := io.ReadAtLeast(r, buf[size-4:], 4); err != nil {
			return 0, err
		}
		rr = bytes.NewReader(buf[4:])
	} else {
		rr = bytes.NewReader(buf[:size])
	}

	byteio.GetUint32LE(rr, &d.CRC32)
	byteio.GetUint32LE(rr, &d.CompressedSize)
	byteio.GetUint32LE(rr, &d.UncompressedSize)

	return int64(size), nil
}

// WriteTo writes a data descriptor to io.Writer.
func (d *dataDescriptor) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	buf.Write([]byte(signDataDescriptor))
	byteio.WriteUint32LE(buf, d.CRC32)
	byteio.WriteUint32LE(buf, d.CompressedSize)
	byteio.WriteUint32LE(buf, d.UncompressedSize)

	if _, err := w.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return int64(buf.Len()), nil
}

// endCentralDirectory represents an end of central directory record in the ZIP specification.
type endCentralDirectory struct {
	numberOfEntries          uint16 // total number of entries in the central directory on this disk
	sizeOfCentralDirectories uint32 // size of the central directory block
	offsetCentralDirectory   uint32 // offset of start of central directory with respect to the starting disk number
	Comment                  string // zip archive comment
}

// ReadFrom reads an end of central directory record from io.Reader.
func (e *endCentralDirectory) ReadFrom(r io.Reader) (int64, error) {
	var sign [4]byte
	if _, err := r.Read(sign[:]); err != nil {
		return 0, err
	}
	if string(sign[:]) != signEndCentralDirectory {
		return 0, errors.New("invalid zip format: not found end of central directory signature")
	}

	var buf [sizeEndCentralDirectory - 4]byte
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	rr := bytes.NewReader(buf[:])
	var (
		numThisDisk      uint16
		numStartDirDisk  uint16
		numEntriesInDisk uint16
		commentSize      uint16
	)
	byteio.GetUint16LE(rr, &numThisDisk)
	byteio.GetUint16LE(rr, &numStartDirDisk)
	byteio.GetUint16LE(rr, &numEntriesInDisk)
	byteio.GetUint16LE(rr, &e.numberOfEntries)
	byteio.GetUint32LE(rr, &e.sizeOfCentralDirectories)
	byteio.GetUint32LE(rr, &e.offsetCentralDirectory)
	byteio.GetUint16LE(rr, &commentSize)

	if numThisDisk != 0 || numStartDirDisk != 0 || numEntriesInDisk != e.numberOfEntries {
		return 0, errors.New("unsupport split zip file")
	}

	if commentSize != 0 {
		commentBuf := make([]byte, commentSize)
		if _, err := r.Read(commentBuf); err != nil {
			return 0, err
		}
		e.Comment = string(commentBuf)
	}

	size := int(sizeEndCentralDirectory) + int(commentSize)
	return int64(size), nil
}

// WriteTo writes an end of central directory record to io.Writer.
func (e *endCentralDirectory) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	w.Write([]byte(signEndCentralDirectory))
	byteio.WriteUint16LE(buf, 0)
	byteio.WriteUint16LE(buf, 0)
	byteio.WriteUint16LE(buf, e.numberOfEntries)
	byteio.WriteUint16LE(buf, e.numberOfEntries)
	byteio.WriteUint32LE(buf, e.sizeOfCentralDirectories)
	byteio.WriteUint32LE(buf, e.offsetCentralDirectory)
	byteio.WriteUint16LE(buf, uint16(len(e.Comment)))

	if _, err := buf.Write([]byte(e.Comment)); err != nil {
		return 0, err
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return int64(buf.Len()), nil
}

// uint32ToUTCTime converts a date/time in uint32 format to time.Time format.
func uint32ToUTCTime(dates uint16, times uint16) time.Time {
	year := int((dates >> 9) & 0x7f)
	monh := int((dates >> 5) & 0x0f)
	days := int((dates >> 0) & 0x1f)
	hour := int((times >> 11) & 0x1f)
	mins := int((times >> 05) & 0x3f)
	secs := int((times >> 00) & 0x1f)

	year += 1980
	secs *= 2

	return time.Date(year, time.Month(monh), days, hour, mins, secs, 0, time.UTC)
}

// utcTimeToUint32 converts time.Time format to uint32 format date/time.
func utcTimeToUint32(t time.Time) (dates uint16, times uint16) {
	t = t.UTC()
	dates |= (uint16(t.Year()) - 1980) << 9
	dates |= uint16(t.Month()) << 5
	dates |= uint16(t.Day())
	times |= uint16(t.Hour()) << 11
	times |= uint16(t.Minute()) << 5
	times |= uint16(t.Second()) / 2
	return dates, times
}
