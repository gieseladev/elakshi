package api

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/errutils"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
)

// Core contains important API fields.
type Core struct {
	DB             *gorm.DB
	LyricsSearcher search.Searcher

	SpotifyClient *spotify.Client
}

// Close closes the core.
func (c *Core) Close() error {
	return errutils.CollectErrors(
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
