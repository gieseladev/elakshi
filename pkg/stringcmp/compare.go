package stringcmp

import (
	"strings"
	"unicode"
)

// TODO maybe replace non-letters with spaces

// RemoveNonLetters removes all runes which aren't unicode letters.
func RemoveNonLetters(s string) string {
	prevIsSpace := false

	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			prevIsSpace = false
			return r
		} else if !prevIsSpace && unicode.IsSpace(r) {
			prevIsSpace = true
			return ' '
		}

		return -1
	}, s)
}

func ContainsAnyOf(text string, substrs ...string) (bool, string) {
	for _, s := range substrs {
		if strings.Contains(text, s) {
			return true, s
		}
	}

	return false, ""
}

func ContainsWords(text, words string) bool {
	return strings.Contains(RemoveNonLetters(text), RemoveNonLetters(words))
}

func WordsContainedInAny(words string, texts ...string) bool {
	words = RemoveNonLetters(words)
	for _, text := range texts {
		if strings.Contains(RemoveNonLetters(text), words) {
			return true
		}
	}

	return false
}
