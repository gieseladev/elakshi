package songtitle

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// isLetterNumber checks whether the rune is either a letter or a number.
func isLetterNumber(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// SplitOnDash splits the given string into parts separated by unicode dashes.
// Dashes must be surrounded by neither letter nor number runes to be separated.
func SplitOnDash(s string) []string {
	var parts []string
	prevEnd := 0
	prevIsLN := false
	for i, r := range s {
		if unicode.Is(unicode.Pd, r) {
			if !prevIsLN {
				nr, _ := utf8.DecodeRuneInString(s[i+utf8.RuneLen(r):])
				if !isLetterNumber(nr) {
					parts = append(parts, s[prevEnd:i])
					prevEnd = i + utf8.RuneLen(r)
				}
			}

			prevIsLN = false
			continue
		}

		if isLetterNumber(r) {
			prevIsLN = true
		} else {
			prevIsLN = false
		}
	}

	parts = append(parts, s[prevEnd:])

	return parts
}

// SplitOnAnyRuneOf splits a string on any rune found in delimiters.
func SplitOnAnyRuneOf(s string, delimiters []rune) []string {
	var parts []string

	delimiterStr := string(delimiters)
	for {
		i := strings.IndexAny(s, delimiterStr)
		if i == -1 {
			break
		}

		parts = append(parts, s[:i])

		_, size := utf8.DecodeRuneInString(s[i:])
		s = s[i+size:]
	}

	parts = append(parts, s)

	return parts
}
