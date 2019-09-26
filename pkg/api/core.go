package api

import (
	"context"
	"github.com/jinzhu/gorm"
)

type Core struct {
	DB *gorm.DB
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
