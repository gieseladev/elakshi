package api

import (
	"context"
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/glyrics/v3"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
	"strings"
)

var (
	ErrEIDNotFound       = errors.New("eid was not found")
	ErrNoExtractorForURI = errors.New("no extractor for uri found")
)

func (c *Core) GetTrack(eid string) (edb.Track, error) {
	trackID, err := edb.DecodeEID(eid)
	if err != nil {
		return edb.Track{}, err
	}

	track, err := edb.GetTrack(c.DB, trackID)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = ErrEIDNotFound
		}

		return track, err
	}

	return track, nil
}

func findLyrics(ctx context.Context, searcher search.Searcher, track edb.Track) (edb.Lyrics, bool) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(track.Name)

	if artist := track.Artist; artist.Name != "" {
		queryBuilder.WriteString(" - ")
		queryBuilder.WriteString(artist.Name)
	}

	// FIXME quality control!
	//		use the "search" method but maybe in reverse?
	info := glyrics.SearchFirst(ctx, searcher, queryBuilder.String())
	if info == nil {
		return edb.Lyrics{}, false
	}

	return edb.Lyrics{
		TrackID:   track.ID,
		SourceURL: info.URL,
		Text:      info.Lyrics,
	}, true
}

var (
	ErrLyricsNotFound = errors.New("lyrics not found")
)

// GetTrackLyrics retrieves lyrics for the track with the given EID.
// The context must contain the api core.
func (c *Core) GetTrackLyrics(ctx context.Context, eid string) (edb.Lyrics, error) {
	trackID, err := edb.DecodeEID(eid)
	if err != nil {
		return edb.Lyrics{}, err
	}

	lyrics, err := edb.GetTrackLyrics(c.DB, trackID)
	if err == nil {
		return lyrics, nil
	} else if !gorm.IsRecordNotFoundError(err) {
		return edb.Lyrics{}, err
	}

	var track edb.Track
	if err := c.DB.Preload("Artist").Take(&track).Error; err != nil {
		return edb.Lyrics{}, err
	}

	if lyrics, ok := findLyrics(ctx, c.LyricsSearcher, track); ok {
		err := c.DB.Create(&lyrics).Error
		return lyrics, err
	}

	return edb.Lyrics{}, ErrLyricsNotFound
}

func (c *Core) GetTrackSource(ctx context.Context, eid string) (AudioSourceResp, error) {
	trackID, err := edb.DecodeEID(eid)
	if err != nil {
		return AudioSourceResp{}, err
	}

	trackSource, err := c.TrackSourceFinder.GetTrackSource(ctx, trackID)

	if err != nil {
		return AudioSourceResp{}, err
	}

	return AudioSourceRespFromTrackSource(trackSource), nil
}

func (c *Core) ResolveURI(ctx context.Context, uri string) (interface{}, error) {
	extractor, ok := c.ExtractorPool.ResolveExtractor(uri)
	if !ok {
		return nil, ErrNoExtractorForURI
	}

	return extractor.Extract(ctx, uri)
}
