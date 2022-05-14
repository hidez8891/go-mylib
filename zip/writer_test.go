package zip

import (
	"bytes"
	"go-mylib/buffer"
	"io"
	"testing"
)

type writerTest struct {
	content string
}

var wtests = []writerTest{
	{
		content: "Hello World!",
	},
}

func TestWriter(t *testing.T) {
	for _, tt := range wtests {
		testWriter(t, tt)
	}
}

func testWriter(t *testing.T, tt writerTest) {
	buf := new(buffer.Buffer)

	zw, err := NewWriter(buffer.NewWriter(buf))
	if err != nil {
		t.Fatalf("NewWriter error=%#v", err)
	}
	fw, err := zw.Create("test.txt")
	if err != nil {
		t.Fatalf("Writer.Create error=%#v", err)
	}
	if _, err := fw.Write([]byte(tt.content)); err != nil {
		t.Fatalf("FileWriter.Write error=%#v", err)
	}
	if err := fw.Close(); err != nil {
		t.Fatalf("FileWriter.Close error=%#v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("Writer.Close error=%#v", err)
	}

	zr, err := NewReader(buffer.NewReader(buf))
	if err != nil {
		t.Fatalf("NewReader error=%#v", err)
	}
	fr, err := zr.Files[0].Open()
	if err != nil {
		t.Fatalf("Reader.Open error=%#v", err)
	}
	tmp := new(bytes.Buffer)
	if _, err := io.Copy(tmp, fr); err != nil {
		t.Fatalf("io.ReadCloser.Read error=%#v", err)
	}
	if err := fr.Close(); err != nil {
		t.Fatalf("io.ReadCloser.Close error=%#v", err)
	}
	if tmp.String() != tt.content {
		t.Errorf("write test content=%q, want=%q", tmp.String(), tt.content)
	}
}
