package edb

import "github.com/jinzhu/gorm"

type RadioStation struct {
	DBModel

	Name    string
	ImageID uint64
	Image   Image

	Genres []Genre `gorm:"MANY2MANY:radio_genres"`
	A      gorm.Model
}
