package stringcmp

import (
	"unicode"
)

// RuneMapper is a function which maps a rune to a rune.
type RuneMapper = func(rune) rune

// ChainMapper creates a new RuneMapper which calls the given RuneMapper
// functions sequentially and return the final result.
func ChainMapper(mappers ...RuneMapper) RuneMapper {
	return func(r rune) rune {
		for _, mapper := range mappers {
			r = mapper(r)
		}

		return r
	}
}

// ReplaceNonLettersWithSpaceMapper maps runes which aren't letters or numbers
// to a space, but it only ever uses one space in a row.
func ReplaceNonLettersWithSpaceMapper() RuneMapper {
	prevIsSpace := false

	return func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			prevIsSpace = false
			return r
		}

		if !prevIsSpace {
			prevIsSpace = true
			return ' '
		}

		return -1
	}
}

// RemoveSpaceMapper returns a RuneMapper that removes all unicode space runes.
func RemoveSpaceMapper() RuneMapper {
	return func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}

		return r
	}
}

// RemoveNonLetterMapper returns a RuneMapper that removes all runes that aren't
// letters or numbers.
func RemoveNonLettersMapper() RuneMapper {
	return func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}

		return -1
	}
}

// LowerMapper maps the rune to its lowercase representation.
func LowerMapper() RuneMapper {
	return unicode.ToLower
}
