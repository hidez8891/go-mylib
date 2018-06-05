// Package dict provides map[] helper function.
package dict

import "reflect"

// Keys return map[]'s key slice
func Keys(dict interface{}) interface{} {
	rv := reflect.ValueOf(dict)
	out := reflect.MakeSlice(reflect.SliceOf(rv.Type().Key()), 0, rv.Len())

	for _, k := range rv.MapKeys() {
		out = reflect.Append(out, k)
	}
	return out.Interface()
}

// Values return map[]'s value slice
func Values(dict interface{}) interface{} {
	rv := reflect.ValueOf(dict)
	out := reflect.MakeSlice(reflect.SliceOf(rv.Type().Elem()), 0, rv.Len())

	for _, k := range rv.MapKeys() {
		v := rv.MapIndex(k)
		out = reflect.Append(out, v)
	}
	return out.Interface()
}
