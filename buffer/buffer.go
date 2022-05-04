package buffer

// Buffer is a variable-sized buffer.
type Buffer struct {
	data []byte
}

// Bytes returns the internal buffer.
func (b *Buffer) Bytes() []byte {
	return b.data[:]
}

// Len returns the length of the buffer's.
func (b *Buffer) Cap() int {
	return cap(b.data)
}

// Len returns the capacity of the buffer's.
func (b *Buffer) Len() int {
	return len(b.data)
}

// grow grows the buffer by at least n bytes.
func (b *Buffer) grow(n int) {
	if cap(b.data) < len(b.data)+n {
		// grow 1.5 size
		resize := cap(b.data)*3/2 + n

		data := make([]byte, resize)
		copy(data, b.data)
		b.data = data[:len(b.data)+n]
	} else {
		b.data = b.data[:len(b.data)+n]
	}
}
