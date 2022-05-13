package utils

func Find[E comparable](s []E, v E) int {
	for i, vs := range s {
		if v == vs {
			return i
		}
	}
	return -1
}

func RemoveIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func ToGenericArray(arr ...interface{}) []interface{} {
	return arr
}
