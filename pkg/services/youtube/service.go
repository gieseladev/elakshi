package youtube

import (
	"context"
	"github.com/jinzhu/gorm"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	ytServiceName = "youtube"
)

type youtubeService struct {
	db      *gorm.DB
	service *youtube.Service
}

func FromClient(db *gorm.DB, service *youtube.Service) *youtubeService {
	return &youtubeService{
		db:      db,
		service: service,
	}
}

func FromAPIKey(ctx context.Context, db *gorm.DB, apiKey string) (*youtubeService, error) {
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	return FromClient(db, service), err
}

func (yt *youtubeService) ServiceID() string {
	return ytServiceName
}
