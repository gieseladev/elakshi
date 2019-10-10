package api

import (
	"errors"
	"github.com/gieseladev/elakshi/pkg/edb"
)

const (
	// TrackType is the type name of track response.
	TrackType = "track"
)

// ResolveResp represents the response given to resolve requests.
type ResolveResp struct {
	Type string   `json:"type"`
	EIDs []string `json:"eids"`
}

// CreateResolveResponse returns a SimpleResolveResp instance for the given
// value.
func CreateResolveResponse(v interface{}) (resp ResolveResp, err error) {
	switch v := v.(type) {
	case edb.Track:
		resp.Type = TrackType
		resp.EIDs = []string{v.EID()}
	case []edb.Track:
		resp.Type = TrackType
		resp.EIDs = make([]string, len(v))
		for i, track := range v {
			resp.EIDs[i] = track.EID()
		}
	default:
		err = errors.New("unknown value type passed")
	}

	return
}

// AudioSourceResp represents an audio source response.
type AudioSourceResp struct {
	Source      string  `json:"source"`
	Identifier  string  `json:"identifier"`
	URI         string  `json:"uri"`
	StartOffset float64 `json:"start_offset"`
	EndOffset   float64 `json:"end_offset"`

	IsLive bool `json:"is_live"`
}

// AudioSourceRespFromTrackSource builds an AudioSourceResp from a
// edb.TrackSource. The track source is expected to have the audio source
// loaded.
func AudioSourceRespFromTrackSource(trackSource edb.TrackSource) AudioSourceResp {
	audioSource := trackSource.AudioSource

	return AudioSourceResp{
		Source:     audioSource.Type,
		Identifier: audioSource.URI,
		// TODO uri?
		URI:         "",
		StartOffset: float64(trackSource.StartOffsetMS) / 1000,
		EndOffset:   float64(trackSource.EndOffsetMS) / 1000,
		IsLive:      false,
	}
}
