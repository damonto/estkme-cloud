package rlpa

import "unicode"

func ToTitle(s string) string {
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}
