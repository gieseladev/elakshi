package api

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/audiosrc"
	"github.com/gieseladev/elakshi/pkg/errutil"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"github.com/gieseladev/elakshi/pkg/service"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
)

// Core contains important data shared between the handlers.
type Core struct {
	DB             *gorm.DB
	LyricsSearcher search.Searcher

	ExtractorPool     *infoextract.ExtractorPool
	TrackSourceFinder *audiosrc.Finder
}

// Close closes the core.
func (c *Core) Close() error {
	return errutil.CollectErrors(
		c.DB.Close(),
	)
}

// AddServices uses the given services to create the extractor pool and track
// source finder. It will panic if either of them is already set, no database is
// set, or an invalid service type is passed.
func (c *Core) AddServices(services ...service.Identifier) {
	if c.ExtractorPool != nil || c.TrackSourceFinder != nil {
		panic("api/core: core already has an extractor pool or track source finder.")
	}

	if c.DB == nil {
		panic("api/core: database must be set to add services.")
	}

	var searchers []audiosrc.Searcher
	var extractors []infoextract.Extractor

	for _, s := range services {
		found := false

		if s, ok := s.(audiosrc.Searcher); ok {
			searchers = append(searchers, s)
			found = true
		}

		if s, ok := s.(infoextract.Extractor); ok {
			extractors = append(extractors, s)
			found = true
		}

		if !found {
			panic("api/core: unexpected service type passed")
		}
	}

	c.TrackSourceFinder = audiosrc.NewFinder(c.DB, searchers...)
	c.ExtractorPool = infoextract.CollectExtractors(extractors...)
}

type coreKey struct{}

// WithCore adds a Core  to a context.
func WithCore(ctx context.Context, core *Core) context.Context {
	return context.WithValue(ctx, coreKey{}, core)
}

// CoreFromContext extracts the Core from a context.
func CoreFromContext(ctx context.Context) *Core {
	return ctx.Value(coreKey{}).(*Core)
}
