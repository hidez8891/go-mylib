package zip

import (
	"bytes"
	"errors"
	"io"
	"time"

	"go-mylib/byteio"
)

// FileHeader represents a file header in a zip file.
type FileHeader struct {
	MinimumVersion   int          // version needed to extract the file
	GenerateVersion  int          // version used to generate the file
	GenerateOS       OSType       // operating system used to generate the file
	Flags            FlagType     // general purpose flag
	Method           MethodType   // compression method
	ModifiedTime     time.Time    // last modification time
	CRC32            uint32       // CRC-32 for uncompressed data
	CompressedSize   uint32       // compressed data size
	UncompressedSize uint32       // uncompressed data size
	FileName         string       // file name
	ExtraFields      []ExtraField // extra field data
	InternalFileAttr uint16       // internal file attributes
	ExternalFileAttr uint32       // external file attributes
	Comment          string       // file comment
}

// NewFileHeader creates a new FileHeader.
func NewFileHeader(name string) *FileHeader {
	return &FileHeader{
		Method:       &MethodDeflated{DefaultCompression},
		ModifiedTime: time.Time{},
		FileName:     name,
		ExtraFields:  make([]ExtraField, 0),
	}
}

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

// OSType represents operating system ID.
type OSType uint16

const (
	OS_MSDOS OSType = 0  // MS-DOS
	OS_UNIX  OSType = 3  // UNIX
	OS_NTFS  OSType = 10 // NTFS
	OS_OSX   OSType = 19 // OSX
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

// localFileHeader represents a local file header in the ZIP specification.
type localFileHeader struct {
	minimumVersion   uint16 // version needed to extract
	flag             uint16 // general purpose bit flag
	method           uint16 // compression method
	modtime          uint16 // last modified file time
	moddate          uint16 // last modified file date
	crc32            uint32 // CRC-32 for uncompressed data
	compressedSize   uint32 // compressed data size
	uncompressedSize uint32 // uncompressed data size
	fileName         []byte // file name
	extraFields      []byte // extra field data
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
		nameSize  uint16
		extraSize uint16
	)
	rr := bytes.NewReader(data[:])
	byteio.GetUint16LE(rr, &h.minimumVersion)
	byteio.GetUint16LE(rr, &h.flag)
	byteio.GetUint16LE(rr, &h.method)
	byteio.GetUint16LE(rr, &h.modtime)
	byteio.GetUint16LE(rr, &h.moddate)
	byteio.GetUint32LE(rr, &h.crc32)
	byteio.GetUint32LE(rr, &h.compressedSize)
	byteio.GetUint32LE(rr, &h.uncompressedSize)
	byteio.GetUint16LE(rr, &nameSize)
	byteio.GetUint16LE(rr, &extraSize)

	if nameSize == 0 {
		return 0, errors.New("invalid local file header: name length is 0")
	}
	h.fileName = make([]byte, nameSize)
	if _, err := r.Read(h.fileName); err != nil {
		return 0, err
	}

	h.extraFields = make([]byte, extraSize)
	if extraSize != 0 {
		if _, err := r.Read(h.extraFields); err != nil {
			return 0, err
		}
	}

	size := int64(sizeLocalFileHeader)
	size += +int64(nameSize) + int64(extraSize)
	return size, nil
}

