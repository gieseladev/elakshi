package edb

import "time"

type AudioSource struct {
	DBModel

	Type string
	URI  string `gorm:"UNIQUE_INDEX"`

	TrackSources []TrackSource
}

type TrackSource struct {
	DBModel

	SourceID uint64
	TrackID  uint64
	Track    Track

	StartOffsetMS uint32 `gorm:"type:integer"`
	EndOffsetMS   uint32 `gorm:"type:integer"`
}

// Length returns the length of the track as a duration.
func (ts TrackSource) Length() time.Duration {
	return time.Duration(ts.EndOffsetMS-ts.StartOffsetMS) * time.Millisecond
}
