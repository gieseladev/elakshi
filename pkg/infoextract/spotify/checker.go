package spotify

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidSpotifyURI = errors.New("invalid spotify uri")
)

func parseURI(uri string) (typ string, id string, err error) {
	if strings.HasPrefix(uri, "spotify:") {
		parts := strings.Split(uri, ":")
		if len(parts) != 3 {
			err = ErrInvalidSpotifyURI
			return
		}

		typ, id = parts[1], parts[2]
		return
	}

	u, err := url.Parse(uri)
	if err != nil {
		return
	}

	// parts = ["" / "typ" / "id"]
	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 {
		err = ErrInvalidSpotifyURI
		return
	}

	typ, id = parts[1], parts[2]
	return
}

func (s *spotifyExtractor) URLHostnames() []string {
	return []string{"open.spotify.com"}
}

func (s *spotifyExtractor) CheckURI(uri string) bool {
	if !strings.HasPrefix(uri, "spotify:") {
		return false
	}

	parts := strings.Split(uri, ":")
	return len(parts) == 3
}