// WriteTo writes a local file header to io.Writer.
func (h *localFileHeader) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	buf.Write([]byte(signLocalFileHeader))
	byteio.WriteUint16LE(buf, h.minimumVersion)
	byteio.WriteUint16LE(buf, h.flag)
	byteio.WriteUint16LE(buf, h.method)
	byteio.WriteUint16LE(buf, h.modtime)
	byteio.WriteUint16LE(buf, h.moddate)
	byteio.WriteUint32LE(buf, h.crc32)
	byteio.WriteUint32LE(buf, h.compressedSize)
	byteio.WriteUint32LE(buf, h.uncompressedSize)

	if len(h.fileName) == 0 {
		return 0, errors.New("invalid local file header: name length is 0")
	}
	byteio.WriteUint16LE(buf, uint16(len(h.fileName)))

	byteio.WriteUint16LE(buf, uint16(len(h.extraFields)))
	buf.Write(h.fileName)
	buf.Write(h.extraFields)

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// copyFromHeader copies a local file header from FileHeader.
func (h *localFileHeader) copyFromHeader(fh *FileHeader) error {
	extras := new(bytes.Buffer)
	for _, extra := range fh.ExtraFields {
		if _, err := extra.WriteTo(extras); err != nil {
			return err
		}
	}

	h.minimumVersion = uint16(fh.MinimumVersion) & 0x00ff
	h.flag = fh.Flags.get() | fh.Method.get()
	h.method = fh.Method.ID()
	h.moddate, h.modtime = uint16FromDosTime(fh.ModifiedTime)
	if fh.Flags.DataDescriptor {
		// if data descriptor is set, crc32 and comp/uncompressed size are must be 0
		h.crc32 = 0
		h.compressedSize = 0
		h.uncompressedSize = 0
	} else {
		h.crc32 = fh.CRC32
		h.compressedSize = fh.CompressedSize
		h.uncompressedSize = fh.UncompressedSize
	}
	h.fileName = []byte(fh.FileName)
	// local file header extra field is should be empty
	h.extraFields = make([]byte, 0)

	return nil
}

// copyToHeader copies a local file header to FileHeader.
func (h *localFileHeader) copyToHeader(fh *FileHeader) error {
	method, err := methodFactory(h.method)
	if err != nil {
		return err
	}

	extra, err := parseExtraFields(h.extraFields)
	if err != nil {
		return err
	}

	fh.MinimumVersion = int(h.minimumVersion)
	fh.Flags.set(h.flag)
	fh.Method = method
	fh.Method.set(h.flag)
	fh.ModifiedTime = uint16ToDosTime(h.moddate, h.modtime)
	fh.CRC32 = h.crc32
	fh.CompressedSize = h.compressedSize
	fh.UncompressedSize = h.uncompressedSize
	fh.FileName = string(h.fileName)
	fh.ExtraFields = extra

	return nil
}

// centralDirectoryHeader represents a central directory header in the ZIP specification.
type centralDirectoryHeader struct {
	generateVersion   uint16 // version used to generate the file
	minimumVersion    uint16 // version needed to extract the file
	flag              uint16 // general purpose bit flag
	method            uint16 // compression method
	modtime           uint16 // last modified file time
	moddate           uint16 // last modified file date
	crc32             uint32 // CRC-32 for uncompressed data
	compressedSize    uint32 // compressed data size
	uncompressedSize  uint32 // uncompressed data size
	diskNumber        uint16 // disk number start
	internalFileAttr  uint16 // internal file attributes
	externalFileAttr  uint32 // external file attributes
	localHeaderOffset uint32 // relative offset of local header
	fileName          []byte // file name
	extraFields       []byte // extra field data
	comment           []byte // file comment
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
		nameSize    uint16
		extraSize   uint16
		commentSize uint16
	)
	rr := bytes.NewReader(data[:])
	byteio.GetUint16LE(rr, &h.generateVersion)
	byteio.GetUint16LE(rr, &h.minimumVersion)
	byteio.GetUint16LE(rr, &h.flag)
	byteio.GetUint16LE(rr, &h.method)
	byteio.GetUint16LE(rr, &h.modtime)
	byteio.GetUint16LE(rr, &h.moddate)
	byteio.GetUint32LE(rr, &h.crc32)
	byteio.GetUint32LE(rr, &h.compressedSize)
	byteio.GetUint32LE(rr, &h.uncompressedSize)
	byteio.GetUint16LE(rr, &nameSize)
	byteio.GetUint16LE(rr, &extraSize)
	byteio.GetUint16LE(rr, &commentSize)
	byteio.GetUint16LE(rr, &h.diskNumber)
	byteio.GetUint16LE(rr, &h.internalFileAttr)
	byteio.GetUint32LE(rr, &h.externalFileAttr)
	byteio.GetUint32LE(rr, &h.localHeaderOffset)

	if h.diskNumber != 0 {
		return 0, errors.New("unsupport split zip file")
	}

	if nameSize == 0 {
		return 0, errors.New("invalid central file header: name length is 0")
	}
	h.fileName = make([]byte, nameSize)
	if _, err := r.Read(h.fileName); err != nil {
		return 0, err
	}

	h.extraFields = make([]byte, extraSize)
	if extraSize != 0 {
		if _, err := r.Read(h.extraFields); err != nil {
			return 0, err
		}
	}

	h.comment = make([]byte, commentSize)
	if commentSize != 0 {
		if _, err := r.Read(h.comment); err != nil {
			return 0, err
		}
	}

	size := int64(sizeLocalFileHeader)
	size += int64(nameSize) + int64(extraSize) + int64(commentSize)
	return size, nil
}

