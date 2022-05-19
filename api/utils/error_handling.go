package utils

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
func MustGet[T any](v T, err error) T {
	Must(err)
	return v
}
