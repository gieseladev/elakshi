package songtitle

import (
	"github.com/gieseladev/elakshi/pkg/songtitle/bracket"
	"github.com/gieseladev/elakshi/pkg/songtitle/label"
	"strings"
)

type Title struct {
	Raw string // Raw content of the title

	BaselineParts    []string
	OtherParts       []string
	GuestAppearances []string

	ContentLabels []string
}

func parsePart(parsed *Title, part string) bool {
	if part == "" {
		return true
	}

	switch {
	case label.IsContentLabel(part):
		parsed.ContentLabels = append(parsed.ContentLabels, part)
		return true
	case label.IsFiller(part):
		// ignore filler
		return true
	}

	guests := getGuestAppearances(part)
	if len(guests) != 0 {
		parsed.GuestAppearances = append(parsed.GuestAppearances, guests...)
		return true
	}

	return false
}

func ParseTitle(title string) Title {
	parsed := Title{
		Raw: title,
	}

	baseline, others := splitBaselineParts(title)

	for partIndex, part := range baseline {
		subparts := splitVisuallyDistinct(part)
		for i := 0; i < len(subparts); i++ {
			s := strings.Join(subparts[i:], " ")

			if parsePart(&parsed, s) {
				p := strings.Join(subparts[:i], " ")
				if p == "" {
					// TODO again, bad idea. Don't popedipop
					baseline = deleteStringSlice(baseline, partIndex)
				} else {
					baseline[partIndex] = p
				}

				break
			}
		}
	}

	for i, part := range others {
		// TODO removing during iteration is probably a bad idea. Use the in-place
		//  filter method
		if parsePart(&parsed, part) {
			// remove the part from others
			others = deleteStringSlice(others, i)
		}
	}

	// add the remaining parts
	parsed.BaselineParts = baseline
	parsed.OtherParts = others

	return parsed
}

func splitBaselineParts(s string) ([]string, []string) {
	parts := bracket.ExtractContents(s)

	important := SplitOnDash(parts[0])
	mapStringSlice(important, strings.TrimSpace)

	// TODO should we SplitOnDash in brackets as well?
	other := parts[1:]
	mapStringSlice(other, strings.TrimSpace)

	return important, other
}

func splitVisuallyDistinct(s string) []string {
	return strings.Fields(s)
}

func getGuestAppearances(s string) []string {
	// TODO do this more efficiently

	parts := splitVisuallyDistinct(s)
	if len(parts) < 2 || !label.IsGuestAppearance(parts[0]) {
		return nil
	}

	s = strings.Join(parts[1:], " ")

	guests := strings.Split(s, ",")
	last := strings.Split(guests[len(guests)-1], "&")
	guests = append(guests[:len(guests)-1], last...)

	mapStringSlice(guests, strings.TrimSpace)

	return guests
}

func CountIrrelevantLabels(s string) int {
	c := 0
	if loc := label.IndexContentLabel(s); loc != nil {
		c += loc[1] - loc[0]
	}

	if loc := label.IndexFiller(s); loc != nil {
		c += loc[1] - loc[0]
	}

	return c
}
