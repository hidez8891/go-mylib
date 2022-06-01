package zip

import (
	"bytes"
	"testing"
	"time"
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
	src := hexToBytes(`
		50 4b 03 04 14 00 08 00 08 00 5c 64 a6 54 04 03
		02 01 78 56 34 12 09 ef cd ab 08 00 04 00 66 69
		6c 65 6e 61 6d 65 04 03 02 01
	`)
	expect := localFileHeader{
		RequireVersion:   0x0014,
		Flags:            FlagType{DataDescriptor: true},
		Method:           &MethodDeflated{DefaultCompression},
		ModifiedTime:     time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
		CRC32:            0x01020304,
		CompressedSize:   0x12345678,
		UncompressedSize: 0xabcdef09,
		FileName:         "filename",
		ExtraFields:      hexToBytes("04 03 02 01"),
	}

	{
		r := bytes.NewReader(src)

		h := new(localFileHeader)
		if _, err := h.ReadFrom(r); err != nil {
			t.Fatal(err)
		}

		if h.RequireVersion != expect.RequireVersion {
			t.Errorf("RequireVersion=%x, want=%x", h.RequireVersion, expect.RequireVersion)
		}
		if h.Flags.get() != expect.Flags.get() {
			t.Errorf("Flags=%#v, want=%#v", h.Flags, expect.Flags)
		}
		if h.Method.ID() != expect.Method.ID() || h.Method.get() != expect.Method.get() {
			t.Errorf("Method=%#v, want=%#v", h.Method, expect.Method)
		}
		if !h.ModifiedTime.Equal(expect.ModifiedTime) {
			t.Errorf("ModifiedTime=%#v, want=%#v", h.ModifiedTime, expect.ModifiedTime)
		}
		if h.CRC32 != expect.CRC32 {
			t.Errorf("CRC32=%x, want=%x", h.CRC32, expect.CRC32)
		}
		if h.CompressedSize != expect.CompressedSize {
			t.Errorf("CompressedSize=%d, want=%d", h.CompressedSize, expect.CompressedSize)
		}
		if h.UncompressedSize != expect.UncompressedSize {
			t.Errorf("UncompressedSize=%d, want=%d", h.UncompressedSize, expect.UncompressedSize)
		}
		if h.FileName != expect.FileName {
			t.Errorf("FileName=%q, want=%q", h.FileName, expect.FileName)
		}

		if len(h.ExtraFields) != len(expect.ExtraFields) {
			t.Fatalf("ExtraFields size=%d, want=%d", len(h.ExtraFields), len(expect.ExtraFields))
		}
		for i := range h.ExtraFields {
			if h.ExtraFields[i] != expect.ExtraFields[i] {
				t.Fatalf("ExtraFields[%d]=%x, want=%x", i, h.ExtraFields[i], expect.ExtraFields[i])
			}
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

func Test_centralDirectoryHeader(t *testing.T) {
	src := hexToBytes(`
        50 4b 01 02 20 00 14 00 08 00 08 00 5c 64 a6 54
		04 03 02 01 78 56 34 12 09 ef cd ab 08 00 04 00
		07 00 00 00 01 00 03 00 00 00 b9 0f 00 00 66 69
		6c 65 6e 61 6d 65 04 03 02 01 63 6F 6D 6D 65 6E
		74
	`)
	expect := centralDirectoryHeader{
		localFileHeader: localFileHeader{
			RequireVersion:   0x0014,
			Flags:            FlagType{DataDescriptor: true},
			Method:           &MethodDeflated{DefaultCompression},
			ModifiedTime:     time.Date(2022, time.Month(5), 6, 12, 34, 56, 0, time.UTC),
			CRC32:            0x01020304,
			CompressedSize:   0x12345678,
			UncompressedSize: 0xabcdef09,
			FileName:         "filename",
			ExtraFields:      hexToBytes("04 03 02 01"),
		},
		GenerateVersion:   0x0020,
		InternalFileAttr:  0x0001,
		ExternalFileAttr:  0x0003,
		LocalHeaderOffset: 0x0fb9,
		Comment:           "comment",
	}

	{
		r := bytes.NewReader(src)

		h := new(centralDirectoryHeader)
		if _, err := h.ReadFrom(r); err != nil {
			t.Fatal(err)
		}

		if h.GenerateVersion != expect.GenerateVersion {
			t.Errorf("GenerateVersion=%x, want=%x", h.GenerateVersion, expect.GenerateVersion)
		}
		if h.RequireVersion != expect.RequireVersion {
			t.Errorf("RequireVersion=%x, want=%x", h.RequireVersion, expect.RequireVersion)
		}
		if h.Flags.get() != expect.Flags.get() {
			t.Errorf("Flags=%v, want=%v", h.Flags, expect.Flags)
		}
		if h.Method.ID() != expect.Method.ID() || h.Method.get() != expect.Method.get() {
			t.Errorf("Method=%v, want=%v", h.Method, expect.Method)
		}
		if !h.ModifiedTime.Equal(expect.ModifiedTime) {
			t.Errorf("ModifiedTime=%#v, want=%#v", h.ModifiedTime, expect.ModifiedTime)
		}
		if h.CRC32 != expect.CRC32 {
			t.Errorf("CRC32=%x, want=%x", h.CRC32, expect.CRC32)
		}
		if h.CompressedSize != expect.CompressedSize {
			t.Errorf("CompressedSize=%d, want=%d", h.CompressedSize, expect.CompressedSize)
		}
		if h.UncompressedSize != expect.UncompressedSize {
			t.Errorf("UncompressedSize=%d, want=%d", h.UncompressedSize, expect.UncompressedSize)
		}
		if h.FileName != expect.FileName {
			t.Errorf("FileName=%q, want=%q", h.FileName, expect.FileName)
		}

		if len(h.ExtraFields) != len(expect.ExtraFields) {
			t.Fatalf("ExtraFields size=%d, want=%d", len(h.ExtraFields), len(expect.ExtraFields))
		}
		for i := range h.ExtraFields {
			if h.ExtraFields[i] != expect.ExtraFields[i] {
				t.Fatalf("ExtraFields[%d]=%x, want=%x", i, h.ExtraFields[i], expect.ExtraFields[i])
			}
		}

		if h.Comment != expect.Comment {
			t.Errorf("Comment=%q, want=%q", h.Comment, expect.Comment)
		}
		if h.InternalFileAttr != expect.InternalFileAttr {
			t.Errorf("InternalFileAttr=%x, want=%x", h.InternalFileAttr, expect.InternalFileAttr)
		}
		if h.ExternalFileAttr != expect.ExternalFileAttr {
			t.Errorf("ExternalFileAttr=%x, want=%x", h.ExternalFileAttr, expect.ExternalFileAttr)
		}
		if h.LocalHeaderOffset != expect.LocalHeaderOffset {
			t.Errorf("LocalHeaderOffset=%x, want=%x", h.LocalHeaderOffset, expect.LocalHeaderOffset)
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
				CRC32:            0x01020304,
				CompressedSize:   0x1f2f3f4f,
				UncompressedSize: 0x5f6f7f8f,
			},
		},
		{
			src: hexToBytes(`
				04 03 02 01 4f 3f 2f 1f 8f 7f 6f 5f
			`),
			expect: dataDescriptor{
				CRC32:            0x01020304,
				CompressedSize:   0x1f2f3f4f,
				UncompressedSize: 0x5f6f7f8f,
			},
		},
	}

	for _, tt := range tests {
		{
			r := bytes.NewReader(tt.src)
			d := new(dataDescriptor)

			if _, err := d.ReadFrom(r); err != nil {
				t.Fatal(err)
			}
			if d.CRC32 != tt.expect.CRC32 {
				t.Errorf("CRC32=%x, want=%x", d.CRC32, tt.expect.CRC32)
			}
			if d.CompressedSize != tt.expect.CompressedSize {
				t.Errorf("CompressedSize=%d, want=%d", d.CompressedSize, tt.expect.CompressedSize)
			}
			if d.UncompressedSize != tt.expect.UncompressedSize {
				t.Errorf("UncompressedSize=%d, want=%d", d.UncompressedSize, tt.expect.UncompressedSize)
			}
		}

		if len(tt.src) != sizeDataDescriptor {
			continue // WriteTo test skip
		}

		{
			w := new(bytes.Buffer)
			if _, err := tt.expect.WriteTo(w); err != nil {
				t.Fatal(err)
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
		numberOfEntries:          10,
		sizeOfCentralDirectories: 0x5678,
		offsetCentralDirectory:   0x1020,
		Comment:                  "comment",
	}

	{
		r := bytes.NewReader(src)

		h := new(endCentralDirectory)
		if _, err := h.ReadFrom(r); err != nil {
			t.Fatal(err)
		}

		if h.numberOfEntries != expect.numberOfEntries {
			t.Errorf("numberOfEntries=%d, want=%d", h.numberOfEntries, expect.numberOfEntries)
		}
		if h.sizeOfCentralDirectories != expect.sizeOfCentralDirectories {
			t.Errorf("sizeOfCentralDirectories=%d, want=%d", h.sizeOfCentralDirectories, expect.sizeOfCentralDirectories)
		}
		if h.offsetCentralDirectory != expect.offsetCentralDirectory {
			t.Errorf("offsetCentralDirectory=%d, want=%d", h.offsetCentralDirectory, expect.offsetCentralDirectory)
		}
		if h.Comment != expect.Comment {
			t.Errorf("Comment=%q, want=%q", h.Comment, expect.Comment)
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
