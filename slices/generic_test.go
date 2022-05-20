package slices

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func helperSliceCompare[T comparable](t *testing.T, ss *Slice[T], tt []T) {
	t.Helper()

	assert.Equalf(t, len(tt), ss.Len(), "Slice length != %d", len(tt))
	assert.Truef(t, ss.Cap() >= len(tt), "Slice capacity < %d", len(tt))

	for i := range tt {
		assert.Equalf(t, tt[i], ss.At(i), "Slice[%d] != %v", i, tt[i])
	}
}

func TestSlice(t *testing.T) {
	var t2 []int

	ss := NewSlice[int]()
	tt := make([]int, 0)
	assert.Equal(t, ss.Len(), 0, "Slice should be empty")
	assert.Truef(t, ss.Cap() >= 0, "Slice capacity < %d", 0)

	ss.Append(1)
	tt = append(tt, 1)
	helperSliceCompare(t, ss, tt)

	ss.Appends(2, 3, 4, 5)
	tt = append(tt, []int{2, 3, 4, 5}...)
	helperSliceCompare(t, ss, tt)

	ss.Set(1, 20)
	tt[1] = 20
	helperSliceCompare(t, ss, tt)

	ss.Insert(2, 30)
	t2 = make([]int, len(tt)+1)
	copy(t2[0:], tt[0:2])
	copy(t2[3:], tt[2:])
	t2[2] = 30
	tt = t2
	helperSliceCompare(t, ss, tt)

	helperSliceCompare(t, ss.Sub(2, 4), tt[2:4])

	ss.Delete(2)
	t2 = make([]int, len(tt)-1)
	copy(t2[0:], tt[0:2])
	copy(t2[2:], tt[3:])
	tt = t2
	helperSliceCompare(t, ss, tt)

	ss.Erase(2, 4)
	t2 = make([]int, len(tt)-2)
	copy(t2[0:], tt[0:2])
	copy(t2[2:], tt[4:])
	tt = t2
	helperSliceCompare(t, ss, tt)

	s2 := ss.Clone()
	helperSliceCompare(t, s2, tt)
	s2.Clear()
	helperSliceCompare(t, s2, make([]int, 0))
	helperSliceCompare(t, ss, tt)

	ss.Reserve(100)
	assert.Truef(t, ss.Cap() >= 100, "Slice capacity < %d", 100)
	helperSliceCompare(t, ss, tt)

	ss.Shrink(1)
	assert.Truef(t, ss.Cap() >= len(tt), "Slice capacity < %d", len(tt))
	helperSliceCompare(t, ss, tt)

	ss.Resize(10) // expand resize
	t2 = make([]int, 10)
	copy(t2, tt)
	tt = t2
	helperSliceCompare(t, ss, tt)

	ss.Resize(5) // shrink resize
	t2 = make([]int, 5)
	copy(t2, tt)
	tt = t2
	helperSliceCompare(t, ss, tt)

	raw := ss.Data()
	assert.Equalf(t, len(tt), len(raw), "Slice.data size != %d", len(tt))
	for i := range tt {
		assert.Equalf(t, tt[i], raw[i], "Slice.data[%d] != %v", i, tt[i])
	}
}

func ExampleSlice() {
	s := NewSlice[int]()

	s.Appends(0, 1, 2)
	for i := 0; i < s.Len(); i++ {
		s.Set(i, 10+s.At(i))
	}

	for i := 0; i < s.Len(); i++ {
		fmt.Printf("s[%d] = %d\n", i, s.At(i))
	}
	fmt.Printf("len(s) = %d\n", s.Len())

	// Output:
	// s[0] = 10
	// s[1] = 11
	// s[2] = 12
	// len(s) = 3
}
