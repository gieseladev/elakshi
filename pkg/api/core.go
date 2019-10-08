package api

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/errutil"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
)

// Core contains important API fields.
type Core struct {
	DB             *gorm.DB
	LyricsSearcher search.Searcher

	ExtractorPool *infoextract.ExtractorPool
}

// Close closes the core.
func (c *Core) Close() error {
	return errutil.CollectErrors(
		c.DB.Close(),
	)
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
