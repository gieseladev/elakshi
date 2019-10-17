package bidi

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

func splitBracketGroupContent(s string, start, end int, matches []bracketMatch) []string {
	var between strings.Builder
	// reserve slot for "between"
	parts := make([]string, 1, len(matches)+1)

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
			parts = append(parts, s[pmOp+1:pmCl])
			i++
			continue
		}

		cParts := splitBracketGroupContent(s, pmOp+1, pmCl, children)
		parts = append(parts, cParts...)

		i = childrenEnd
	}

	between.WriteString(s[prevCl+1 : end])
	parts[0] = between.String()

	return parts
}

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

	return splitBracketGroupContent(s, 0, len(s), matches)
}
