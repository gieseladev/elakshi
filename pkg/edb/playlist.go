package edb

type PlaylistTrack struct {
	DBModel

	TrackID  uint64 `gorm:"NOT NULL"`
	Track    Track
	AuthorID uint64 `gorm:"NOT NULL"`

	PlaylistID uint64 `gorm:"INDEX;NOT NULL"`
}

type Playlist struct {
	DBModel

	Name     string
	AuthorID uint64 `gorm:"NOT NULL"`

	ImageID *uint64
	Image   Image

	Tracks []PlaylistTrack
}

func (p Playlist) Namespace() string {
	return "playlist"
}
