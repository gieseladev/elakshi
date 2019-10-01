package edb

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Track struct {
	DBModel

	// TODO statistics

	ISRC string `gorm:"UNIQUE"`

	Name              string
	Images            []Image `gorm:"MANY2MANY:track_images"`
	ArtistID          uint64
	Artist            Artist
	AdditionalArtists []Artist `gorm:"MANY2MANY:track_artists"`
	AlbumID           uint64
	Album             Album
	LengthMS          uint32 `gorm:"type:integer"`
	ReleaseDate       time.Time
	Genres            []Genre `gorm:"MANY2MANY:track_genres"`
}

// AllArtists a slice containing all artists.
func (track *Track) AllArtists() []Artist {
	artists := make([]Artist, len(track.AdditionalArtists)+1)
	artists[0] = track.Artist
	copy(artists[1:], track.AdditionalArtists)

	return artists
}

// Length returns the length of the track as a duration.
func (track *Track) Length() time.Duration {
	return time.Duration(track.LengthMS) * time.Millisecond
}

func GetTrack(db *gorm.DB, trackID uint64) (Track, error) {
	var track Track
	err := db.Take(&track, trackID).Error

	return track, err
}

type Artist struct {
	DBModel

	Name   string
	Images []Image `gorm:"MANY2MANY:artist_images"`

	StartDate time.Time
	EndDate   time.Time
	Genres    []Genre `gorm:"MANY2MANY:artist_genres"`
}

// IsEmpty checks whether the artist is empty.
func (a *Artist) IsEmpty() bool {
	return a == nil || a.Name == ""
}

type Album struct {
	DBModel

	Name   string
	Images []Image `gorm:"MANY2MANY:album_images"`
	// TODO artists order?
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

const allParentGenres = `
WITH RECURSIVE cte_parent_genre AS (
	SELECT id, name, parent_id
	FROM genres
	WHERE id = ?
	
	UNION
	
	SELECT p.id, p.name, p.parent_id
	FROM genres p
		INNER JOIN cte_parent_genre c
		ON p.id = c.parent_id
)

SELECT *
FROM cte_parent_genre;
`
const allSubGenres = `
WITH RECURSIVE cte_subgenre AS (
    SELECT id, name, parent_id
    FROM genres
    WHERE id = ?

    UNION

    SELECT c.id, c.name, c.parent_id
    FROM genres c
             INNER JOIN cte_subgenre p ON p.id = c.parent_id
)

SELECT *
FROM cte_subgenre;
`

func GetParentGenres(db *gorm.DB, genreID uint64) ([]Genre, error) {
	var genres []Genre
	err := db.Raw(allParentGenres, genreID).Scan(genres).Error
	return genres, err
}

func GetSubGenres(db *gorm.DB, genreID uint64) ([]Genre, error) {
	var genres []Genre
	err := db.Raw(allSubGenres, genreID).Scan(genres).Error
	return genres, err
}
