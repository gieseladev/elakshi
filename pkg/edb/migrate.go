package edb

import (
	"github.com/gieseladev/elakshi/pkg/errutils"
	"github.com/jinzhu/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return errutils.CollectErrors(
		db.AutoMigrate(&Track{}).
			AddForeignKey("artist_id", "artists(id)", "SET NULL", "CASCADE").
			AddForeignKey("album_id", "albums(id)", "SET NULL", "CASCADE").
			Error,
		db.AutoMigrate(&Genre{}).
			AddForeignKey("parent_id", "genres(id)", "SET NULL", "CASCADE").
			Error,
		db.AutoMigrate(&Lyrics{}).
			AddForeignKey("track_id", "tracks(id)", "CASCADE", "CASCADE").
			Error,
		db.AutoMigrate(&RadioStation{}).
			AddForeignKey("image_id", "images(id)", "SET NULL", "CASCADE").
			Error,
		db.AutoMigrate(&Playlist{}).
			AddForeignKey("image_id", "images(id)", "SET NULL", "CASCADE").
			Error,
		db.AutoMigrate(&PlaylistTrack{}).
			AddForeignKey("track_id", "tracks(id)", "CASCADE", "CASCADE").
			AddForeignKey("playlist_id", "playlists(id)", "CASCADE", "CASCADE").
			Error,
		db.AutoMigrate(
			&ExternalRef{},
			&Image{},
			&AudioSource{},
			&Artist{}, &Album{},
		).Error,
		db.AutoMigrate(&TrackSource{}).
			AddForeignKey("source_id", "audio_sources(id)", "CASCADE", "CASCADE").
			AddForeignKey("track_id", "tracks(id)", "CASCADE", "CASCADE").
			Error,
		db.Table("track_artists").
			AddForeignKey("track_id", "tracks(id)", "CASCADE", "CASCADE").
			AddForeignKey("artist_id", "artists(id)", "CASCADE", "CASCADE").
			Error,
	)
}
