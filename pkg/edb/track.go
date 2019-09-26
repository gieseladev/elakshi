package edb

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Track struct {
	DBModel

	// TODO statistics

	ISRC string `gorm:"UNIQUE"`

	Name        string
	Images      []Image  `gorm:"MANY2MANY:track_images"`
	Artists     []Artist `gorm:"MANY2MANY:track_artists"`
	AlbumID     uint64
	Album       Album
	Length      time.Duration
	ReleaseDate time.Time
	Genres      []Genre `gorm:"MANY2MANY:track_genres"`
}

func (track *Track) IsEmpty() bool {
	return track.ID != 0
}

func GetTrack(db *gorm.DB, trackID uint64) (bool, Track) {
	var track Track
	db.First(&track, trackID)

	return !track.IsEmpty(), track
}

type Artist struct {
	DBModel

	Name   string
	Images []Image `gorm:"MANY2MANY:artist_images"`

	StartDate time.Time
	EndDate   time.Time
	Genres    []Genre `gorm:"MANY2MANY:artist_genres"`
}

type Album struct {
	DBModel

	Name    string
	Images  []Image  `gorm:"MANY2MANY:album_images"`
	Artists []Artist `gorm:"MANY2MANY:album_artists"`

	// TODO length? Calculate dynamically or update on user change

	ReleaseDate time.Time
	Genres      []Genre `gorm:"MANY2MANY:album_genres"`
}

type Genre struct {
	DBModel

	Name string `gorm:"UNIQUE_INDEX:uix_genre"`

	ParentID uint64 `gorm:"UNIQUE_INDEX:uix_genre"`
	Parent   *Genre
}
