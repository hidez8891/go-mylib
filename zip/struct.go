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

const (
	madebyMSDOS   uint16 = 0x00 // made by MS-DOS
	madebyUNIX    uint16 = 0x03 // made by UNIX
	madebyWindows uint16 = 0x0a // made by Windows
	madebyOSX     uint16 = 0x13 // made by OS X (Darwin)
)

// VersionType represents a version of a zip content.
type VersionType uint16

func newVersionType(version int) VersionType {
	var v VersionType
	v.SetVersion(version)
	return v
}

// Version returns a version number of a zip content.
func (v VersionType) Version() int {
	return int(v & 0x00ff)
}

// MadeOS returns a generate OS of a zip content.
func (v VersionType) MadeOS() string {
	switch uint16(v >> 8) {
	case madebyMSDOS:
		return "MS-DOS"
	case madebyUNIX:
		return "UNIX"
	case madebyWindows:
		return "Windows"
	case madebyOSX:
		return "OS X"
	default:
		return "Unknown OS"
	}
}

// SetVersion sets a version number of a zip content.
func (v *VersionType) SetVersion(version int) {
	*v = VersionType((uint16(version) & 0x00ff) | (uint16(*v) & 0xff00))
}

// SetMadeOS sets a generate OS of a zip content.
func (v *VersionType) SetMadeOS(os string) {
	var id uint16
	switch os {
	case "MS-DOS":
		id = madebyMSDOS
	case "UNIX":
		id = madebyUNIX
	case "Windows":
		id = madebyWindows
	case "OS X":
		id = madebyOSX
	default:
		id = 0x00 // default value
	}

	*v = VersionType((id << 8) | (uint16(*v) & 0x00ff))
}

// localFileHeader represents a local file header in the ZIP specification.
type localFileHeader struct {
	RequireVersion   VersionType  // version needed to extract
	Flags            FlagType     // general purpose bit flag
	Method           MethodType   // compression method
	ModifiedTime     time.Time    // last modified file date/time
	CRC32            uint32       // CRC-32 for uncompressed data
	CompressedSize   uint32       // compressed data size
	UncompressedSize uint32       // uncompressed data size
	FileName         string       // file name
	ExtraFields      []ExtraField // extra field data
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
	byteio.GetUint16LE(rr, (*uint16)(&h.RequireVersion))
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

	h.ExtraFields = make([]ExtraField, 0)
	if extraSize != 0 {
		extraBuf := make([]byte, extraSize)
		if _, err := io.ReadAtLeast(r, extraBuf, len(extraBuf)); err != nil {
			return 0, err
		}

		if extras, err := parseExtraFields(extraBuf); err != nil {
			return 0, err
		} else {
			h.ExtraFields = extras
		}
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
	byteio.WriteUint16LE(buf, uint16(h.RequireVersion))
	byteio.WriteUint16LE(buf, flag)
	byteio.WriteUint16LE(buf, method)
	byteio.WriteUint16LE(buf, modtime)
	byteio.WriteUint16LE(buf, moddate)
	byteio.WriteUint32LE(buf, h.CRC32)
	byteio.WriteUint32LE(buf, h.CompressedSize)
	byteio.WriteUint32LE(buf, h.UncompressedSize)

	if len(h.FileName) == 0 {
		return 0, errors.New("invalid file name: name length is 0")
	}
	byteio.WriteUint16LE(buf, uint16(len(h.FileName)))

	extras := new(bytes.Buffer)
	for _, extra := range h.ExtraFields {
		if _, err := extra.WriteTo(extras); err != nil {
			return 0, err
		}
	}
	byteio.WriteUint16LE(buf, uint16(extras.Len()))

	buf.Write([]byte(h.FileName))
	buf.Write(extras.Bytes())

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// centralDirectoryHeader represents a central directory header in the ZIP specification.
type centralDirectoryHeader struct {
	localFileHeader
	GenerateVersion   VersionType // version made by
	InternalFileAttr  uint16      // internal file attributes
	ExternalFileAttr  uint32      // external file attributes
	LocalHeaderOffset uint32      // relative offset of local header
	Comment           string      // file comment
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
	byteio.GetUint16LE(rr, (*uint16)(&h.GenerateVersion))
	byteio.GetUint16LE(rr, (*uint16)(&h.RequireVersion))
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

	h.ExtraFields = make([]ExtraField, 0)
	if extraSize != 0 {
		extraBuf := make([]byte, extraSize)
		if _, err := io.ReadAtLeast(r, extraBuf, len(extraBuf)); err != nil {
			return 0, err
		}

		if extras, err := parseExtraFields(extraBuf); err != nil {
			return 0, err
		} else {
			h.ExtraFields = extras
		}
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
	byteio.WriteUint16LE(buf, uint16(h.GenerateVersion))
	byteio.WriteUint16LE(buf, uint16(h.RequireVersion))
	byteio.WriteUint16LE(buf, flag)
	byteio.WriteUint16LE(buf, method)
	byteio.WriteUint16LE(buf, modtime)
	byteio.WriteUint16LE(buf, moddate)
	byteio.WriteUint32LE(buf, h.CRC32)
	byteio.WriteUint32LE(buf, h.CompressedSize)
	byteio.WriteUint32LE(buf, h.UncompressedSize)

	if len(h.FileName) == 0 {
		return 0, errors.New("invalid file name: name length is 0")
	}
	byteio.WriteUint16LE(buf, uint16(len(h.FileName)))

	extras := new(bytes.Buffer)
	for _, extra := range h.ExtraFields {
		if _, err := extra.WriteTo(extras); err != nil {
			return 0, err
		}
	}
	byteio.WriteUint16LE(buf, uint16(extras.Len()))

	byteio.WriteUint16LE(buf, uint16(len(h.Comment)))
	byteio.WriteUint16LE(buf, 0)
	byteio.WriteUint16LE(buf, h.InternalFileAttr)
	byteio.WriteUint32LE(buf, h.ExternalFileAttr)
	byteio.WriteUint32LE(buf, h.LocalHeaderOffset)

	buf.Write([]byte(h.FileName))
	buf.Write(extras.Bytes())
	buf.Write([]byte(h.Comment))

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
	if dates == 0 && times == 0 {
		return time.Time{}
	}

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
	if t.IsZero() || t.Year() < 1980 {
		return 0, 0
	}

	t = t.UTC()
	dates |= (uint16(t.Year()) - 1980) << 9
	dates |= uint16(t.Month()) << 5
	dates |= uint16(t.Day())
	times |= uint16(t.Hour()) << 11
	times |= uint16(t.Minute()) << 5
	times |= uint16(t.Second()) / 2
	return dates, times
}
