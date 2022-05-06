package byteio

import (
	"encoding/binary"
	"errors"
	"io"
)

var errInsufficientReadSize = errors.New("insufficient read size")

// ReadUint8 returns uint8 number read from the Reader.
func ReadUint8(r io.Reader) (uint8, error) {
	buf := make([]byte, 1)
	rn, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if rn != len(buf) {
		return 0, errInsufficientReadSize
	}
	n := uint8(buf[0])
	return n, nil
}

// ReadUint16LE returns little-endian's uint16 number read from the Reader.
func ReadUint16LE(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	rn, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if rn != len(buf) {
		return 0, errInsufficientReadSize
	}
	n := binary.LittleEndian.Uint16(buf)
	return n, nil
}

// ReadUint16BE returns big-endian's uint16 number read from the Reader.
func ReadUint16BE(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	rn, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if rn != len(buf) {
		return 0, errInsufficientReadSize
	}
	n := binary.BigEndian.Uint16(buf)
	return n, nil
}

// ReadUint32LE returns little-endian's uint32 number read from the Reader.
func ReadUint32LE(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	rn, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if rn != len(buf) {
		return 0, errInsufficientReadSize
	}
	n := binary.LittleEndian.Uint32(buf)
	return n, nil
}

// ReadUint32BE returns big-endian's uint32 number read from the Reader.
func ReadUint32BE(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	rn, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if rn != len(buf) {
		return 0, errInsufficientReadSize
	}
	n := binary.BigEndian.Uint32(buf)
	return n, nil
}

// ReadUint64LE returns little-endian's uint64 number read from the Reader.
func ReadUint64LE(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	rn, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if rn != len(buf) {
		return 0, errInsufficientReadSize
	}
	n := binary.LittleEndian.Uint64(buf)
	return n, nil
}

// ReadUint64BE returns big-endian's uint64 number read from the Reader.
func ReadUint64BE(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	rn, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	if rn != len(buf) {
		return 0, errInsufficientReadSize
	}
	n := binary.BigEndian.Uint64(buf)
	return n, nil
}

// GetUint8 gets uint8 number read from the Reader.
func GetUint8(r io.Reader, v *uint8) (err error) {
	*v, err = ReadUint8(r)
	return err
}

// GetUint16LE gets uint8 number read from the Reader.
func GetUint16LE(r io.Reader, v *uint16) (err error) {
	*v, err = ReadUint16LE(r)
	return err
}

// GetUint16BE gets uint8 number read from the Reader.
func GetUint16BE(r io.Reader, v *uint16) (err error) {
	*v, err = ReadUint16BE(r)
	return err
}

// GetUint32LE gets uint8 number read from the Reader.
func GetUint32LE(r io.Reader, v *uint32) (err error) {
	*v, err = ReadUint32LE(r)
	return err
}

// GetUint32BE gets uint8 number read from the Reader.
func GetUint32BE(r io.Reader, v *uint32) (err error) {
	*v, err = ReadUint32BE(r)
	return err
}

// GetUint64LE gets uint8 number read from the Reader.
func GetUint64LE(r io.Reader, v *uint64) (err error) {
	*v, err = ReadUint64LE(r)
	return err
}

// GetUint64BE gets uint8 number read from the Reader.
func GetUint64BE(r io.Reader, v *uint64) (err error) {
	*v, err = ReadUint64BE(r)
	return err
}
