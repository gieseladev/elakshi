package audiosrc

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/service"
)

type Result struct {
	TrackSource edb.TrackSource

	// TODO score...
}

type Searcher interface {
	service.Identifier

	// Search searches for audio sources for the given track.
	// The track needs to have all associations pre-loaded.
	// The results must only include valid results and must be ordered from
	// best to worst.
	Search(ctx context.Context, track edb.Track) ([]Result, error)

	GenerateTrackURI(source edb.TrackSource) string
}
