package main

import (
	"unicode/utf8"
)

func rotateRunes(s string, shift int) string {
	sLen := utf8.RuneCountInString(s)

	if sLen == 0 || shift%sLen == 0 {
		return string([]rune(s))
	}

	shiftedS := make([]rune, sLen)
	shift = ((shift % sLen) + sLen) % sLen

	for i, val := range []rune(s) {
		shiftedI := (i - shift + sLen) % sLen
		shiftedS[shiftedI] = val
	}

	return string(shiftedS)
}
