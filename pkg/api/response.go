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

// CreateResolveResponse returns a SimpleresolveResp instance for the given
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
