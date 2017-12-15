package slice

import "reflect"

func Each(slice interface{}, callback func(i int)) {
	rv := reflect.ValueOf(slice)
	length := rv.Len()

	for i := 0; i < length; i++ {
		callback(i)
	}
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
