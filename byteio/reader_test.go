package byteio

import (
	"bytes"
	"testing"
)

func TestReadUint8(t *testing.T) {
	tests := []struct {
		source []byte
		expect uint8
	}{
		{
			[]byte("\x01"),
			0x01,
		},
		{
			[]byte("\xfe"),
			0xfe,
		},
	}

	for _, tt := range tests {
		n, err := ReadUint8(bytes.NewReader(tt.source))
		if err != nil {
			t.Fatal(err)
		}
		if n != tt.expect {
			t.Fatalf("ReadUint8 read=%d, want=%d", n, tt.expect)
		}
	}
}

func TestReadUint16LE(t *testing.T) {
	tests := []struct {
		source []byte
		expect uint16
	}{
		{
			[]byte("\x02\x01"),
			0x0102,
		},
		{
			[]byte("\xff\xfe"),
			0xfeff,
		},
	}

	for _, tt := range tests {
		n, err := ReadUint16LE(bytes.NewReader(tt.source))
		if err != nil {
			t.Fatal(err)
		}
		if n != tt.expect {
			t.Fatalf("ReadUint16LE read=%d, want=%d", n, tt.expect)
		}
	}
}

func TestReadUint16BE(t *testing.T) {
	tests := []struct {
		source []byte
		expect uint16
	}{
		{
			[]byte("\x02\x01"),
			0x0201,
		},
		{
			[]byte("\xff\xfe"),
			0xfffe,
		},
	}

	for _, tt := range tests {
		n, err := ReadUint16BE(bytes.NewReader(tt.source))
		if err != nil {
			t.Fatal(err)
		}
		if n != tt.expect {
			t.Fatalf("ReadUint16BE read=%d, want=%d", n, tt.expect)
		}
	}
}

func TestReadUint32LE(t *testing.T) {
	tests := []struct {
		source []byte
		expect uint32
	}{
		{
			[]byte("\x04\x03\x02\x01"),
			0x01020304,
		},
		{
			[]byte("\xff\xfe\xfd\xfc"),
			0xfcfdfeff,
		},
	}

	for _, tt := range tests {
		n, err := ReadUint32LE(bytes.NewReader(tt.source))
		if err != nil {
			t.Fatal(err)
		}
		if n != tt.expect {
			t.Fatalf("ReadUint32LE read=%d, want=%d", n, tt.expect)
		}
	}
}

func TestReadUint32BE(t *testing.T) {
	tests := []struct {
		source []byte
		expect uint32
	}{
		{
			[]byte("\x04\x03\x02\x01"),
			0x04030201,
		},
		{
			[]byte("\xff\xfe\xfd\xfc"),
			0xfffefdfc,
		},
	}

	for _, tt := range tests {
		n, err := ReadUint32BE(bytes.NewReader(tt.source))
		if err != nil {
			t.Fatal(err)
		}
		if n != tt.expect {
			t.Fatalf("ReadUint32BE read=%d, want=%d", n, tt.expect)
		}
	}
}

func TestReadUint64LE(t *testing.T) {
	tests := []struct {
		source []byte
		expect uint64
	}{
		{
			[]byte("\x08\x07\x06\x05\x04\x03\x02\x01"),
			0x0102030405060708,
		},
		{
			[]byte("\xff\xfe\xfd\xfc\xfb\xfa\xf9\xf8"),
			0xf8f9fafbfcfdfeff,
		},
	}

	for _, tt := range tests {
		n, err := ReadUint64LE(bytes.NewReader(tt.source))
		if err != nil {
			t.Fatal(err)
		}
		if n != tt.expect {
			t.Fatalf("ReadUint64LE read=%d, want=%d", n, tt.expect)
		}
	}
}

func TestReadUint64BE(t *testing.T) {
	tests := []struct {
		source []byte
		expect uint64
	}{
		{
			[]byte("\x08\x07\x06\x05\x04\x03\x02\x01"),
			0x0807060504030201,
		},
		{
			[]byte("\xff\xfe\xfd\xfc\xfb\xfa\xf9\xf8"),
			0xfffefdfcfbfaf9f8,
		},
	}

	for _, tt := range tests {
		n, err := ReadUint64BE(bytes.NewReader(tt.source))
		if err != nil {
			t.Fatal(err)
		}
		if n != tt.expect {
			t.Fatalf("ReadUint64BE read=%d, want=%d", n, tt.expect)
		}
	}
}
