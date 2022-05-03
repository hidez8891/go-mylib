// Package maps provides map[] helper function.
package maps

// Keys return map[]'s key slice
func Keys[K comparable, T any](m map[K]T) []K {
	out := make([]K, 0, len(m))
	for k, _ := range m {
		out = append(out, k)
	}
	return out
}

// Values return map[]'s value slice
func Values[K comparable, T any](m map[K]T) []T {
	out := make([]T, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
