package slice

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEach(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		expect := 15

		acc := 0
		Each(array, func(i int) {
			acc += array[i]
		})

		if acc != expect {
			t.Fatalf("Each: accumulate %v, want %v, get %v", array, expect, acc)
		}
	}

	{
		array := []string{"a", "b", "c", "d"}
		expect := "abcd"

		acc := ""
		Each(array, func(i int) {
			acc += array[i]
		})

		if acc != expect {
			t.Fatalf("Each: accumulate %v, want %v, get %v", array, expect, acc)
		}
	}
}

func TestFilter(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		expect := []int{1, 3, 5}
		pred := func(i int) bool { return array[i]%2 == 1 }

		result := Filter(array, pred)
		if reflect.DeepEqual(result, expect) == false {
			t.Fatalf("Filter: want %v, get %v", expect, result)
		}
	}

	{
		array := []string{"a", "ab", "ac", "bd", "e"}
		expect := []string{"ab", "ac", "bd"}
		pred := func(i int) bool { return len(array[i]) > 1 }

		result := Filter(array, pred)
		if reflect.DeepEqual(result, expect) == false {
			t.Fatalf("Filter: want %v, get %v", expect, result)
		}
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		array  interface{}
		value  interface{}
		expect int
	}{
		{[]int{1, 2, 3, 4, 5}, 5, 4},
		{[]int{1, 2, 3, 4, 5}, 6, -1},
		{[]string{"a", "ab", "ba", "bb", "b"}, "b", 4},
		{[]string{"a", "ab", "ba", "bb", "b"}, "c", -1},
	}

	for _, test := range tests {
		index := Find(test.array, test.value)
		if index != test.expect {
			t.Fatalf("Find(%v, %v), want %v, get %v",
				test.array, test.value, test.expect, index)
		}
	}
}

func TestFindIf(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		value := 5
		expect := 4

		index := FindIf(array, func(i int) bool {
			return array[i] == value
		})

		if index != expect {
			t.Fatalf("FindIf(%v, %v), want %v, get %v",
				array, value, expect, index)
		}
	}

	{
		array := []string{"a", "b", "c", "d"}
		value := "d"
		expect := 3

		index := FindIf(array, func(i int) bool {
			return array[i] == value
		})

		if index != expect {
			t.Fatalf("FindIf(%v, %v), want %v, get %v",
				array, value, expect, index)
		}
	}
}

func TestIncludes(t *testing.T) {
	tests := []struct {
		array  interface{}
		value  interface{}
		expect bool
	}{
		{[]int{1, 2, 3, 4, 5}, 5, true},
		{[]int{1, 2, 3, 4, 5}, 6, false},
		{[]string{"a", "ab", "ba", "bb", "b"}, "b", true},
		{[]string{"a", "ab", "ba", "bb", "b"}, "c", false},
	}

	for _, test := range tests {
		result := Includes(test.array, test.value)
		if result != test.expect {
			t.Fatalf("Includes(%v, %v), want %v, get %v",
				test.array, test.value, test.expect, result)
		}
	}
}

func TestMap(t *testing.T) {
	array := []int{1, 2, 3, 4, 5}
	expect := []string{"1", "2", "3", "4", "5"}

	result := Map(array, func(i int) string {
		return fmt.Sprintf("%d", array[i])
	}).([]string)

	if reflect.DeepEqual(result, expect) == false {
		t.Fatalf("Map want %v, get %v", expect, result)
	}
}
