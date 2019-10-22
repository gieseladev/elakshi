// Code generated using tools/genwordmap on 22 Oct 19 16:30 UTC! DO NOT EDIT.
// Args: --package label CONTENT_LABELS FILLER GUEST_APPEARANCE

package label

import "strings"

var contentLabelTokens = map[string]struct{}{
	"1080p":   {},
	"4k":      {},
	"720p":    {},
	"full hd": {},
	"hd":      {},
	"uhd":     {},
}

func IsContentLabel(s string) bool {
	s = strings.ToLower(s)
	if _, found := contentLabelTokens[s]; found {
		return true
	}
	return false
}

var fillerTokens = map[string]struct{}{
	"official music video": {},
	"official video":       {},
	"official":             {},
	"original mix":         {},
	"original":             {},
}

func IsFiller(s string) bool {
	s = strings.ToLower(s)
	if _, found := fillerTokens[s]; found {
		return true
	}
	return false
}

var guestAppearanceTokens = map[string]struct{}{
	"f.":        {},
	"f/":        {},
	"feat.":     {},
	"featuring": {},
	"ft.":       {},
	"with":      {},
}

func IsGuestAppearance(s string) bool {
	s = strings.ToLower(s)
	if _, found := guestAppearanceTokens[s]; found {
		return true
	}
	return false
}
