package zip

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func hexTo(s rune) byte {
	if rune('0') <= s && s <= rune('9') {
		return byte(s - rune('0'))
	}
	if rune('a') <= s && s <= rune('f') {
		return byte(s - rune('a') + 10)
	}
	if rune('A') <= s && s <= rune('F') {
		return byte(s - rune('A') + 10)
	}
	panic("not hex rune")
}

func isHex(s rune) bool {
	if rune('0') <= s && s <= rune('9') {
		return true
	}
	if rune('a') <= s && s <= rune('f') {
		return true
	}
	if rune('A') <= s && s <= rune('F') {
		return true
	}
	return false
}

func hexToBytes(hex string) []byte {
	out := make([]byte, 0)

	for i := 0; i < len(hex); {
		s := rune(hex[i])
		if !isHex(s) {
			i += 1
			continue
		}
		b := hexTo(s)

		s = rune(hex[i+1])
		if !isHex(s) {
			panic("no hex string")
		}
		b = (b << 4) | hexTo(s)

		out = append(out, b)
		i += 2
	}

	return out
}

func Test_localFileHeader(t *testing.T) {
	tests := []struct {
		src    []byte
		dst    []byte
		expect *FileHeader
	}{
		{
			src: hexToBytes(`
				50 4b 03 04 14 00 00 00 08 00 5c 64 a6 54 04 03
				02 01 78 56 34 12 09 ef cd ab 08 00 06 00 66 69
				6c 65 6e 61 6d 65 ff ee 02 00 00 00
			`),
			dst: hexToBytes(`
				50 4b 03 04 14 00 00 00 08 00 5c 64 a6 54 04 03
				02 01 78 56 34 12 09 ef cd ab 08 00 00 00 66 69
				6c 65 6e 61 6d 65
			`),
			expect: &FileHeader{
				MinimumVersion:   0x0014,
				Flags:            FlagType{},
				Method:           &MethodDeflated{DefaultCompression},
				ModifiedTime:     time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
				CRC32:            0x01020304,
				CompressedSize:   0x12345678,
				UncompressedSize: 0xabcdef09,
				FileName:         "filename",
				ExtraFields: []ExtraField{
					&ExtraUnknown{
						tag:  0xeeff,
						Data: hexToBytes("ff ee 02 00 00 00"),
					},
				},
			},
		},
		{
			src: hexToBytes(`
				50 4b 03 04 14 00 08 00 08 00 5c 64 a6 54 04 03
				02 01 78 56 34 12 09 ef cd ab 08 00 06 00 66 69
				6c 65 6e 61 6d 65 ff ee 02 00 00 00
			`),
			dst: hexToBytes(`
				50 4b 03 04 14 00 08 00 08 00 5c 64 a6 54 00 00
				00 00 00 00 00 00 00 00 00 00 08 00 00 00 66 69
				6c 65 6e 61 6d 65
			`),
			expect: &FileHeader{
				MinimumVersion:   0x0014,
				Flags:            FlagType{DataDescriptor: true},
				Method:           &MethodDeflated{DefaultCompression},
				ModifiedTime:     time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
				CRC32:            0x01020304,
				CompressedSize:   0x12345678,
				UncompressedSize: 0xabcdef09,
				FileName:         "filename",
				ExtraFields: []ExtraField{
					&ExtraUnknown{
						tag:  0xeeff,
						Data: hexToBytes("ff ee 02 00 00 00"),
					},
				},
			},
		},
	}

	for _, test := range tests {
		expect := test.expect
		{
			r := bytes.NewReader(test.src)

			fh := new(localFileHeader)
			if _, err := fh.ReadFrom(r); err != nil {
				t.Fatalf("ReadFrom: %v", err)
			}

			h := new(FileHeader)
			if err := fh.copyToHeader(h); err != nil {
				t.Fatalf("copyToHeader: %v", err)
			}

			cmpopt := cmp.AllowUnexported(ExtraUnknown{})
			if diff := cmp.Diff(expect, h, cmpopt); diff != "" {
				t.Errorf("unexpected header (-want +got):\n%s", diff)
			}
			if len(h.ExtraFields) != len(expect.ExtraFields) {
				t.Fatalf("ExtraFields size=%d, want=%d", len(h.ExtraFields), len(expect.ExtraFields))
			}
		}
		{
			fh := new(localFileHeader)
			if err := fh.copyFromHeader(expect); err != nil {
				t.Fatalf("copyFromHeader: %v", err)
			}

			w := new(bytes.Buffer)
			if _, err := fh.WriteTo(w); err != nil {
				t.Fatal(err)
			}
			dst := w.Bytes()

			if len(dst) != len(test.dst) {
				t.Fatalf("WriteTo write size=%d, want=%d", len(dst), len(test.dst))
			}
			for i := range dst {
				if dst[i] != test.dst[i] {
					t.Fatalf("WriteTo result[%d]=%x, want=%x", i, dst[i], test.dst[i])
				}
			}
		}
	}
}

