package youtube

import (
	"context"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func NewClient(apiKey string) (*youtube.Service, error) {
	return youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
}
