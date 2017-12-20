package dict

import (
	"reflect"
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	dict := map[string]int{"a": 1, "b": 2, "c": 3}
	expect := []string{"a", "b", "c"}

	result := Keys(dict).([]string)
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

	result := Values(dict).([]int)
	sort.Slice(result, func(a, b int) bool {
		return result[a] < result[b]
	})

	if reflect.DeepEqual(expect, result) == false {
		t.Fatalf("Values: want %v, get %v", expect, result)
	}
}
