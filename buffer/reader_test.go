package buffer

import (
	"io"
	"testing"
)

func TestReader(t *testing.T) {
	buf := new(Buffer)
	{
		w := NewWriter(buf)
		w.Write([]byte("Hello World"))
	}

	r := NewReader(buf)

	// io.Reader
	{
		expect := "Hello"
		data := make([]byte, len(expect))

		n, err := r.Read(data)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(expect) {
			t.Fatalf("Reader.Read read size=%d, want=%d", n, len(expect))
		}

		result := string(data)
		if result != expect {
			t.Fatalf("Reader.Read read %q, want %q", result, expect)
		}
	}
	{
		expect := " World"
		data := make([]byte, len(expect))

		n, err := r.Read(data)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(expect) {
			t.Fatalf("Reader.Read read size=%d, want=%d", n, len(expect))
		}

		result := string(data)
		if result != expect {
			t.Fatalf("Reader.Read read %q, want %q", result, expect)
		}
	}
	{
		data := make([]byte, 1)
		_, err := r.Read(data)
		if err != io.EOF {
			t.Fatalf("Reader.Read get %#v, want io.EOF", err)
		}
	}

	// io.ReaderAt
	{
		pos := len("Hello ")
		expect := "World"
		data := make([]byte, len(expect))

		n, err := r.ReadAt(data, int64(pos))
		if err != nil {
			t.Fatal(err)
		}
		if n != len(expect) {
			t.Fatalf("Reader.ReadAt read size=%d, want=%d", n, len(expect))
		}

		result := string(data)
		if result != expect {
			t.Fatalf("Reader.ReadAt read %q, want %q", result, expect)
		}
	}

	// io.ReadSeeker
	{
		pos := len("Hello ")
		expect := "World"
		data := make([]byte, len(expect))

		off, err := r.Seek(int64(pos), io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
		if off != int64(pos) {
			t.Fatalf("Reader.Seek move offset=%d, want=%d", off, pos)
		}

		n, err := r.Read(data)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(expect) {
			t.Fatalf("Reader.Read read size=%d, want=%d", n, len(expect))
		}

		result := string(data)
		if result != expect {
			t.Fatalf("Reader.Read read %q, want %q", result, expect)
		}
	}
}
