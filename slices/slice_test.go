package slices

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
		result := Any(test.array, func(v int) bool {
			return v == test.value
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
		result := All(test.array, func(v int) bool {
			return v%2 == 0
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
		ForEach(array, func(v int) {
			acc += v
		})

		if acc != expect {
			t.Fatalf("ForEach: accumulate %v, want %v, get %v", array, expect, acc)
		}
	}

	{
		array := []string{"a", "b", "c", "d"}
		expect := "abcd"

		acc := ""
		ForEach(array, func(v string) {
			acc += v
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
		pred := func(v int) bool { return v%2 == 1 }

		result := Filter(array, pred)
		if reflect.DeepEqual(result, expect) == false {
			t.Fatalf("Filter: want %v, get %v", expect, result)
		}
	}

	{
		array := []string{"a", "ab", "ac", "bd", "e"}
		expect := []string{"ab", "ac", "bd"}
		pred := func(v string) bool { return len(v) > 1 }

		result := Filter(array, pred)
		if reflect.DeepEqual(result, expect) == false {
			t.Fatalf("Filter: want %v, get %v", expect, result)
		}
	}
}

func TestFindIndex(t *testing.T) {
	{
		tests := []struct {
			array  []int
			value  int
			expect int
		}{
			{[]int{1, 2, 3, 4, 5}, 5, 4},
			{[]int{1, 2, 3, 4, 5}, 6, -1},
		}

		for _, test := range tests {
			index := FindIndex(test.array, test.value)
			if index != test.expect {
				t.Fatalf("FindIndex(%v, %v), want %v, get %v",
					test.array, test.value, test.expect, index)
			}
		}
	}

	{
		tests := []struct {
			array  []string
			value  string
			expect int
		}{
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
}

func TestFindIndexIf(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		value := 5
		expect := 4

		index := FindIndexIf(array, func(v int) bool {
			return v == value
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

		index := FindIndexIf(array, func(s string) bool {
			return s == value
		})

		if index != expect {
			t.Fatalf("FindIndexIf(%v, %v), want %v, get %v",
				array, value, expect, index)
		}
	}
}

func TestIncludes(t *testing.T) {
	{
		tests := []struct {
			array  []int
			value  int
			expect bool
		}{
			{[]int{1, 2, 3, 4, 5}, 5, true},
			{[]int{1, 2, 3, 4, 5}, 6, false},
		}

		for _, test := range tests {
			result := Includes(test.array, test.value)
			if result != test.expect {
				t.Fatalf("Includes(%v, %v), want %v, get %v",
					test.array, test.value, test.expect, result)
			}
		}
	}

	{
		tests := []struct {
			array  []string
			value  string
			expect bool
		}{
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
}

func TestMap(t *testing.T) {
	array := []int{1, 2, 3, 4, 5}
	expect := []string{"1", "2", "3", "4", "5"}

	result := Map(array, func(v int) string {
		return fmt.Sprintf("%d", v)
	})

	if reflect.DeepEqual(result, expect) == false {
		t.Fatalf("Map: want %v, get %v", expect, result)
	}
}

func TestReduce(t *testing.T) {
	{
		array := []int{1, 2, 3, 4, 5}
		expect := 1015

		result := Reduce(array, func(acc, v int) int {
			return acc + v
		}, 1000)

		if expect != result {
			t.Fatalf("Reduce: want %v, get %v", expect, result)
		}
	}

	{
		array := []int{1, 2, 3, 4, 5}
		expect := "pre-12345"

		result := Reduce(array, func(acc string, v int) string {
			return acc + fmt.Sprintf("%d", v)
		}, "pre-")

		if expect != result {
			t.Fatalf("Reduce: want %v, get %v", expect, result)
		}
	}
}

func ExampleAny() {
	array := []int{1, 2, 3, 4, 5}

	fmt.Println(Any(array, func(v int) bool {
		return v%2 == 0
	}))
	// Output:
	// true
}

func ExampleAll() {
	array := []int{2, 4, 6, 8, 10}

	fmt.Println(All(array, func(v int) bool {
		return v%2 == 0
	}))
	// Output:
	// true
}

func ExampleForEach() {
	array := []int{1, 2, 3, 4, 5}

	ForEach(array, func(v int) {
		fmt.Println(v)
	})
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleFilter() {
	array := []int{1, 2, 3, 4, 5}
	result := Filter(array, func(v int) bool {
		return v%2 == 1
	})

	for _, v := range result {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 3
	// 5
}

func ExampleFindIndex() {
	array := []int{1, 2, 3, 4, 5}

	fmt.Println(FindIndex(array, 4))
	// Output:
	// 3
}

func ExampleFindIndexIf() {
	array := []int{1, 2, 3, 4, 5}

	fmt.Println(FindIndexIf(array, func(v int) bool {
		return v%2 == 0
	}))
	// Output:
	// 1
}

func ExampleIncludes() {
	array := []int{1, 2, 3, 4, 5}

	fmt.Println(Includes(array, 4))
	// Output:
	// true
}

func ExampleMap() {
	array := []int{1, 2, 3, 4, 5}
	result := Map(array, func(v int) int {
		return v * 2
	})

	for _, v := range result {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
	// 8
	// 10
}

func ExampleReduce() {
	array := []int{1, 2, 3, 4, 5}

	fmt.Println(Reduce(array, func(acc string, v int) string {
		return acc + fmt.Sprintf("%d", v)
	}, ""))
	// Output:
	// 12345
}
