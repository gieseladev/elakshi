package edb

import "github.com/jinzhu/gorm"

type Lyrics struct {
	DBModel

	TrackID   uint64 `gorm:"INDEX"`
	SourceURL string
	Text      string
}

// GetTrackLyrics retrieves the lyrics for a given track.
func GetTrackLyrics(db *gorm.DB, trackID uint64) (Lyrics, error) {
	var lyrics Lyrics
	err := db.Take(&lyrics, "track_id = ?", trackID).Error

	return lyrics, err
}
