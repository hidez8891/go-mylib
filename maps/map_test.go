package maps

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	dict := map[string]int{"a": 1, "b": 2, "c": 3}
	expect := []string{"a", "b", "c"}

	result := Keys(dict)
	sort.Slice(result, func(a, b int) bool {
		return result[a] < result[b]
	})

	if reflect.DeepEqual(expect, result) == false {
		t.Fatalf("Keys: want %v, get %v", expect, result)
	}
}

func TestValues(t *testing.T) {
	dict := map[string]int{"a": 1, "b": 2, "c": 3}
	expect := []int{1, 2, 3}

	result := Values(dict)
	sort.Slice(result, func(a, b int) bool {
		return result[a] < result[b]
	})

	if reflect.DeepEqual(expect, result) == false {
		t.Fatalf("Values: want %v, get %v", expect, result)
	}
}

func ExampleKeys() {
	dict := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := Keys(dict)

	for _, k := range keys {
		fmt.Println(k)
	}
	// Unordered Output:
	// a
	// b
	// c
}

func ExampleValues() {
	dict := map[string]int{"a": 1, "b": 2, "c": 3}
	vals := Values(dict)

	for _, v := range vals {
		fmt.Println(v)
	}
	// Unordered Output:
	// 1
	// 2
	// 3
}
