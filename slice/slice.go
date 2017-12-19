package slice

import "reflect"

func Each(slice interface{}, callback func(i int)) {
	rv := reflect.ValueOf(slice)
	length := rv.Len()

	for i := 0; i < length; i++ {
		callback(i)
	}
}

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

func Find(slice interface{}, value interface{}) int {
	rv := reflect.ValueOf(slice)
	return FindIf(slice, func(i int) bool {
		return reflect.DeepEqual(rv.Index(i).Interface(), value)
	})
}

func FindIf(slice interface{}, cond func(i int) bool) int {
	rv := reflect.ValueOf(slice)
	length := rv.Len()

	for i := 0; i < length; i++ {
		if cond(i) {
			return i
		}
	}
	return -1
}

func Includes(slice interface{}, value interface{}) bool {
	return Find(slice, value) >= 0
}

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
