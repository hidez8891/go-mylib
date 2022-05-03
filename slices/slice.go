// Package slice provides slice helper function.
package slice

import "reflect"

// Any tests whether at least one value in the slice pass pred().
func Any(slice interface{}, pred func(i int) bool) bool {
	return FindIndexIf(slice, pred) >= 0
}

// All tests whether all values in the slice pass pred().
func All(slice interface{}, pred func(i int) bool) bool {
	rv := reflect.ValueOf(slice)
	length := rv.Len()

	for i := 0; i < length; i++ {
		if pred(i) == false {
			return false
		}
	}
	return true
}

// ForEach executes callback() for each slice's value.
func ForEach(slice interface{}, callback func(i int)) {
	rv := reflect.ValueOf(slice)
	length := rv.Len()

	for i := 0; i < length; i++ {
		callback(i)
	}
}

// Filter creates a new slice with all values that pass pred().
func Filter(slice interface{}, pred func(i int) bool) interface{} {
	rv := reflect.ValueOf(slice)
	out := reflect.MakeSlice(reflect.SliceOf(rv.Type().Elem()), 0, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		if pred(i) {
			out = reflect.Append(out, rv.Index(i))
		}
	}
	return out.Interface()
}

// FindIndex returns index of the value in the slice.
// Otherwise -1 is returned.
func FindIndex(slice interface{}, value interface{}) int {
	rv := reflect.ValueOf(slice)
	return FindIndexIf(slice, func(i int) bool {
		return reflect.DeepEqual(rv.Index(i).Interface(), value)
	})
}

// FindIndexIf returns index of the value in the slice that satisfy cond().
// Otherwise -1 is returned.
func FindIndexIf(slice interface{}, cond func(i int) bool) int {
	rv := reflect.ValueOf(slice)
	length := rv.Len()

	for i := 0; i < length; i++ {
		if cond(i) {
			return i
		}
	}
	return -1
}

// Includes determines whether the slice includes the value.
func Includes(slice interface{}, value interface{}) bool {
	return FindIndex(slice, value) >= 0
}

// Map creates a new slice with the results of calling conv() on
// every element in the slice.
func Map(slice interface{}, conv interface{}) interface{} {
	rv := reflect.ValueOf(slice)
	fun := reflect.ValueOf(conv)
	out := reflect.MakeSlice(reflect.SliceOf(fun.Type().Out(0)), 0, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		rs := fun.Call([]reflect.Value{reflect.ValueOf(i)})
		out = reflect.Append(out, rs[0])
	}
	return out.Interface()
}

// Reduce applies a function against accumulater() and each element
// in the slice to reduce it to a single value.
func Reduce(slice interface{}, accumulater interface{}) interface{} {
	rv := reflect.ValueOf(slice)
	fun := reflect.ValueOf(accumulater)
	acc := reflect.New(fun.Type().Out(0)).Elem()

	for i := 0; i < rv.Len(); i++ {
		rs := fun.Call([]reflect.Value{acc, reflect.ValueOf(i)})
		acc = rs[0]
	}

	return acc.Interface()
}
