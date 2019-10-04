package edb

import "time"

type AudioSource struct {
	DBModel

	Type string `gorm:"UNIQUE_INDEX:uix_uri_type;NOT NULL"`
	URI  string `gorm:"UNIQUE_INDEX:uix_uri_type;NOT NULL"`

	TrackSources []TrackSource
}

type TrackSource struct {
	DBModel

	AudioSourceID uint64 `gorm:"NOT NULL"`
	TrackID       uint64 `gorm:"NOT NULL"`
	Track         Track

	StartOffsetMS uint32 `gorm:"type:integer"`
	EndOffsetMS   uint32 `gorm:"type:integer"`
}

// Length returns the length of the track as a duration.
func (ts TrackSource) Length() time.Duration {
	return time.Duration(ts.EndOffsetMS-ts.StartOffsetMS) * time.Millisecond
}
