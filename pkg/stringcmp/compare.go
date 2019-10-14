package stringcmp

import (
	"strings"
)

var letterLowerMapper = ChainMapper(ReplaceNonLettersWithSpaceMapper(), LowerMapper())

// GetWordsFocusedString prepares a string in a manner which makes it useful for
// word by word comparisons.
func GetWordsFocusedString(s string) string {
	return strings.Map(letterLowerMapper, s)
}

// ContainsWords checks whether the given substring is contained in s.
// The check ignores spaces within the substring, as long as it is surrounded by
// either spaces or word ends.
func ContainsWords(s, substring string) bool {
	substring = strings.Map(RemoveNonLettersMapper(), substring)

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

// WordsContainedInAny checks whether the given string is contained in any of
// the provided strings.
func WordsContainedInAny(words string, texts ...string) bool {
	words = GetWordsFocusedString(words)
	for _, text := range texts {
		if strings.Contains(GetWordsFocusedString(text), words) {
			return true
		}
	}

	return false
}
