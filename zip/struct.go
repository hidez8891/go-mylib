package zip

import (
	"bytes"
	"errors"
	"io"
	"time"

	"go-mylib/byteio"
)

const (
	signLocalFileHeader        string = "PK\x03\x04"
	signCentralDirectoryHeader string = "PK\x01\x02"
	signDataDescriptor         string = "PK\x07\x08"
	signEndCentralDirectory    string = "PK\x05\x06"

	sizeLocalFileHeader        int = 30
	sizeCentralDirectoryHeader int = 46
	sizeDataDescriptor         int = 16
	sizeEndCentralDirectory    int = 22
)

const (
	FlagDataDescriptor uint16 = 0x0008
	FlagUTF8           uint16 = 0x0800
)

const (
	MethodStored   uint16 = 0
	MethodDeflated uint16 = 8
)

type localFileHeader struct {
	RequireVersion   uint16
	Flags            uint16
	Method           uint16
	ModifiedTime     time.Time
	CRC32            uint32
	CompressedSize   uint32
	UncompressedSize uint32
	FileName         string
	ExtraFields      []byte
}

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

	rr := bytes.NewReader(data[:])
	byteio.GetUint16LE(rr, &h.RequireVersion)
	byteio.GetUint16LE(rr, &h.Flags)
	byteio.GetUint16LE(rr, &h.Method)
	var modtime, moddate uint16
	byteio.GetUint16LE(rr, &modtime)
	byteio.GetUint16LE(rr, &moddate)
	h.ModifiedTime = uint32ToUTCTime(moddate, modtime)
	byteio.GetUint32LE(rr, &h.CRC32)
	byteio.GetUint32LE(rr, &h.CompressedSize)
	byteio.GetUint32LE(rr, &h.UncompressedSize)
	var nameSize, extraSize uint16
	byteio.GetUint16LE(rr, &nameSize)
	byteio.GetUint16LE(rr, &extraSize)

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

func (h *localFileHeader) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	buf.Write([]byte(signLocalFileHeader))
	byteio.WriteUint16LE(buf, h.RequireVersion)
	byteio.WriteUint16LE(buf, h.Flags)
	byteio.WriteUint16LE(buf, h.Method)
	moddate, modtime := utcTimeToUint32(h.ModifiedTime)
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

type centralDirectoryHeader struct {
	localFileHeader
	GenerateVersion   uint16
	InternalFileAttr  uint16
	ExternalFileAttr  uint32
	LocalHeaderOffset uint32
	Comment           string
}

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

	rr := bytes.NewReader(data[:])
	byteio.GetUint16LE(rr, &h.GenerateVersion)
	byteio.GetUint16LE(rr, &h.RequireVersion)
	byteio.GetUint16LE(rr, &h.Flags)
	byteio.GetUint16LE(rr, &h.Method)
	var modtime, moddate uint16
	byteio.GetUint16LE(rr, &modtime)
	byteio.GetUint16LE(rr, &moddate)
	h.ModifiedTime = uint32ToUTCTime(moddate, modtime)
	byteio.GetUint32LE(rr, &h.CRC32)
	byteio.GetUint32LE(rr, &h.CompressedSize)
	byteio.GetUint32LE(rr, &h.UncompressedSize)
	var nameSize, extraSize, commentSize uint16
	byteio.GetUint16LE(rr, &nameSize)
	byteio.GetUint16LE(rr, &extraSize)
	byteio.GetUint16LE(rr, &commentSize)
	var diskNumber uint16
	byteio.GetUint16LE(rr, &diskNumber)
	byteio.GetUint16LE(rr, &h.InternalFileAttr)
	byteio.GetUint32LE(rr, &h.ExternalFileAttr)
	byteio.GetUint32LE(rr, &h.LocalHeaderOffset)

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

func (h *centralDirectoryHeader) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)
	w.Write([]byte(signCentralDirectoryHeader))
	byteio.WriteUint16LE(buf, h.GenerateVersion)
	byteio.WriteUint16LE(buf, h.RequireVersion)
	byteio.WriteUint16LE(buf, h.Flags)
	byteio.WriteUint16LE(buf, h.Method)
	moddate, modtime := utcTimeToUint32(h.ModifiedTime)
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

type dataDescriptor struct {
	CRC32            uint32
	CompressedSize   uint32
	UncompressedSize uint32
}

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

type endCentralDirectory struct {
	numberOfEntries          uint16
	sizeOfCentralDirectories uint32
	offsetCentralDirectory   uint32
	Comment                  string
}

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
