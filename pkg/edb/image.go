package edb

type Image struct {
	DBModel

	SourceURI string `gorm:"INDEX"`
	URI       string
}

func (i Image) Namespace() string {
	return "image"
}
