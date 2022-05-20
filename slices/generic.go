package slices

// Slice provides a generic slice methods.
type Slice[T comparable] struct {
	s []T
}

// NewSlice returns Slice[T].
func NewSlice[T comparable]() *Slice[T] {
	return &Slice[T]{
		s: make([]T, 0),
	}
}

// Cap returns Slice[T] capacity.
func (s *Slice[T]) Cap() int {
	return cap(s.s)
}

// Len returns Slice[T] length.
func (s *Slice[T]) Len() int {
	return len(s.s)
}

// Shrink reduces the capacity to `cap`.
func (s *Slice[T]) Shrink(cap int) {
	size := s.Len()
	if cap-size > 0 {
		ns := make([]T, size, cap)
		copy(ns, s.s)
		s.s = ns
	}
}

// Reserve extends the capacity to `cap`.
func (s *Slice[T]) Reserve(cap int) {
	size := s.Len()
	n := cap - size
	if n > 0 {
		s.s = append(s.s, make([]T, n)...)[:size]
	}
}

// Resize changes the number of elements to `len`.
func (s *Slice[T]) Resize(len int) {
	size := s.Len()
	n := len - size
	if n > 0 {
		s.s = append(s.s, make([]T, n)...)
	} else {
		s.s = s.s[:len]
	}
}

// At returns the i-th elemen.
func (s *Slice[T]) At(i int) T {
	return s.s[i]
}

// Set sets `v` to the i-th element.
func (s *Slice[T]) Set(i int, v T) {
	s.s[i] = v
}

// Insert inserts `v` into the i-th element.
func (s *Slice[T]) Insert(i int, v T) {
	s.s = append(s.s[:i+1], s.s[i:]...)
	s.s[i] = v
}

// Append appends `v` to the end of the slice.
func (s *Slice[T]) Append(v T) {
	s.s = append(s.s, v)
}

// Appends adds one or more elements to the end.
func (s *Slice[T]) Appends(vs ...T) {
	s.s = append(s.s, vs...)
}

// Delete deletes the i-th element.
func (s *Slice[T]) Delete(i int) {
	copy(s.s[i:], s.s[i+1:])
	s.s = s.s[:len(s.s)-1]
}

// Erase deletes elements from `begin` to `end - 1`.
func (s *Slice[T]) Erase(begin int, end int) {
	n := s.Len() - (end - begin)
	copy(s.s[begin:], s.s[end:])
	s.s = s.s[:n]
}

// Clear erases all elements.
func (s *Slice[T]) Clear() {
	s.s = make([]T, 0)
}

// Clone returns a Slice with all elements copied.
func (s *Slice[T]) Clone() *Slice[T] {
	ns := make([]T, len(s.s))
	copy(ns, s.s)
	return &Slice[T]{
		s: ns,
	}
}

// Sub returns a reference to an Slice from `begin` to `end - 1`.
func (s *Slice[T]) Sub(begin int, end int) *Slice[T] {
	return &Slice[T]{
		s: s.s[begin:end],
	}
}

// Data returns the raw internal array.
func (s *Slice[T]) Data() []T {
	return s.s[:]
}
