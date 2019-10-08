package spotify

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/infoextract"
)

func (s *spotifyExtractor) ExtractorID() string {
	return spotifyServiceName
}

func (s *spotifyExtractor) Extract(ctx context.Context, uri string) (interface{}, error) {
	typ, id, err := parseURI(uri)
	if err == ErrInvalidSpotifyURI {
		return nil, infoextract.ErrURIInvalid
	} else if err != nil {
		return nil, err
	}

	switch typ {
	case "track":
		// TODO handle spotify 404 errors and resolve to uri invalid or something
		return s.GetTrack(ctx, id)
	default:
		// TODO global invalid "type" error
		return nil, infoextract.ErrURIInvalid
	}
}
