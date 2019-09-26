package edb

import (
	"time"
)

type AudioSource struct {
	DBModel

	Type string
	URI  string `gorm:"UNIQUE"`
}

type TrackSource struct {
	DBModel

	SourceID uint64
	TrackID  uint64

	StartOffset time.Duration
	EndOffset   time.Duration
}
