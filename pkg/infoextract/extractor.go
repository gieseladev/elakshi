package infoextract

import (
	"context"
	"errors"
	"github.com/gieseladev/elakshi/pkg/service"
)

var (
	ErrURIInvalid = errors.New("uri invalid")
)

type Extractor interface {
	service.Identifier

	Extract(ctx context.Context, uri string) (interface{}, error)
}
