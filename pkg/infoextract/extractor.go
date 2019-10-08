package infoextract

import (
	"context"
	"errors"
)

var (
	ErrURIInvalid = errors.New("uri invalid")
)

type Extractor interface {
	ExtractorID() string

	Extract(ctx context.Context, uri string) (interface{}, error)
}
