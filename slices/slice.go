// Package slice provides slice helper function.
package slice

// Any tests whether at least one value in the slice pass pred().
func Any[T comparable](s []T, pred func(v T) bool) bool {
	return FindIndexIf(s, pred) >= 0
}

// All tests whether all values in the slice pass pred().
func All[T comparable](s []T, pred func(v T) bool) bool {
	for _, v := range s {
		if pred(v) == false {
			return false
		}
	}
	return true
}

// ForEach executes callback() for each slice's value.
func ForEach[T comparable](s []T, callback func(v T)) {
	for _, v := range s {
		callback(v)
	}
}

// Filter creates a new slice with all values that pass pred().
func Filter[T comparable](s []T, pred func(v T) bool) []T {
	r := make([]T, 0, len(s))
	for _, v := range s {
		if pred(v) {
			r = append(r, v)
		}
	}
	return r
}

// FindIndex returns index of the value in the slice.
// Otherwise -1 is returned.
func FindIndex[T comparable](s []T, v T) int {
	return FindIndexIf(s, func(t T) bool {
		return v == t
	})
}

// FindIndexIf returns index of the value in the slice that satisfy cond().
// Otherwise -1 is returned.
func FindIndexIf[T comparable](s []T, cond func(v T) bool) int {
	for i, v := range s {
		if cond(v) {
			return i
		}
	}
	return -1
}

// Includes determines whether the slice includes the value.
func Includes[T comparable](s []T, v T) bool {
	return FindIndex(s, v) >= 0
}

// Map creates a new slice with the results of calling conv() on
// every element in the slice.
func Map[T, R any](s []T, conv func(T) R) []R {
	r := make([]R, 0, len(s))
	for _, v := range s {
		r = append(r, conv(v))
	}
	return r
}

// Reduce applies a function against accumulater() and each element
// in the slice to reduce it to a single value.
func Reduce[T, R any](s []T, accumulater func(R, T) R, init R) R {
	acc := init
	for _, v := range s {
		acc = accumulater(acc, v)
	}
	return acc
}
