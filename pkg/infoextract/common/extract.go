package common

import (
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/jinzhu/gorm"
	"strings"
)

// TODO don't create genres here! Same goes for images as they're overwritten
//  later on anyway.

// GetGenresByName returns genres for the given names and creates new ones
// for those not found.
func GetGenresByName(db *gorm.DB, names ...string) ([]edb.Genre, error) {
	for i, name := range names {
		names[i] = strings.ToLower(name)
	}

	existingGenres := make([]edb.Genre, 0, len(names))
	err := db.Where("name in (?)", names).Find(&existingGenres).Error
	if err != nil {
		return nil, err
	}

	// if we found all genres we can exit early
	if len(existingGenres) == len(names) {
		return existingGenres, nil
	}

	// build a set of all genre names we already found
	foundGenres := map[string]struct{}{}
	for _, genre := range existingGenres {
		foundGenres[genre.Name] = struct{}{}
	}

	// find all genres that don't already exist and create them
	for _, name := range names {
		if _, found := foundGenres[name]; found {
			continue
		}
		genre := edb.Genre{
			Name: name,
		}

		if err := db.Create(&genre).Error; err != nil {
			return nil, err
		}

		existingGenres = append(existingGenres, genre)
	}

	return existingGenres, nil
}
