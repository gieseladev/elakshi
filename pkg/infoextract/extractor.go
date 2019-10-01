package infoextract

import "context"

type Extractor interface {
	ExtractorID() string

	CanExtract(ctx context.Context, uri string) (bool, error)

	Extract(ctx context.Context, uri string) error
}
