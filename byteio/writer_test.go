package byteio

import (
	"bytes"
	"testing"
)

func TestWriteUint8(t *testing.T) {
	tests := []struct {
		source uint8
		expect []byte
	}{
		{
			0x01,
			[]byte("\x01"),
		},
		{
			0xfe,
			[]byte("\xfe"),
		},
	}

	for _, tt := range tests {
		w := new(bytes.Buffer)
		err := WriteUint8(w, tt.source)
		if err != nil {
			t.Fatal(err)
		}
		if w.Len() != len(tt.expect) {
			t.Fatalf("WriteUint8 write size %d, want %d", w.Len(), len(tt.expect))
		}
		buf := w.Bytes()
		for i := range tt.expect {
			if buf[i] != tt.expect[i] {
				t.Fatalf("WriteUint8 write[%d]=%d, want=%d", buf[i], i, tt.expect[i])
			}
		}
	}
}

func TestWriteUint16LE(t *testing.T) {
	tests := []struct {
		source uint16
		expect []byte
	}{
		{
			0x0102,
			[]byte("\x02\x01"),
		},
		{
			0xfffe,
			[]byte("\xfe\xff"),
		},
	}

	for _, tt := range tests {
		w := new(bytes.Buffer)
		err := WriteUint16LE(w, tt.source)
		if err != nil {
			t.Fatal(err)
		}
		if w.Len() != len(tt.expect) {
			t.Fatalf("WriteUint16LE write size %d, want %d", w.Len(), len(tt.expect))
		}
		buf := w.Bytes()
		for i := range tt.expect {
			if buf[i] != tt.expect[i] {
				t.Fatalf("WriteUint16LE write[%d]=%d, want=%d", buf[i], i, tt.expect[i])
			}
		}
	}
}

func TestWriteUint16BE(t *testing.T) {
	tests := []struct {
		source uint16
		expect []byte
	}{
		{
			0x0102,
			[]byte("\x01\x02"),
		},
		{
			0xfffe,
			[]byte("\xff\xfe"),
		},
	}

	for _, tt := range tests {
		w := new(bytes.Buffer)
		err := WriteUint16BE(w, tt.source)
		if err != nil {
			t.Fatal(err)
		}
		if w.Len() != len(tt.expect) {
			t.Fatalf("WriteUint16BE write size %d, want %d", w.Len(), len(tt.expect))
		}
		buf := w.Bytes()
		for i := range tt.expect {
			if buf[i] != tt.expect[i] {
				t.Fatalf("WriteUint16BE write[%d]=%d, want=%d", buf[i], i, tt.expect[i])
			}
		}
	}
}

func TestWriteUint32LE(t *testing.T) {
	tests := []struct {
		source uint32
		expect []byte
	}{
		{
			0x01020304,
			[]byte("\x04\x03\x02\x01"),
		},
		{
			0xfffefdfc,
			[]byte("\xfc\xfd\xfe\xff"),
		},
	}

	for _, tt := range tests {
		w := new(bytes.Buffer)
		err := WriteUint32LE(w, tt.source)
		if err != nil {
			t.Fatal(err)
		}
		if w.Len() != len(tt.expect) {
			t.Fatalf("WriteUint32LE write size %d, want %d", w.Len(), len(tt.expect))
		}
		buf := w.Bytes()
		for i := range tt.expect {
			if buf[i] != tt.expect[i] {
				t.Fatalf("WriteUint32LE write[%d]=%d, want=%d", buf[i], i, tt.expect[i])
			}
		}
	}
}

func TestWriteUint32BE(t *testing.T) {
	tests := []struct {
		source uint32
		expect []byte
	}{
		{
			0x01020304,
			[]byte("\x01\x02\x03\x04"),
		},
		{
			0xfffefdfc,
			[]byte("\xff\xfe\xfd\xfc"),
		},
	}

	for _, tt := range tests {
		w := new(bytes.Buffer)
		err := WriteUint32BE(w, tt.source)
		if err != nil {
			t.Fatal(err)
		}
		if w.Len() != len(tt.expect) {
			t.Fatalf("WriteUint32BE write size %d, want %d", w.Len(), len(tt.expect))
		}
		buf := w.Bytes()
		for i := range tt.expect {
			if buf[i] != tt.expect[i] {
				t.Fatalf("WriteUint32BE write[%d]=%d, want=%d", buf[i], i, tt.expect[i])
			}
		}
	}
}

func TestWriteUint64LE(t *testing.T) {
	tests := []struct {
		source uint64
		expect []byte
	}{
		{
			0x0102030405060708,
			[]byte("\x08\x07\x06\x05\x04\x03\x02\x01"),
		},
		{
			0xfffefdfcfbfaf9f8,
			[]byte("\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"),
		},
	}

	for _, tt := range tests {
		w := new(bytes.Buffer)
		err := WriteUint64LE(w, tt.source)
		if err != nil {
			t.Fatal(err)
		}
		if w.Len() != len(tt.expect) {
			t.Fatalf("WriteUint64LE write size %d, want %d", w.Len(), len(tt.expect))
		}
		buf := w.Bytes()
		for i := range tt.expect {
			if buf[i] != tt.expect[i] {
				t.Fatalf("WriteUint64LE write[%d]=%d, want=%d", buf[i], i, tt.expect[i])
			}
		}
	}
}

func TestWriteUint64BE(t *testing.T) {
	tests := []struct {
		source uint64
		expect []byte
	}{
		{
			0x0102030405060708,
			[]byte("\x01\x02\x03\x04\x05\x06\x07\x08"),
		},
		{
			0xfffefdfcfbfaf9f8,
			[]byte("\xff\xfe\xfd\xfc\xfb\xfa\xf9\xf8"),
		},
	}

	for _, tt := range tests {
		w := new(bytes.Buffer)
		err := WriteUint64BE(w, tt.source)
		if err != nil {
			t.Fatal(err)
		}
		if w.Len() != len(tt.expect) {
			t.Fatalf("WriteUint64BE write size %d, want %d", w.Len(), len(tt.expect))
		}
		buf := w.Bytes()
		for i := range tt.expect {
			if buf[i] != tt.expect[i] {
				t.Fatalf("WriteUint64BE write[%d]=%d, want=%d", buf[i], i, tt.expect[i])
			}
		}
	}
}
