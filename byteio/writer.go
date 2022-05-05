package byteio

import (
	"encoding/binary"
	"errors"
	"io"
)

var errInsufficientWriteSize = errors.New("insufficient write size")

// WriteUint8 writes uint8 number to the Writer.
func WriteUint8(w io.Writer, v uint8) error {
	buf := make([]byte, 1)
	buf[0] = byte(v)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errInsufficientWriteSize
	}
	return nil
}

// WriteUint16LE writes little-endian's uint16 number to the Writer.
func WriteUint16LE(w io.Writer, v uint16) error {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, v)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errInsufficientWriteSize
	}
	return nil
}

// WriteUint16BE writes big-endian's uint16 number to the Writer.
func WriteUint16BE(w io.Writer, v uint16) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errInsufficientWriteSize
	}
	return nil
}

// WriteUint32LE writes little-endian's uint32 number to the Writer.
func WriteUint32LE(w io.Writer, v uint32) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errInsufficientWriteSize
	}
	return nil
}

// WriteUint32BE writes big-endian's uint32 number to the Writer.
func WriteUint32BE(w io.Writer, v uint32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errInsufficientWriteSize
	}
	return nil
}

// WriteUint64LE writes little-endian's uint64 number to the Writer.
func WriteUint64LE(w io.Writer, v uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errInsufficientWriteSize
	}
	return nil
}

// WriteUint64BE writes big-endian's uint64 number to the Writer.
func WriteUint64BE(w io.Writer, v uint64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return errInsufficientWriteSize
	}
	return nil
}
