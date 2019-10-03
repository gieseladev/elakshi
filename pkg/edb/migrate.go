package edb

import (
	"fmt"
	"github.com/gieseladev/elakshi/pkg/errutils"
	"github.com/jinzhu/gorm"
	"reflect"
)

// TODO some m2m relations require onDelete to be RESTRICT!

func m2mConstraint(db *gorm.DB, model interface{}, fieldName string) error {
	s := db.NewScope(model)
	field, ok := s.FieldByName(fieldName)
	if !ok {
		panic("edb/migrate: field not found")
	}

	relation := field.Relationship
	if relation == nil || relation.Kind != "many_to_many" {
		panic("edb/migrate: invalid relationship")
	}

	jtHandler := relation.JoinTableHandler.(*gorm.JoinTableHandler)

	jtName := jtHandler.Table(db)

	sfKeys := jtHandler.SourceForeignKeys()
	if len(sfKeys) != 1 {
		panic("edb/migrate: invalid amount of source foreign keys")
	}

	dfKeys := jtHandler.DestinationForeignKeys()
	if len(sfKeys) != 1 {
		panic("edb/migrate: invalid amount of destination foreign keys")
	}

	stName := s.QuotedTableName()
	dtName := db.NewScope(reflect.New(jtHandler.Destination.ModelType).Interface()).QuotedTableName()

	sfName := s.Quote(sfKeys[0].DBName)
	sfaName := s.Quote(sfKeys[0].AssociationDBName)

	dfName := s.Quote(dfKeys[0].DBName)
	dfaName := s.Quote(dfKeys[0].AssociationDBName)

	return db.Table(jtName).
		AddForeignKey(sfName, fmt.Sprintf("%s(%s)", stName, sfaName), "CASCADE", "CASCADE").
		AddForeignKey(dfName, fmt.Sprintf("%s(%s)", dtName, dfaName), "CASCADE", "CASCADE").
		Error
}

func autoMigrateArtist(db *gorm.DB) error {
	if err := db.AutoMigrate(&Artist{}).Error; err != nil {
		return err
	}

	return errutils.CollectErrors(
		m2mConstraint(db, &Artist{}, "Genres"),
		m2mConstraint(db, &Artist{}, "Images"),
		m2mConstraint(db, &Artist{}, "ExternalReferences"),
	)

}

func autoMigrateAlbum(db *gorm.DB) error {
	if err := db.AutoMigrate(&Album{}).Error; err != nil {
		return err
	}

	return errutils.CollectErrors(
		m2mConstraint(db, &Album{}, "Artists"),
		m2mConstraint(db, &Album{}, "Genres"),
		m2mConstraint(db, &Album{}, "Images"),
		m2mConstraint(db, &Album{}, "ExternalReferences"),
	)

}

func autoMigrateTrack(db *gorm.DB) error {
	err := db.AutoMigrate(&Track{}).
		AddForeignKey("artist_id", "artists(id)", "SET NULL", "CASCADE").
		AddForeignKey("album_id", "albums(id)", "SET NULL", "CASCADE").
		Error
	if err != nil {
		return err
	}

	err = errutils.CollectErrors(
		m2mConstraint(db, &Track{}, "AdditionalArtists"),
		m2mConstraint(db, &Track{}, "Genres"),
		m2mConstraint(db, &Track{}, "Images"),
		m2mConstraint(db, &Track{}, "ExternalReferences"),
	)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&AudioSource{}).Error; err != nil {
		return err
	}

	return db.AutoMigrate(&TrackSource{}).
		AddForeignKey("source_id", "audio_sources(id)", "CASCADE", "CASCADE").
		AddForeignKey("track_id", "tracks(id)", "CASCADE", "CASCADE").
		Error
}

func autoMigratePlaylist(db *gorm.DB) error {
	err := db.AutoMigrate(&Playlist{}).
		AddForeignKey("image_id", "images(id)", "SET NULL", "CASCADE").
		Error
	if err != nil {
		return err
	}

	return db.AutoMigrate(&PlaylistTrack{}).
		AddForeignKey("track_id", "tracks(id)", "CASCADE", "CASCADE").
		AddForeignKey("playlist_id", "playlists(id)", "CASCADE", "CASCADE").
		Error
}

func autoMigrateRadio(db *gorm.DB) error {
	err := db.AutoMigrate(&RadioStation{}).
		AddForeignKey("image_id", "images(id)", "SET NULL", "CASCADE").
		Error
	if err != nil {
		return err
	}

	return m2mConstraint(db, &RadioStation{}, "Genres")
}

func AutoMigrate(db *gorm.DB) error {
	err := errutils.CollectErrors(
		db.AutoMigrate(&Genre{}).
			AddForeignKey("parent_id", "genres(id)", "SET NULL", "CASCADE").
			Error,
		db.AutoMigrate(
			&ExternalRef{},
			&Image{},
		).Error,
	)
	if err != nil {
		return err
	}

	err = errutils.CollectErrors(
		autoMigrateArtist(db),
		autoMigrateAlbum(db),
	)
	if err != nil {
		return err
	}

	err = errutils.CollectErrors(
		autoMigrateRadio(db),
		autoMigrateTrack(db),
	)
	if err != nil {
		return err
	}

	return errutils.CollectErrors(
		autoMigratePlaylist(db),
		db.AutoMigrate(&Lyrics{}).
			AddForeignKey("track_id", "tracks(id)", "CASCADE", "CASCADE").
			Error,
	)
}