func Test_centralDirectoryHeader(t *testing.T) {
	src := hexToBytes(`
        50 4b 01 02 20 00 14 00 08 00 08 00 5c 64 a6 54
		04 03 02 01 78 56 34 12 09 ef cd ab 08 00 06 00
		07 00 00 00 01 00 03 00 00 00 b9 0f 00 00 66 69
		6c 65 6e 61 6d 65 ff ee 02 00 00 00 63 6F 6D 6D
		65 6E 74
	`)
	expect := &FileHeader{
		MinimumVersion:   0x0014,
		GenerateVersion:  0x20,
		GenerateOS:       OS_MSDOS,
		Flags:            FlagType{DataDescriptor: true},
		Method:           &MethodDeflated{DefaultCompression},
		ModifiedTime:     time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
		CRC32:            0x01020304,
		CompressedSize:   0x12345678,
		UncompressedSize: 0xabcdef09,
		FileName:         "filename",
		ExtraFields: []ExtraField{
			&ExtraUnknown{
				tag:  0xeeff,
				Data: hexToBytes("ff ee 02 00 00 00"),
			},
		},
		InternalFileAttr: 0x0001,
		ExternalFileAttr: 0x0003,
		Comment:          "comment",
	}
	offset := uint32(0x0fb9)

	{
		r := bytes.NewReader(src)

		dh := new(centralDirectoryHeader)
		if _, err := dh.ReadFrom(r); err != nil {
			t.Fatalf("ReadFrom: %v", err)
		}

		h := new(FileHeader)
		if err := dh.copyToHeader(h); err != nil {
			t.Fatalf("copyToHeader: %v", err)
		}

		cmpopt := cmp.AllowUnexported(ExtraUnknown{})
		if diff := cmp.Diff(expect, h, cmpopt); diff != "" {
			t.Errorf("unexpected header (-want +got):\n%s", diff)
		}
		if offset != dh.localHeaderOffset {
			t.Errorf("offset=%d, want=%d", dh.localHeaderOffset, offset)
		}
		if len(h.ExtraFields) != len(expect.ExtraFields) {
			t.Fatalf("ExtraFields size=%d, want=%d", len(h.ExtraFields), len(expect.ExtraFields))
		}
	}
	{
		dh := new(centralDirectoryHeader)
		if err := dh.copyFromHeader(expect); err != nil {
			t.Fatalf("copyToHeader: %v", err)
		}
		dh.localHeaderOffset = offset

		w := new(bytes.Buffer)
		if _, err := dh.WriteTo(w); err != nil {
			t.Fatalf("WriteTo: %v", err)
		}
		dst := w.Bytes()

		if len(dst) != len(src) {
			t.Fatalf("WriteTo write size=%d, want=%d", len(dst), len(src))
		}
		for i := range dst {
			if dst[i] != src[i] {
				t.Fatalf("WriteTo result[%d]=%x, want=%x", i, dst[i], src[i])
			}
		}
	}
}

func Test_dataDescriptor(t *testing.T) {
	tests := []struct {
		src    []byte
		expect dataDescriptor
	}{
		{
			src: hexToBytes(`
				50 4b 07 08 04 03 02 01 4f 3f 2f 1f 8f 7f 6f 5f
			`),
			expect: dataDescriptor{
				crc32:            0x01020304,
				compressedSize:   0x1f2f3f4f,
				uncompressedSize: 0x5f6f7f8f,
			},
		},
		{
			src: hexToBytes(`
				04 03 02 01 4f 3f 2f 1f 8f 7f 6f 5f
			`),
			expect: dataDescriptor{
				crc32:            0x01020304,
				compressedSize:   0x1f2f3f4f,
				uncompressedSize: 0x5f6f7f8f,
			},
		},
	}

	for _, tt := range tests {
		{
			r := bytes.NewReader(tt.src)
			d := new(dataDescriptor)

			if _, err := d.ReadFrom(r); err != nil {
				t.Fatalf("ReadFrom: %v", err)
			}

			cmpopt := cmp.AllowUnexported(dataDescriptor{})
			if diff := cmp.Diff(tt.expect, *d, cmpopt); diff != "" {
				t.Errorf("unexpected data descriptor (-want +got):\n%s", diff)
			}
		}

		if len(tt.src) != sizeDataDescriptor {
			continue // WriteTo test skip
		}

		{
			w := new(bytes.Buffer)
			if _, err := tt.expect.WriteTo(w); err != nil {
				t.Fatalf("WriteTo: %v", err)
			}
			dst := w.Bytes()

			if len(dst) != len(tt.src) {
				t.Fatalf("WriteTo write size=%d, want=%d", len(dst), len(tt.src))
			}
			for i := range dst {
				if dst[i] != tt.src[i] {
					t.Fatalf("WriteTo result[%d]=%x, want=%x", i, dst[i], tt.src[i])
				}
			}
		}
	}
}

func Test_endCentralDirectory(t *testing.T) {
	src := hexToBytes(`
        50 4B 05 06 00 00 00 00 0a 00 0a 00 78 56 00 00
		20 10 00 00 07 00 63 6F 6D 6D 65 6E 74
	`)
	expect := endCentralDirectory{
		numberOfDisk:             0,
		numberOfStartDirDisk:     0,
		numberOfEntriesThisDisk:  10,
		numberOfEntries:          10,
		sizeOfCentralDirectories: 0x5678,
		offsetCentralDirectory:   0x1020,
		comment:                  []byte("comment"),
	}

	{
		r := bytes.NewReader(src)

		h := new(endCentralDirectory)
		if _, err := h.ReadFrom(r); err != nil {
			t.Fatal(err)
		}

		cmpopt := cmp.AllowUnexported(endCentralDirectory{})
		if diff := cmp.Diff(expect, *h, cmpopt); diff != "" {
			t.Errorf("unexpected data descriptor (-want +got):\n%s", diff)
		}
	}

	{
		w := new(bytes.Buffer)
		if _, err := expect.WriteTo(w); err != nil {
			t.Fatal(err)
		}
		dst := w.Bytes()

		if len(dst) != len(src) {
			t.Fatalf("WriteTo write size=%d, want=%d", len(dst), len(src))
		}
		for i := range dst {
			if dst[i] != src[i] {
				t.Fatalf("WriteTo result[%d]=%x, want=%x", i, dst[i], src[i])
			}
		}
	}
}
