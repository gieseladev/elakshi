package spotify

import (
	"context"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
)

// NewClient creates a new spotify.Client authenticated using client credentials
// flow.
func NewClient(ctx context.Context, clientID, clientSecret string) (*spotify.Client, error) {
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

	return &client, nil
}
