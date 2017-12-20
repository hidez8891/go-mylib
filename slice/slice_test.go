package slice

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAny(t *testing.T) {
	tests := []struct {
		array  []int
		value  int
		expect bool
	}{
		{[]int{1, 2, 3, 4, 5}, 5, true},
		{[]int{1, 2, 3, 4, 5}, 6, false},
	}

	for _, test := range tests {
		result := Any(test.array, func(i int) bool {
			return test.array[i] == test.value
		})
		if result != test.expect {
			t.Fatalf("Any: want %v, get %v", test.expect, result)
		}
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		array  []int
		expect bool
	}{
		{[]int{1, 2, 3, 4, 5}, false},
		{[]int{2, 4, 6, 8, 10}, true},
	}

	for _, test := range tests {
		result := All(test.array, func(i int) bool {
			return test.array[i]%2 == 0
		})
		if result != test.expect {
			t.Fatalf("All: want %v, get %v", test.expect, result)
		}
	}
}

func TestForEach(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		expect := 15

		acc := 0
		ForEach(array, func(i int) {
			acc += array[i]
		})

		if acc != expect {
			t.Fatalf("ForEach: accumulate %v, want %v, get %v", array, expect, acc)
		}
	}

	{
		array := []string{"a", "b", "c", "d"}
		expect := "abcd"

		acc := ""
		ForEach(array, func(i int) {
			acc += array[i]
		})

		if acc != expect {
			t.Fatalf("ForEach: accumulate %v, want %v, get %v", array, expect, acc)
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

func TestFindIndex(t *testing.T) {
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
		index := FindIndex(test.array, test.value)
		if index != test.expect {
			t.Fatalf("FindIndex(%v, %v), want %v, get %v",
				test.array, test.value, test.expect, index)
		}
	}
}

func TestFindIndexIf(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		value := 5
		expect := 4

		index := FindIndexIf(array, func(i int) bool {
			return array[i] == value
		})

		if index != expect {
			t.Fatalf("FindIndexIf(%v, %v), want %v, get %v",
				array, value, expect, index)
		}
	}

	{
		array := []string{"a", "b", "c", "d"}
		value := "d"
		expect := 3

		index := FindIndexIf(array, func(i int) bool {
			return array[i] == value
		})

		if index != expect {
			t.Fatalf("FindIndexIf(%v, %v), want %v, get %v",
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
		t.Fatalf("Map: want %v, get %v", expect, result)
	}
}

func TestReduce(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		expect := 15

		result := Reduce(array, func(acc, i int) int {
			return acc + array[i]
		}).(int)

		if expect != result {
			t.Fatalf("Reduce: want %v, get %v", expect, result)
		}
	}

	{
		array := []int{1, 2, 3, 4, 5}
		expect := "12345"

		result := Reduce(array, func(acc string, i int) string {
			return acc + fmt.Sprintf("%d", array[i])
		}).(string)

		if expect != result {
			t.Fatalf("Reduce: want %v, get %v", expect, result)
		}
	}
}
