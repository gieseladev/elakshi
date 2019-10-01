package edb

type PlaylistTrack struct {
	DBModel

	TrackID uint64
	Track   Track
	Author  uint64

	PlaylistID uint64
}

type Playlist struct {
	DBModel

	Name    string
	ImageID uint64
	Image   Image

	Tracks []PlaylistTrack
}

func (p Playlist) Namespace() string {
	return "playlist"
}
