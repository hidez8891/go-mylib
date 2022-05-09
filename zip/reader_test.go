package zip

import (
	"bytes"
	"io"
	"os"
	"testing"
)

type readerTest struct {
	file    string
	content string
}

var tests = []readerTest{
	{
		file:    "test-dd.zip",
		content: "Hello World!",
	},
	{
		file:    "test-nodd.zip",
		content: "Hello World!",
	},
	{
		file:    "test-dd-nosign.zip",
		content: "Hello World!",
	},
}

func TestReader(t *testing.T) {
	for _, tt := range tests {
		testReader(t, tt)
	}
}

func testReader(t *testing.T, tt readerTest) {
	r, err := os.Open("tests/" + tt.file)
	if err != nil {
		t.Fatalf("%s: %v", tt.file, err)
	}
	defer r.Close()

	zr, err := NewReader(r)
	if err != nil {
		t.Fatalf("%s: %v", tt.file, err)
	}

	if len(zr.Files) != 1 {
		t.Fatalf("%s: zf.Files size=%d, want=%d", tt.file, len(zr.Files), 1)
	}

	fr, err := zr.Files[0].Open()
	if err != nil {
		t.Fatalf("%s: %v", tt.file, err)
	}
	defer fr.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, fr); err != nil {
		t.Fatalf("%s: %v", tt.file, err)
	}

	content := buf.String()
	if content != tt.content {
		t.Fatalf("%s: zf.Files[0] content=%q, want=%q", tt.file, content, tt.content)
	}
}
