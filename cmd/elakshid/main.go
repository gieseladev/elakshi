package main

import (
	"context"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/api/http"
	"github.com/gieseladev/elakshi/pkg/api/wamp"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/service"
	"github.com/gieseladev/elakshi/pkg/services/spotify"
	"github.com/gieseladev/elakshi/pkg/services/youtube"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
)

// TODO when finding an audio source which already contains a track spanning
//  the entire duration, link the tracks!

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

func getServices(db *gorm.DB) []service.Identifier {
	ctx := context.Background()
	var services []service.Identifier

	yt, err := youtube.FromAPIKey(ctx, db, os.Getenv("YOUTUBE_API_KEY"))
	if err != nil {
		panic(err)
	}

	services = append(services, yt)

	spotifyService, err := spotify.FromToken(ctx, db, os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		panic(err)
	}

	services = append(services, spotifyService)

	return services
}

func getCore() *api.Core {
	lyricsSearcher := &search.Google{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	}

	db := getDB()

	core := &api.Core{
		DB:             db,
		LyricsSearcher: lyricsSearcher,
	}

	core.AddServices(getServices(db)...)

	return core
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
