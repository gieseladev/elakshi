package youtube

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"net/url"
)

func (yt *youtubeExtractor) ExtractorID() string {
	return ytServiceName
}

func (yt *youtubeExtractor) Extract(ctx context.Context, uri string) (interface{}, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, infoextract.ErrURIInvalid
	}

	videoID := u.Query().Get("v")
	if videoID == "" {
		return nil, infoextract.ErrURIInvalid
	}

	tracks, err := yt.GetTracks(ctx, videoID)
	if err == ErrIDInvalid {
		return nil, infoextract.ErrURIInvalid
	}

	return tracks, err
}
