package zip

import (
	"go-mylib/buffer"
	"io/ioutil"
	"os"
	"testing"
)

func TestWriter(t *testing.T) {
	var wtests = []string{
		"data-descriptor",
		"no-data-descriptor",
		"comment",
		"empty-mtime",
	}

	for _, name := range wtests {
		tt := tests[name]
		t.Run(name, func(t *testing.T) {
			testWriter(t, tt)
		})
	}
}

func testWriter(t *testing.T, tt testcase) {
	buf := new(buffer.Buffer)

	zw, err := NewWriter(buffer.NewWriter(buf))
	if err != nil {
		t.Fatalf("NewWriter error=%#v", err)
	}
	zw.Comment = tt.zipcomment

	fh := NewFileHeader(tt.filename)
	fh.Flags = tt.flags
	fh.ModifiedTime = tt.mtime
	fh.Comment = tt.comment

	fw, err := zw.CreateFromHeader(fh)
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

	// Reader read test
	r := buffer.NewReader(buf)
	testcaseCompare(t, r, tt)

	// binary compare
	bin1 := buf.Bytes()
	bin2, err := ioutil.ReadFile("tests/" + tt.path)
	if err != nil {
		t.Fatalf("ioutil.ReadFile error=%#v", err)
	}
	if len(bin1) != len(bin2) {
		t.Fatalf("File written size=%d, want=%d", len(bin1), len(bin2))
	}
	for i := 0; i < len(bin1); i++ {
		if bin1[i] != bin2[i] {
			t.Fatalf("File body[%0x]=%0x, want=%0x", i, bin1[i], bin2[i])
		}
	}
}

func TestWriterAutoClose(t *testing.T) {
	tt := testcase{
		filename: "test.txt",
		content:  "Hello World",
	}

	buf := new(buffer.Buffer)
	zw, err := NewWriter(buffer.NewWriter(buf))
	if err != nil {
		t.Fatalf("NewWriter error=%#v", err)
	}

	fw, err := zw.Create(tt.filename)
	if err != nil {
		t.Fatalf("Writer.Create error=%#v", err)
	}

	if _, err := fw.Write([]byte(tt.content)); err != nil {
		t.Fatalf("FileWriter.Write error=%#v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("Writer.Close error=%#v", err)
	}

	// Reader read test
	r := buffer.NewReader(buf)
	testcaseCompare(t, r, tt)
}

func TestWriterCopy(t *testing.T) {
	var wtests = []string{
		"data-descriptor",
		"no-data-descriptor",
		"comment",
	}

	for _, name := range wtests {
		tt := tests[name]
		t.Run(name, func(t *testing.T) {
			// open zip
			r, err := os.Open("tests/" + tt.path)
			if err != nil {
				t.Fatalf("os.Open error=%#v", err)
			}
			defer r.Close()
			zr, err := NewReader(r)
			if err != nil {
				t.Fatalf("NewReader error=%#v", err)
			}

			// copy test
			buf := new(buffer.Buffer)
			zw, err := NewWriter(buffer.NewWriter(buf))
			if err != nil {
				t.Fatalf("NewWriter error=%#v", err)
			}
			for _, file := range zr.Files {
				if err := zw.Copy(file); err != nil {
					t.Errorf("Copy error=%#v", err)
				}
			}
			zw.Comment = zr.Comment
			if err := zw.Close(); err != nil {
				t.Fatalf("Writer.Close error=%#v", err)
			}

			// read test
			rr := buffer.NewReader(buf)
			testcaseCompare(t, rr, tt)
		})
	}
}

func TestWriterCopyFromReader(t *testing.T) {
	var wtests = []struct {
		src string
		dst string
	}{
		{
			src: "data-descriptor",
			dst: "no-data-descriptor",
		},
	}

	for _, test := range wtests {
		src := tests[test.src]
		dst := tests[test.dst]

		t.Run(test.src, func(t *testing.T) {
			// open file
			r, err := os.Open("tests/" + src.path)
			if err != nil {
				t.Fatalf("os.Open error=%#v", err)
			}
			defer r.Close()
			zr, err := NewReader(r)
			if err != nil {
				t.Fatalf("NewReader error=%#v", err)
			}

			// copy
			buf := new(buffer.Buffer)
			zw, err := NewWriter(buffer.NewWriter(buf))
			if err != nil {
				t.Fatalf("NewWriter error=%#v", err)
			}
			for _, file := range zr.Files {
				// change FileHeader
				fh := file.FileHeader
				fh.Flags.DataDescriptor = false

				// file copy
				fr, err := file.OpenRaw()
				if err != nil {
					t.Fatalf("File.OpenRaw error=%#v", err)
				}
				if err := zw.CopyFromReader(&fh, fr); err != nil {
					t.Errorf("Copy error=%#v", err)
				}
				fr.Close()
			}
			zw.Comment = zr.Comment
			if err := zw.Close(); err != nil {
				t.Fatalf("Writer.Close error=%#v", err)
			}

			// read test
			rr := buffer.NewReader(buf)
			testcaseCompare(t, rr, dst)
		})
	}
}
