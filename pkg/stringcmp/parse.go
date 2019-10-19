package stringcmp

import (
	"container/list"
	"github.com/gieseladev/elakshi/pkg/stringcmp/bracket"
	"strings"
	"unicode"
	"unicode/utf8"
)

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
					prevEnd = i + 1
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

func mapStringList(l *list.List, f func(string) string) {
	for e := l.Front(); e != nil; e = e.Next() {
		e.Value = f(e.Value.(string))
	}
}

func mapConcatStringList(l *list.List, f func(string) []string) {
	for e := l.Front(); e != nil; {
		strs := f(e.Value.(string))
		if len(strs) == 0 {
			e, _ = e.Next(), l.Remove(e)
			continue
		}

		e.Value = strs[0]
		for _, s := range strs[1:] {
			e = l.InsertAfter(s, e)
		}

		e = e.Next()
	}
}

func filterStringList(l *list.List, f func(string) bool) {
	for e := l.Front(); e != nil; {
		if !f(e.Value.(string)) {
			e, _ = e.Next(), l.Remove(e)
			continue
		}

		e = e.Next()
	}
}

func listToStringSlice(l *list.List) []string {
	s := make([]string, l.Len())
	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		s[i] = e.Value.(string)
		i++
	}

	return s
}

// SplitParts splits the given string into parts.
func SplitParts(s string) []string {
	l := list.New()
	l.PushBack(s)

	mapConcatStringList(l, bracket.SplitBracketGroupContent)
	mapConcatStringList(l, SplitOnDash)

	mapStringList(l, strings.TrimSpace)

	filterStringList(l, func(s string) bool {
		return s != ""
	})

	return listToStringSlice(l)
}
