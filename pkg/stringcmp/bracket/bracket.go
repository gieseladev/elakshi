package bracket

import (
	"sort"
	"strings"
	"unicode"
)

type bracketMatch struct {
	closingRune rune
	OpenIndex   int
	CloseIndex  int
}

func bracketGroupMatches(s string) []bracketMatch {
	var matches []bracketMatch

	var potentialMatches []bracketMatch
	for i, r := range s {
		if closingRune, ok := oPairs[r]; ok {
			potentialMatches = append(potentialMatches, bracketMatch{
				closingRune: closingRune,
				OpenIndex:   i,
			})
			continue
		}

		if len(potentialMatches) != 0 && unicode.Is(unicode.Pe, r) {
			for j := len(potentialMatches) - 1; j >= 0; j-- {
				match := potentialMatches[j]

				if match.closingRune != r {
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

	prevCl := start - 1

	for i := 0; i < len(matches); {
		pmOp := matches[i].OpenIndex
		pmCl := matches[i].CloseIndex

		// find all matches within the current match.
		childrenEnd := i + 1
		for ; childrenEnd < len(matches); childrenEnd++ {
			if matches[childrenEnd].CloseIndex > pmCl {
				break
			}
		}

		between.WriteString(s[prevCl+1 : pmOp])
		prevCl = pmCl

		children := matches[i+1 : childrenEnd]
		if len(children) == 0 {
			// no children, trivial
			parts[i+1] = s[pmOp+1 : pmCl]
			i++
			continue
		}

		splitBracketGroupContent(s, pmOp+1, pmCl, children, parts[i+1:childrenEnd+1])

		i = childrenEnd
	}

	between.WriteString(s[prevCl+1 : end])
	parts[0] = between.String()
}

// SplitBracketGroupContent extracts the contents in brackets and returns them
// separately.
// The first element is always the contents not located in brackets.
// The remaining elements are sorted by their starting bracket position.
//
// The content of brackets within other brackets is extracted separately and
// excluded from the parent brackets (ex: "(a (b))" will yield ["", "a ", "b"]).
func SplitBracketGroupContent(s string) []string {
	matches := bracketGroupMatches(s)
	switch len(matches) {
	case 0:
		return []string{s}
	case 1:
		match := matches[0]
		return []string{
			// surrounding strings
			s[:match.OpenIndex] + s[match.CloseIndex+1:],
			// bracket contents
			s[match.OpenIndex+1 : match.CloseIndex],
		}
	}

	parts := make([]string, len(matches)+1)
	splitBracketGroupContent(s, 0, len(s), matches, parts)

	return parts
}
