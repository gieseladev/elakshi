package audiosrc

import (
	"context"
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/jinzhu/gorm"
)

type Finder struct {
	db       *gorm.DB
	searcher Searcher
}

func NewFinder(db *gorm.DB, searchers ...Searcher) *Finder {
	return &Finder{
		db:       db,
		searcher: CollectSearchers(searchers...),
	}
}

func (f *Finder) SearchOne(ctx context.Context, track edb.Track) (Result, error) {
	results, err := f.searcher.Search(ctx, track)
	if err != nil {
		return Result{}, err
	}

	if len(results) == 0 {
		return Result{}, errors.New("no results found")
	}

	for i, result := range results {
		trackSource := &result.TrackSource
		audioSource := trackSource.AudioSource

		if f.db.NewRecord(audioSource) {
			err := f.db.FirstOrCreate(audioSource, edb.AudioSource{
				Type: audioSource.Type,
				URI:  audioSource.URI,
			}).Error
			if err != nil {
				return Result{}, err
			}
		}

		if f.db.NewRecord(trackSource) {
			err := f.db.
				Set("gorm:association_autoupdate", false).
				Create(&trackSource).
				Error
			if err != nil {
				return Result{}, err
			}
		}

		results[i] = result
	}

	return results[0], nil
}

func (f *Finder) GetTrackSource(ctx context.Context, trackID uint64) (edb.TrackSource, error) {
	trackSource, err := edb.GetTrackSource(f.db, trackID)
	if err == nil {
		return trackSource, err
	} else if !gorm.IsRecordNotFoundError(err) {
		return edb.TrackSource{}, err
	}

	var track edb.Track
	err = f.db.
		Set("gorm:auto_preload", true).
		Take(&track, trackID).Error
	if err != nil {
		return edb.TrackSource{}, err
	}

	result, err := f.SearchOne(ctx, track)
	if err != nil {
		return edb.TrackSource{}, err
	}

	return result.TrackSource, nil
}