// WriteTo writes a central directory header to io.Writer.
func (h *centralDirectoryHeader) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	w.Write([]byte(signCentralDirectoryHeader))
	byteio.WriteUint16LE(buf, h.generateVersion)
	byteio.WriteUint16LE(buf, h.minimumVersion)
	byteio.WriteUint16LE(buf, h.flag)
	byteio.WriteUint16LE(buf, h.method)
	byteio.WriteUint16LE(buf, h.modtime)
	byteio.WriteUint16LE(buf, h.moddate)
	byteio.WriteUint32LE(buf, h.crc32)
	byteio.WriteUint32LE(buf, h.compressedSize)
	byteio.WriteUint32LE(buf, h.uncompressedSize)

	if len(h.fileName) == 0 {
		return 0, errors.New("invalid central file header: name length is 0")
	}
	byteio.WriteUint16LE(buf, uint16(len(h.fileName)))

	byteio.WriteUint16LE(buf, uint16(len(h.extraFields)))
	byteio.WriteUint16LE(buf, uint16(len(h.comment)))
	byteio.WriteUint16LE(buf, h.diskNumber)
	byteio.WriteUint16LE(buf, h.internalFileAttr)
	byteio.WriteUint32LE(buf, h.externalFileAttr)
	byteio.WriteUint32LE(buf, h.localHeaderOffset)
	buf.Write(h.fileName)
	buf.Write(h.extraFields)
	buf.Write(h.comment)

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// copyFromHeader copies a central directory header from FileHeader.
func (h *centralDirectoryHeader) copyFromHeader(fh *FileHeader) error {
	extras := new(bytes.Buffer)
	for _, extra := range fh.ExtraFields {
		if _, err := extra.WriteTo(extras); err != nil {
			return err
		}
	}

	h.generateVersion = uint16(fh.GenerateOS)<<8 | uint16(fh.GenerateVersion)&0x00ff
	h.minimumVersion = uint16(fh.MinimumVersion) & 0x00ff
	h.flag = fh.Flags.get() | fh.Method.get()
	h.method = fh.Method.ID()
	h.moddate, h.modtime = uint16FromDosTime(fh.ModifiedTime)
	h.crc32 = fh.CRC32
	h.compressedSize = fh.CompressedSize
	h.uncompressedSize = fh.UncompressedSize
	h.fileName = []byte(fh.FileName)
	h.extraFields = extras.Bytes()
	h.internalFileAttr = fh.InternalFileAttr
	h.externalFileAttr = fh.ExternalFileAttr
	h.comment = []byte(fh.Comment)

	return nil
}

// copyToHeader copies a central directory header to FileHeader.
func (h *centralDirectoryHeader) copyToHeader(fh *FileHeader) error {
	method, err := methodFactory(h.method)
	if err != nil {
		return err
	}

	extra, err := parseExtraFields(h.extraFields)
	if err != nil {
		return err
	}

	fh.MinimumVersion = int(h.minimumVersion)
	fh.GenerateOS = OSType(h.generateVersion >> 8)
	fh.GenerateVersion = int(h.generateVersion & 0x00ff)
	fh.Flags.set(h.flag)
	fh.Method = method
	fh.Method.set(h.flag)
	fh.ModifiedTime = uint16ToDosTime(h.moddate, h.modtime)
	fh.CRC32 = h.crc32
	fh.CompressedSize = h.compressedSize
	fh.UncompressedSize = h.uncompressedSize
	fh.FileName = string(h.fileName)
	fh.ExtraFields = extra
	fh.InternalFileAttr = h.internalFileAttr
	fh.ExternalFileAttr = h.externalFileAttr
	fh.Comment = string(h.comment)

	return nil
}

