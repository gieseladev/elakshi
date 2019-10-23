// Code generated using tools/genwordmap on 23 Oct 19 14:10 UTC! DO NOT EDIT.
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
func IndexContentLabel(s string) []int {
	s = strings.ToLower(s)
	for token, _ := range contentLabelTokens {
		if i := strings.Index(s, token); i > -1 {
			return []int{i, i + len(token)}
		}
	}
	return nil
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
func IndexFiller(s string) []int {
	s = strings.ToLower(s)
	for token, _ := range fillerTokens {
		if i := strings.Index(s, token); i > -1 {
			return []int{i, i + len(token)}
		}
	}
	return nil
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
func IndexGuestAppearance(s string) []int {
	s = strings.ToLower(s)
	for token, _ := range guestAppearanceTokens {
		if i := strings.Index(s, token); i > -1 {
			return []int{i, i + len(token)}
		}
	}
	return nil
}
