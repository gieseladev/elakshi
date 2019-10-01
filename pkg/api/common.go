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
	ErrEIDNotFound = errors.New("eid was not found")
)

func GetTrack(db *gorm.DB, eid string) (edb.Track, error) {
	trackID, err := edb.DecodeEID(eid)
	if err != nil {
		return edb.Track{}, err
	}

	track, err := edb.GetTrack(db, trackID)
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
func GetTrackLyrics(ctx context.Context, eid string) (edb.Lyrics, error) {
	trackID, err := edb.DecodeEID(eid)
	if err != nil {
		return edb.Lyrics{}, err
	}

	core := CoreFromContext(ctx)
	if core == nil {
		return edb.Lyrics{}, errors.New("context without api core passed")
	}

	lyrics, err := edb.GetTrackLyrics(core.DB, trackID)
	if err == nil {
		return lyrics, nil
	} else if !gorm.IsRecordNotFoundError(err) {
		return edb.Lyrics{}, err
	}

	var track edb.Track
	if err := core.DB.Preload("Artist").Take(&track).Error; err != nil {
		return edb.Lyrics{}, err
	}

	if lyrics, ok := findLyrics(ctx, core.LyricsSearcher, track); ok {
		core.DB.Create(&lyrics)
		return lyrics, nil
	}

	return edb.Lyrics{}, ErrLyricsNotFound
}
