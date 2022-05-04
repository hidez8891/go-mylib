package buffer

import "testing"

func TestBufferEmpty(t *testing.T) {
	buf := new(Buffer)
	if buf.Len() != 0 {
		t.Errorf("Buffer.Len=%d, want %d", buf.Len(), 0)
	}
	if buf.Cap() != 0 {
		t.Errorf("Buffer.Cap=%d, want %d", buf.Cap(), 0)
	}
}

func TestBufferGrow(t *testing.T) {
	buf := new(Buffer)
	buf.grow(10)
	if buf.Len() != 10 {
		t.Errorf("Buffer.Len=%d, want %d", buf.Len(), 10)
	}
	if buf.Cap() < 10 {
		t.Errorf("Buffer.Cap=%d, at least %d", buf.Cap(), 10)
	}
}
