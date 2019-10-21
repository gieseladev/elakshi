package songtitle

import (
	"github.com/gieseladev/elakshi/pkg/songtitle/bracket"
	"strings"
)

type Title struct {
	Raw string // Raw content of the title

	BaselineParts    []string
	OtherParts       []string
	GuestAppearances []string

	ContentLabels []string
}

func ParseTitle(title string) Title {
	parsed := Title{
		Raw: title,
	}

	baseline, others := splitBaselineParts(title)

	for i, part := range others {
		guests := getGuestAppearances(part)

		switch {
		case len(guests) != 0:
			parsed.GuestAppearances = append(parsed.GuestAppearances, guests...)
		case isContentLabel(part):
			parsed.ContentLabels = append(parsed.ContentLabels, part)
		case isFiller(part):
		default:
			continue
		}

		// remove the part from others
		others = deleteStringSlice(others, i)
	}

	// run feature detection

	// baseline may contain ft. at any point, but "other" must start with it.

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

func getGuestAppearances(s string) []string {
	// detect the guest thing

	if strings.HasPrefix(s, "feat.") {
		s = strings.TrimSpace(s[len("feat."):])
	} else {
		return nil
	}

	// TODO split by "&", ",", and others
	guests := SplitOnAnyRuneOf(s, []rune{'&', ','})
	mapStringSlice(guests, strings.TrimSpace)

	return guests
}

func isContentLabel(s string) bool {
	// TODO
	return false
}

func isFiller(s string) bool {
	// TODO
	return false
}
