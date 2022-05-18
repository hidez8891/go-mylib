package zip

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

type testcase struct {
	path       string
	filename   string
	content    string
	comment    string
	flags      uint16
	mtime      time.Time
	zipcomment string
}

var tests = map[string]testcase{
	"data-descriptor": {
		path:     "test-dd.zip",
		filename: "test.txt",
		content:  "Hello World!",
		flags:    FlagDataDescriptor,
		mtime:    time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
	},
	"no-data-descriptor": {
		path:     "test-nodd.zip",
		filename: "test.txt",
		content:  "Hello World!",
		flags:    0,
		mtime:    time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
	},
	"data-descriptor-without-sign": {
		path:     "test-dd-nosign.zip",
		filename: "test.txt",
		content:  "Hello World!",
		flags:    FlagDataDescriptor,
		mtime:    time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
	},
	"comment": {
		path:       "test-comment.zip",
		filename:   "test.txt",
		content:    "Hello World!",
		comment:    "file comment",
		flags:      0,
		mtime:      time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
		zipcomment: "zip comment",
	},
}

func TestReader(t *testing.T) {
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			testReader(t, tt)
		})
	}
}

func testReader(t *testing.T, tt testcase) {
	r, err := os.Open("tests/" + tt.path)
	if err != nil {
		t.Fatalf("os.Open error=%v", err)
	}
	defer r.Close()

	testcaseCompare(t, r, tt)
}

func testcaseCompare(t *testing.T, r io.ReadSeeker, tt testcase) {
	zr, err := NewReader(r)
	if err != nil {
		t.Fatalf("NewReader error=%v", err)
	}

	if len(zr.Files) != 1 {
		t.Fatalf("Reader.Files size=%d, want=%d", len(zr.Files), 1)
	}
	if zr.Comment != tt.zipcomment {
		t.Errorf("zip comment get %q, want %q", zr.Comment, tt.zipcomment)
	}

	f := zr.Files[0]
	if f.FileName != tt.filename {
		t.Errorf("Filename get %q, want %q", f.FileName, tt.filename)
	}
	if f.ModifiedTime != tt.mtime {
		t.Errorf("ModifiedTime get %v, want %v", f.ModifiedTime, tt.mtime)
	}
	if f.Flags != tt.flags {
		t.Errorf("Flags get %x, want %x", f.Flags, tt.flags)
	}
	if f.Comment != tt.comment {
		t.Errorf("Comment get %q, want %q", f.Comment, tt.comment)
	}

	fr, err := f.Open()
	if err != nil {
		t.Fatalf("%s: %v", tt.path, err)
	}
	defer fr.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, fr); err != nil {
		t.Fatalf("%s: %v", tt.path, err)
	}

	content := buf.String()
	if content != tt.content {
		t.Fatalf("%s: zf.Files[0] content=%q, want=%q", tt.path, content, tt.content)
	}
}
