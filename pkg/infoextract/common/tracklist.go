package common

import "time"

// TODO move to standalone library

type TracklistTrackInfo struct {
	Track string

	StartOffset time.Duration
	EndOffset   time.Duration
}

func ExtractTracklistFromText(text string, trackLength time.Duration) []TracklistTrackInfo {
	return []TracklistTrackInfo{}
}
