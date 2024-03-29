package edb

type RadioStation struct {
	DBModel

	Name    string
	ImageID uint64 `gorm:"NOT NULL"`
	Image   Image

	Genres []Genre `gorm:"MANY2MANY:radio_genres"`
}

func (r RadioStation) Namespace() string {
	return "radio"
}
