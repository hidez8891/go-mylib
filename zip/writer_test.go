package zip

import (
	"go-mylib/buffer"
	"io/ioutil"
	"testing"
)

var wtests = []string{
	"data-descriptor",
	"no-data-descriptor",
	"comment",
	"empty-mtime",
}

func TestWriter(t *testing.T) {
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

	fw, err := zw.Create(tt.filename)
	if err != nil {
		t.Fatalf("Writer.Create error=%#v", err)
	}
	fw.Flags = tt.flags
	fw.ModifiedTime = tt.mtime
	fw.Comment = tt.comment

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
	fw.ModifiedTime = tt.mtime

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
