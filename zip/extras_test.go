package zip

import (
	"bytes"
	"testing"
	"time"
)

func TestExtraUnknown(t *testing.T) {
	tests := []struct {
		data []byte
		tag  uint16
	}{
		{
			data: hexToBytes(`
				ff fe 04 00 11 22 33 44
			`),
			tag: 0xfeff,
		},
	}

	for i, test := range tests {
		var e ExtraUnknown

		r := bytes.NewReader(test.data)
		if _, err := e.ReadFrom(r); err != nil {
			t.Fatalf("table#%d ReadFrom: %v", i, err)
		}
		if e.Tag() != test.tag {
			t.Errorf("table#%d Tag=%x, want=%x", i, e.Tag(), test.tag)
		}

		w := new(bytes.Buffer)
		if _, err := e.WriteTo(w); err != nil {
			t.Fatalf("table#%d WriteTo: %v", i, err)
		}
		if w.Len() != len(test.data) {
			t.Fatalf("table#%d WriteTo: written=%d, want=%d", i, w.Len(), len(test.data))
		}
		out := w.Bytes()
		for j := range test.data {
			if out[j] != test.data[j] {
				t.Fatalf("table#%d WriteTo: write[%d]=%x, want=%x", i, j, out[j], test.data[j])
			}
		}
	}
}

func TestExtraNTFS(t *testing.T) {
	tests := []struct {
		data  []byte
		mtime time.Time
		ctime time.Time
		atime time.Time
	}{
		{
			data: hexToBytes(`
				         0A 00 20 00 00-00 00 00 01 00 18 00 9B
				F8 4B B4 5E 7A D8 01 D3-5D 02 F1 5F 7A D8 01 9B
				F8 4B B4 5E 7A D8 01
			`),
			mtime: time.Date(2022, 6, 7, 11, 6, 57, 7821851, time.UTC),
			ctime: time.Date(2022, 6, 7, 11, 6, 57, 7821851, time.UTC),
			atime: time.Date(2022, 6, 7, 11, 15, 49, 1375571, time.UTC),
		},
	}

	for i, test := range tests {
		var e ExtraNTFS

		r := bytes.NewReader(test.data)
		if _, err := e.ReadFrom(r); err != nil {
			t.Fatalf("table#%d ReadFrom: %v", i, err)
		}
		if e.Mtime.Equal(test.mtime) == false {
			t.Errorf("table#%d Mtime=%v, want=%v", i, e.Mtime, test.mtime)
		}
		if e.Ctime.Equal(test.ctime) == false {
			t.Errorf("table#%d Ctime=%v, want=%v", i, e.Ctime, test.ctime)
		}
		if e.Atime.Equal(test.atime) == false {
			t.Errorf("table#%d Atime=%v, want=%v", i, e.Atime, test.atime)
		}

		w := new(bytes.Buffer)
		if _, err := e.WriteTo(w); err != nil {
			t.Fatalf("table#%d WriteTo: %v", i, err)
		}
		if w.Len() != len(test.data) {
			t.Fatalf("table#%d WriteTo: written=%d, want=%d", i, w.Len(), len(test.data))
		}
		out := w.Bytes()
		for j := range test.data {
			if out[j] != test.data[j] {
				t.Fatalf("table#%d WriteTo: write[%d]=%x, want=%x", i, j, out[j], test.data[j])
			}
		}
	}
}
