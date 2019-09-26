package edb

import "github.com/jinzhu/gorm"

type Lyrics struct {
	DBModel

	TrackID uint64 `gorm:"INDEX"`
	Text    string
}

func GetLyrics(db *gorm.DB, eid string) (bool, Lyrics) {
	var lyrics Lyrics
	db.First(&lyrics, "track_id = ?", eid)

	return lyrics != Lyrics{}, lyrics
}
