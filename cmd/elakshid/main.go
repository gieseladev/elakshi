package main

import (
	"context"
	"encoding/json"
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

func testSpotify() {
	client, err := spotify.NewClient(context.Background(), os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		panic(err)
	}

	extractor := spotify.NewExtractor(client)
	track, err := extractor.Test("6habFhsOp2NvshLv26DqMb")
	if err != nil {
		panic(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	_ = e.Encode(track)
}

func main() {
	testSpotify()

	db, err := gorm.Open("postgres", "user=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer func() { _ = db.Close() }()

	if err := edb.AutoMigrate(db); err != nil {
		panic(err)
	}

	lyricsSearcher := &search.Google{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	}

	ctx := api.WithCore(context.Background(), &api.Core{
		DB:             db,
		LyricsSearcher: lyricsSearcher,
	})

	handler := api.CollectHandlers(
		http.NewHTTPHandler(ctx, ":8800"),
	)

	if err := handler.Start(); err != nil {
		panic(err)
	}

	log.Println("server running")
	<-handler.Done()
}
