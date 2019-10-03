package main

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/api/http"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract/spotify"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
)

func getDB() *gorm.DB {
	db, err := gorm.Open("postgres", "user=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err := edb.AutoMigrate(db); err != nil {
		panic(err)
	}

	return db
}

func getCore() *api.Core {
	lyricsSearcher := &search.Google{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	}

	spotifyClient, err := spotify.NewClient(context.Background(), os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		panic(err)
	}

	return &api.Core{
		DB:             getDB(),
		LyricsSearcher: lyricsSearcher,

		SpotifyClient: spotifyClient,
	}
}

func main() {
	core := getCore()
	defer func() { _ = core.Close() }()

	ctx := api.WithCore(context.Background(), core)

	handler := api.CollectHandlers(
		http.NewHTTPHandler(ctx, ":8800"),
	)

	if err := handler.Start(); err != nil {
		panic(err)
	}

	log.Println("server running")
	<-handler.Done()
}
