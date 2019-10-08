package youtube

import (
	"net/url"
	"strings"
)

func isYoutubeURL(u *url.URL) bool {
	hostname := strings.TrimPrefix(u.Hostname(), "www.")

	if !(hostname == "youtu.be" || strings.HasPrefix(hostname, "youtube.")) {
		return false
	}

	q := u.Query()
	return q.Get("v") != ""
}

func (yt *youtubeExtractor) URLHostnames() []string {
	return []string{"youtube.com", "youtu.be"}
}

func (yt *youtubeExtractor) CheckURL(u *url.URL) bool {
	return isYoutubeURL(u)
}
