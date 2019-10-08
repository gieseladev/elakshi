package main

import (
	"context"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/api/http"
	"github.com/gieseladev/elakshi/pkg/api/wamp"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"github.com/gieseladev/elakshi/pkg/infoextract/spotify"
	"github.com/gieseladev/elakshi/pkg/infoextract/youtube"
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

	db.LogMode(true)

	if err := edb.AutoMigrate(db); err != nil {
		panic(err)
	}

	return db
}

func getExtractorPool(db *gorm.DB) *infoextract.ExtractorPool {
	pool := &infoextract.ExtractorPool{}

	youtubeClient, err := youtube.NewClient(context.Background(), os.Getenv("YOUTUBE_API_KEY"))
	if err != nil {
		panic(err)
	}
	pool.AddExtractors(youtube.NewExtractor(db, youtubeClient))

	spotifyClient, err := spotify.NewClient(context.Background(), os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		panic(err)
	}
	pool.AddExtractors(spotify.NewExtractor(db, spotifyClient))

	return pool
}

func getCore() *api.Core {
	lyricsSearcher := &search.Google{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	}

	db := getDB()

	return &api.Core{
		DB:             db,
		LyricsSearcher: lyricsSearcher,

		ExtractorPool: getExtractorPool(db),
	}
}

func getHandler(ctx context.Context) api.Handler {
	c, err := client.ConnectNet(ctx, os.Getenv("WAMP_ROUTER_URL"), client.Config{
		Realm: os.Getenv("WAMP_REALM"),
		Debug: true,
	})
	if err != nil {
		panic(err)
	}

	return api.CollectHandlers(
		http.NewHTTPHandler(ctx, ":8800"),
		wamp.NewWAMPHandler(ctx, c),
	)
}

func main() {
	core := getCore()
	defer func() { _ = core.Close() }()

	ctx := api.WithCore(context.Background(), core)

	handler := getHandler(ctx)

	if err := handler.Start(); err != nil {
		panic(err)
	}

	log.Println("server running")
	<-handler.Done()
}
