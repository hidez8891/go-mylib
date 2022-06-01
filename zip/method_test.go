package zip

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestMethodStore(t *testing.T) {
	expect := "Hello"

	tmp := new(bytes.Buffer)
	w := newStoreWriter(tmp, 0)
	if _, err := w.Write([]byte(expect)); err != nil {
		t.Fatalf("Store Write error=%#v", err)
	}
	w.Close()

	buf := new(bytes.Buffer)
	r := newStoreReader(bytes.NewReader(tmp.Bytes()), 0)
	if _, err := io.Copy(buf, r); err != nil {
		t.Fatalf("Store Read error=%#v", err)
	}
	r.Close()

	output := buf.String()
	if expect != output {
		t.Fatalf("Store Write/Read get=%q, want=%q", output, expect)
	}
}

func TestMethodDeflate(t *testing.T) {
	expect := "Hello"

	levels := []uint16{
		0x00 << 1,
		0x01 << 1,
		0x02 << 1,
		0x03 << 1,
	}

	for _, level := range levels {
		level := level
		t.Run(fmt.Sprintf("Level=%02x", level), func(t *testing.T) {
			tmp := new(bytes.Buffer)
			w := newDeflateWriter(tmp, level)
			if _, err := w.Write([]byte(expect)); err != nil {
				t.Fatalf("Deflate Write error=%#v", err)
			}
			w.Close()

			buf := new(bytes.Buffer)
			r := newDeflateReader(bytes.NewReader(tmp.Bytes()), level)
			if _, err := io.Copy(buf, r); err != nil {
				t.Fatalf("Deflate Read error=%#v", err)
			}
			r.Close()

			output := buf.String()
			if expect != output {
				t.Fatalf("Deflate Write/Read get=%q, want=%q", output, expect)
			}
		})
	}
}