// dataDescriptor represents a data descriptor in the ZIP specification.
type dataDescriptor struct {
	crc32            uint32 // CRC-32 for uncompressed data
	compressedSize   uint32 // compressed data size
	uncompressedSize uint32 // uncompressed data size
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

	byteio.GetUint32LE(rr, &d.crc32)
	byteio.GetUint32LE(rr, &d.compressedSize)
	byteio.GetUint32LE(rr, &d.uncompressedSize)

	return int64(size), nil
}

// WriteTo writes a data descriptor to io.Writer.
func (d *dataDescriptor) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	buf.Write([]byte(signDataDescriptor))
	byteio.WriteUint32LE(buf, d.crc32)
	byteio.WriteUint32LE(buf, d.compressedSize)
	byteio.WriteUint32LE(buf, d.uncompressedSize)

	if _, err := w.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return int64(buf.Len()), nil
}

// endCentralDirectory represents an end of central directory record in the ZIP specification.
type endCentralDirectory struct {
	numberOfDisk             uint16 // number of this disk
	numberOfStartDirDisk     uint16 // number of the disk with the start of the central directory.
	numberOfEntriesThisDisk  uint16 // total number of entries in the central directory on this disk
	numberOfEntries          uint16 // total number of entries in the central directory
	sizeOfCentralDirectories uint32 // size of the central directory block
	offsetCentralDirectory   uint32 // offset of start of central directory with respect to the starting disk number
	comment                  []byte // zip archive comment
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

	var (
		commentSize uint16
	)
	rr := bytes.NewReader(buf[:])
	byteio.GetUint16LE(rr, &e.numberOfDisk)
	byteio.GetUint16LE(rr, &e.numberOfStartDirDisk)
	byteio.GetUint16LE(rr, &e.numberOfEntriesThisDisk)
	byteio.GetUint16LE(rr, &e.numberOfEntries)
	byteio.GetUint32LE(rr, &e.sizeOfCentralDirectories)
	byteio.GetUint32LE(rr, &e.offsetCentralDirectory)
	byteio.GetUint16LE(rr, &commentSize)

	if e.numberOfDisk != 0 || e.numberOfStartDirDisk != 0 || e.numberOfEntriesThisDisk != e.numberOfEntries {
		return 0, errors.New("unsupport split zip file")
	}

	e.comment = make([]byte, commentSize)
	if commentSize != 0 {
		if _, err := r.Read(e.comment); err != nil {
			return 0, err
		}
	}

	size := int(sizeEndCentralDirectory) + int(commentSize)
	return int64(size), nil
}

// WriteTo writes an end of central directory record to io.Writer.
func (e *endCentralDirectory) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	w.Write([]byte(signEndCentralDirectory))
	byteio.WriteUint16LE(buf, e.numberOfDisk)
	byteio.WriteUint16LE(buf, e.numberOfStartDirDisk)
	byteio.WriteUint16LE(buf, e.numberOfEntriesThisDisk)
	byteio.WriteUint16LE(buf, e.numberOfEntries)
	byteio.WriteUint32LE(buf, e.sizeOfCentralDirectories)
	byteio.WriteUint32LE(buf, e.offsetCentralDirectory)
	byteio.WriteUint16LE(buf, uint16(len(e.comment)))
	buf.Write(e.comment)

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// uint16ToDosTime converts a date/time in uint16 to a MS-DOS time.
func uint16ToDosTime(dates uint16, times uint16) time.Time {
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

// uint16FromDosTime converts a MS-DOS time to a date/time in uint16.
func uint16FromDosTime(t time.Time) (dates uint16, times uint16) {
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
