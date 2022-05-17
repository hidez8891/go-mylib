package zip

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

type readerTest struct {
	path     string
	filename string
	content  string
	flags    uint16
	mtime    time.Time
}

var tests = map[string]readerTest{
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
}

func TestReader(t *testing.T) {
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			testReader(t, tt)
		})
	}
}

func testReader(t *testing.T, tt readerTest) {
	r, err := os.Open("tests/" + tt.path)
	if err != nil {
		t.Fatalf("os.Open error=%v", err)
	}
	defer r.Close()

	zr, err := NewReader(r)
	if err != nil {
		t.Fatalf("NewReader error=%v", err)
	}

	if len(zr.Files) != 1 {
		t.Fatalf("zip.Reader.Files size=%d, want=%d", len(zr.Files), 1)
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
