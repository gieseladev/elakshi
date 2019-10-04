package youtube

import (
	"context"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func NewClient(ctx context.Context, apiKey string) (*youtube.Service, error) {
	return youtube.NewService(ctx, option.WithAPIKey(apiKey))
}
