package api

import (
	"context"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/glyrics/v3"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
	"strings"
)

var (
	ErrEIDInvalid  = errors.New("eid invalid")
	ErrEIDNotFound = errors.New("eid was not found")
)

var NoPaddingEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// EncodeEID encodes an id into an eid.
func EncodeEID(id uint64) string {
	var data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, id)

	return NoPaddingEncoding.EncodeToString(data)
}

// DecodeEID converts the encoded id into its integer representation.
// Returns ErrEIDInvalid if the eid is invalid.
func DecodeEID(eid string) (uint64, error) {
	if NoPaddingEncoding.DecodedLen(len(eid)) > 8 {
		return 0, ErrEIDInvalid
	}

	data := make([]byte, 8)
	_, err := NoPaddingEncoding.Decode(data, []byte(eid))
	if err != nil {
		return 0, ErrEIDInvalid
	}

	return binary.LittleEndian.Uint64(data), nil
}

func GetTrack(db *gorm.DB, eid string) (edb.Track, error) {
	trackID, err := DecodeEID(eid)
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
	trackID, err := DecodeEID(eid)
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
