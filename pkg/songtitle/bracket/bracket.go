package bracket

//go:generate go run github.com/gieseladev/elakshi/tools/genbidi

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type bracketMatch struct {
	OpeningRune rune
	ClosingRune rune
	OpenIndex   int
	CloseIndex  int
}

func (b bracketMatch) OpenIndexEnd() int {
	return b.OpenIndex + utf8.RuneLen(b.OpeningRune)
}

func (b bracketMatch) CloseIndexEnd() int {
	return b.CloseIndex + utf8.RuneLen(b.ClosingRune)
}

func bracketGroupMatches(s string) []bracketMatch {
	var matches []bracketMatch

	var potentialMatches []bracketMatch
	for i, r := range s {
		if closingRune, ok := oPairs[r]; ok {
			potentialMatches = append(potentialMatches, bracketMatch{
				OpeningRune: r,
				ClosingRune: closingRune,
				OpenIndex:   i,
			})
			continue
		}

		if len(potentialMatches) != 0 && unicode.Is(unicode.Pe, r) {
			for j := len(potentialMatches) - 1; j >= 0; j-- {
				match := potentialMatches[j]

				if match.ClosingRune != r {
					continue
				}

				match.CloseIndex = i
				matches = append(matches, match)
				potentialMatches = potentialMatches[:j]
				break
			}
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].OpenIndex < matches[j].OpenIndex
	})

	return matches
}

func splitBracketGroupContent(s string, start, end int, matches []bracketMatch, parts []string) {
	var between strings.Builder

	prevClEnd := start

	for i := 0; i < len(matches); {
		pmOp := matches[i].OpenIndex
		pmOpEnd := matches[i].OpenIndexEnd()
		pmCl := matches[i].CloseIndex
		pmClEnd := matches[i].CloseIndexEnd()

		between.WriteString(s[prevClEnd:pmOp])
		prevClEnd = pmClEnd

		// find all matches within the current match.
		childrenEnd := i + 1
		for ; childrenEnd < len(matches); childrenEnd++ {
			if matches[childrenEnd].CloseIndex > pmCl {
				break
			}
		}

		children := matches[i+1 : childrenEnd]
		switch len(children) {
		case 0:
			// no children, trivial
			parts[i+1] = s[pmOpEnd:pmCl]
		case 1:
			cOp := children[0].OpenIndex
			cClOp := children[0].OpenIndexEnd()
			cCl := children[0].CloseIndex
			cClEnd := children[0].CloseIndexEnd()

			// surrounding text for parent
			parts[i+1] = s[pmOpEnd:cOp] + s[cClEnd:pmCl]
			// enclosed text to child
			parts[i+2] = s[cClOp:cCl]
		default:
			splitBracketGroupContent(s, pmOpEnd, pmCl, children, parts[i+1:childrenEnd+1])
		}

		i = childrenEnd
	}

	between.WriteString(s[prevClEnd:end])
	parts[0] = between.String()
}

// ExtractContents extracts the contents in brackets and returns them
// separately.
// The first element is always the contents not located in brackets.
// The remaining elements are sorted by their starting bracket position.
//
// The content of brackets within other brackets is extracted separately and
// excluded from the parent brackets (ex: "(a (b))" will yield ["", "a ", "b"]).
func ExtractContents(s string) []string {
	matches := bracketGroupMatches(s)
	switch len(matches) {
	case 0:
		return []string{s}
	case 1:
		match := matches[0]
		return []string{
			// surrounding strings
			s[:match.OpenIndex] + s[match.CloseIndexEnd():],
			// bracket contents
			s[match.OpenIndexEnd():match.CloseIndex],
		}
	}

	parts := make([]string, len(matches)+1)
	splitBracketGroupContent(s, 0, len(s), matches, parts)

	return parts
}
