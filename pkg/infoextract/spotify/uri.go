package spotify

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidSpotifyURI is returned for invalid Spotify URIs.
	ErrInvalidSpotifyURI = errors.New("invalid spotify uri")
)

func parseSpotifyURI(uri string) (typ string, id string, err error) {
	parts := strings.Split(uri, ":")
	if len(parts) != 3 {
		err = ErrInvalidSpotifyURI
		return
	}

	typ, id = parts[1], parts[2]
	return
}
