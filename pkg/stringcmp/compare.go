package stringcmp

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var letterLowerMapper = ChainMapper(ReplaceNonLettersWithSpaceMapper(), LowerMapper())

// GetWordsFocusedString prepares a string in a manner which makes it useful for
// word by word comparisons.
func GetWordsFocusedString(s string) string {
	return strings.Map(letterLowerMapper, s)
}

// ContainsSurrounded checks whether the substring is contained in s and is
// surrounded by (unicode) spaces.
func ContainsSurrounded(s, substring string) bool {
	if len(substring) == 0 {
		return true
	}

	i := strings.Index(s, substring)
	if i == -1 {
		return false
	}

	// check if preceded by space
	if i > 0 {
		r, _ := utf8.DecodeLastRuneInString(s[:i])
		if !unicode.IsSpace(r) {
			return false
		}
	}

	// check if followed by space
	end := i + len(substring)
	if end < len(s) {
		r, _ := utf8.DecodeRuneInString(s[end:])
		if !unicode.IsSpace(r) {
			return false
		}
	}

	return true
}

// ContainsSurroundedIgnoreSpace checks whether the given substring is contained in s.
// The check ignores spaces within the substring, as long as it is surrounded by
// spaces in s.
func ContainsSurroundedIgnoreSpace(s, substring string) bool {
	substring = strings.Map(RemoveSpaceMapper(), substring)

	words := strings.Fields(s)
	for i, word := range words {
		missing := substring
		for strings.HasPrefix(missing, word) {
			missing = missing[len(word):]
			if missing == "" {
				return true
			}

			i++
			if i == len(words) {
				break
			}

			word = words[i]
		}
	}

	return false
}

// ContainsAnyOf checks if the given text contains any of the provided
// substrings. It returns the substring that matched or the empty string if
// none matched.
func ContainsAnyOf(text string, substrs ...string) string {
	for _, s := range substrs {
		if strings.Contains(text, s) {
			return s
		}
	}

	return ""
}
