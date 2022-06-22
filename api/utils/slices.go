package utils

func FindFunc[E any](input []E, f func(E) bool) int {
	for i, vs := range input {
		if f(vs) {
			return i
		}
	}
	return -1
}

func Find[E comparable](input []E, v E) int {
	return FindFunc(input, func(value E) bool { return value == v })
}

func RemoveIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}
