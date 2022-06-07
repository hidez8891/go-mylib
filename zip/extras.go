package zip

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"go-mylib/byteio"
)

// ExtraFields represents a extra field interface.
type ExtraField interface {
	Tag() uint16
	ReadFrom(r io.Reader) (int64, error)
	WriteTo(w io.Writer) (int64, error)
}

// parseExtraFields parses the extra field list from the buffer.
func parseExtraFields(buf []byte) ([]ExtraField, error) {
	extraFields := make([]ExtraField, 0)
	size := 0

	r := bytes.NewReader(buf)
	for size < len(buf) {
		e := &ExtraUnknown{}
		if _, err := e.ReadFrom(r); err != nil {
			return nil, err
		}

		var extra ExtraField
		switch e.Tag() {
		case extraNTFSTag:
			extra = &ExtraNTFS{}
		default:
			extra = nil
		}

		if extra != nil {
			r2 := bytes.NewReader(e.Data)
			if _, err := extra.ReadFrom(r2); err != nil {
				return nil, err
			}
			extraFields = append(extraFields, extra)
		} else {
			extraFields = append(extraFields, e)
		}

		size += len(e.Data)
	}

	return extraFields, nil
}

// ExtraUnknown represents a unknown extra field.
type ExtraUnknown struct {
	tag  uint16
	Data []byte
}

// Tag returns the tag ID of the extra field.
func (e ExtraUnknown) Tag() uint16 {
	return e.tag
}

// ReadFrom reads the extra field from the reader.
func (e *ExtraUnknown) ReadFrom(r io.Reader) (int64, error) {
	e.Data = make([]byte, 4)
	if _, err := r.Read(e.Data); err != nil {
		return 0, err
	}

	var (
		tag  uint16
		size uint16
	)
	br := bytes.NewReader(e.Data)
	byteio.GetUint16LE(br, &tag)
	byteio.GetUint16LE(br, &size)
	e.tag = tag

	e.Data = append(e.Data, make([]byte, size)...)
	if _, err := r.Read(e.Data[4:]); err != nil {
		return 0, err
	}

	return 4 + int64(size), nil
}

// WriteTo writes the extra field to the writer.
func (e ExtraUnknown) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(e.Data)
	return int64(n), err
}

// extraNTFSTag is the tag ID of NTFS extra field.
const extraNTFSTag uint16 = 0x000a

// ExtraNTFS represents a extra field for NTFS.
type ExtraNTFS struct {
	Mtime time.Time // last modification time
	Atime time.Time // last access time
	Ctime time.Time // creation time
}

// Tag returns the tag ID of the extra field.
func (e ExtraNTFS) Tag() uint16 {
	return extraNTFSTag
}

// ReadFrom reads the extra field from the reader.
func (e *ExtraNTFS) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return 0, err
	}

	var (
		tag  uint16
		size uint16
	)
	br := bytes.NewReader(buf)
	byteio.GetUint16LE(br, &tag)
	byteio.GetUint16LE(br, &size)
	if tag != extraNTFSTag {
		return 0, errors.New("extra field is not NTFS")
	}

	buf = make([]byte, size)
	if _, err := r.Read(buf); err != nil {
		return 0, err
	}
	br = bytes.NewReader(buf)
	byteio.ReadUint32LE(br) // reserved area

	var (
		ttag  uint16
		tsize uint16
		mtime uint64
		atime uint64
		ctime uint64
	)
	byteio.GetUint16LE(br, &ttag)
	byteio.GetUint16LE(br, &tsize)
	if ttag != 0x0001 {
		return 0, fmt.Errorf("undefined NTFS attributes: %0x", ttag)
	}
	if tsize != 24 {
		return 0, fmt.Errorf("unexpected NTFS attributes size: %d", tsize)
	}
	byteio.GetUint64LE(br, &mtime)
	byteio.GetUint64LE(br, &atime)
	byteio.GetUint64LE(br, &ctime)

	e.Mtime = uint64ToWin32Time(mtime)
	e.Atime = uint64ToWin32Time(atime)
	e.Ctime = uint64ToWin32Time(ctime)

	return 4 + int64(size), nil
}

// WriteTo writes the extra field to the writer.
func (e ExtraNTFS) WriteTo(w io.Writer) (int64, error) {
	buf := new(bytes.Buffer)

	byteio.WriteUint16LE(buf, extraNTFSTag)
	byteio.WriteUint16LE(buf, 4+4+24)
	byteio.WriteUint32LE(buf, 0) // reserved
	byteio.WriteUint16LE(buf, 0x0001)
	byteio.WriteUint16LE(buf, 24)
	byteio.WriteUint64LE(buf, uint64FromWin32Time(e.Mtime))
	byteio.WriteUint64LE(buf, uint64FromWin32Time(e.Atime))
	byteio.WriteUint64LE(buf, uint64FromWin32Time(e.Ctime))

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// uint64ToWin32Time converts a uint64 to a win32 system time.
func uint64ToWin32Time(t uint64) time.Time {
	unixstart := uint64(0x019DB1DED53E8000)
	ticks := uint64(10000000)

	sec := int64((t - unixstart) / ticks)
	nsec := int64(t % ticks)
	return time.Unix(sec, nsec)
}

// uint64FromWin32Time converts a win32 system time to a uint64.
func uint64FromWin32Time(t time.Time) uint64 {
	unixstart := uint64(0x019DB1DED53E8000)
	ticks := uint64(10000000)

	utime := (uint64(t.Unix()) * ticks) + unixstart
	utime += uint64(t.Nanosecond())
	return utime
}
