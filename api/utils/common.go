package utils

import "strings"

func RepeatWithSeparator(toRepeat string, times int, separator string) string {
	return strings.TrimRight(strings.Repeat(toRepeat+separator, times), separator)
}

func GetMultiParam(n int) string {
	return RepeatWithSeparator("?", n, ",")
}
