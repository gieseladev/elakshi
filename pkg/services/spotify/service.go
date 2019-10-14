package spotify

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	spotifyServiceName = "spotify"
)

// TODO tracks should also resolve cross references in the form of title - artist
// 		or title - album.

type spotifyService struct {
	db     *gorm.DB
	client *spotify.Client
}

func FromClient(db *gorm.DB, client *spotify.Client) *spotifyService {
	return &spotifyService{db: db, client: client}
}

func FromToken(ctx context.Context, db *gorm.DB, clientID, clientSecret string) (*spotifyService, error) {
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotify.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	client := spotify.Authenticator{}.NewClient(token)
	client.AutoRetry = true

	return FromClient(db, &client), nil
}
