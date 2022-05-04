package buffer

import (
	"io"
	"testing"
)

func TestWriter(t *testing.T) {
	buf := new(Buffer)
	w := NewWriter(buf)

	// io.Writer
	{
		source := "Hello"
		expect := "Hello"

		n, err := w.Write([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		if n != len(source) {
			t.Fatalf("Writer.Write written size=%d, want=%d", n, len(source))
		}

		result := string(buf.Bytes())
		if result != expect {
			t.Fatalf("Writer.Write written %q, want %q", result, expect)
		}
	}
	{
		source := " World"
		expect := "Hello World"

		n, err := w.Write([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		if n != len(source) {
			t.Fatalf("Writer.Write written size=%d, want=%d", n, len(source))
		}

		result := string(buf.Bytes())
		if result != expect {
			t.Fatalf("Writer.Write written %q, want %q", result, expect)
		}
	}

	// io.WriterAt
	{
		pos := len("Hello ")
		source := "Golang"
		expect := "Hello Golang"

		n, err := w.WriteAt([]byte(source), int64(pos))
		if err != nil {
			t.Fatal(err)
		}
		if n != len(source) {
			t.Fatalf("Writer.WriteAt written size=%d, want=%d", n, len(source))
		}

		result := string(buf.Bytes())
		if result != expect {
			t.Fatalf("Writer.WriteAt written %q, want %q", result, expect)
		}
	}

	// io.WriteSeeker
	{
		pos := len("Hello ")
		source := "World!!"
		expect := "Hello World!!"

		off, err := w.Seek(int64(pos), io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
		if off != int64(pos) {
			t.Fatalf("Writer.Seek move offset=%d, want=%d", off, pos)
		}

		n, err := w.Write([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		if n != len(source) {
			t.Fatalf("Writer.Write written size=%d, want=%d", n, len(source))
		}

		result := string(buf.Bytes())
		if result != expect {
			t.Fatalf("Writer.Write written %q, want %q", result, expect)
		}
	}
}
