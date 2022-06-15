package zip

import (
	"fmt"
	"go-mylib/buffer"
	"io"
	"log"
	"os"
	"time"
)

func ExampleReader() {
	// Open a zip archive for reading
	r, err := os.Open("tests/test-comment.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Open a new zip reader
	zr, err := NewReader(r)
	if err != nil {
		log.Fatal(err)
	}

	// Show contents of the zip archive
	for _, f := range zr.Files {
		fmt.Printf("%s:\n", f.FileName)

		fr, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(os.Stdout, fr); err != nil {
			log.Fatal(err)
		}
		fr.Close()
		fmt.Println()
	}
	// Output:
	// test.txt:
	// Hello World!
}

func ExampleWriter() {
	// Create a buffer (io.WriteSeeker)
	bw := buffer.NewWriter(new(buffer.Buffer))

	// Create a new zip writer
	zw, err := NewWriter(bw)
	if err != nil {
		log.Fatal(err)
	}

	// Add a file to the zip archive
	var files = []struct {
		name string
		body string
	}{
		{"file1.txt", "file1 content"},
		{"file2.txt", "file2 content"},
	}
	for _, file := range files {
		f, err := zw.Create(file.name)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write([]byte(file.body)); err != nil {
			log.Fatal(err)
		}

		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}

	// Close the zip writer
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
}

func ExampleWriter_header() {
	// Create a buffer (io.WriteSeeker)
	bw := buffer.NewWriter(new(buffer.Buffer))

	// Create a new zip writer
	zw, err := NewWriter(bw)
	if err != nil {
		log.Fatal(err)
	}

	// Add a file to the zip archive
	var files = []struct {
		name  string
		body  string
		mtime time.Time
		atime time.Time
		ctime time.Time
	}{
		{"file1.txt", "file1 content", time.Now(), time.Now(), time.Now()},
		{"file2.txt", "file2 content", time.Now(), time.Now(), time.Now()},
	}
	for _, file := range files {
		header := NewFileHeader(file.name)
		header.ModifiedTime = file.mtime

		ntfsTime := &ExtraNTFS{
			Mtime: file.mtime,
			Atime: file.atime,
			Ctime: file.ctime,
		}
		header.ExtraFields = append(header.ExtraFields, ntfsTime)

		f, err := zw.CreateFromHeader(header)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write([]byte(file.body)); err != nil {
			log.Fatal(err)
		}

		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}

	// Close the zip writer
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
}

func ExampleWriter_copy() {
	// Open a zip archive for reading
	r, err := os.Open("tests/test-comment.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	zr, err := NewReader(r)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new zip writer
	bw := buffer.NewWriter(new(buffer.Buffer))
	zw, err := NewWriter(bw)
	if err != nil {
		log.Fatal(err)
	}

	// update a file name in the zip archive
	var updates = []struct {
		oldname string
		newname string
	}{
		{"file1.txt", "file2.txt"},
	}
	for _, file := range zr.Files {
		index := -1
		for i, update := range updates {
			if update.oldname != file.FileName {
				index = i
				break
			}
		}

		if index == -1 {
			// copy the file
			if err := zw.Copy(file); err != nil {
				log.Fatal(err)
			}
		} else {
			// update the file
			update := updates[index]

			header := file.FileHeader
			header.FileName = update.newname

			fr, err := file.Open()
			if err != nil {
				log.Fatal(err)
			}
			if err := zw.CopyFromReader(&header, fr); err != nil {
				log.Fatal(err)
			}
			if err := fr.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Close the zip writer
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}
}
