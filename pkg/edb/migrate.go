package edb

import (
	"github.com/jinzhu/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Image{},
		&Lyrics{},
		&Playlist{}, &PlaylistTrack{},
		&RadioStation{},
		&AudioSource{}, &TrackSource{},
		&Track{}, &Artist{}, &Album{}, &Genre{},
	).Error
}
