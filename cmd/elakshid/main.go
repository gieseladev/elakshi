package main

import (
	"context"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/api/http"
	"github.com/gieseladev/elakshi/pkg/api/wamp"
	"github.com/gieseladev/elakshi/pkg/audiosrc"
	ytsearch "github.com/gieseladev/elakshi/pkg/audiosrc/youtube"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/elakshi/pkg/infoextract"
	"github.com/gieseladev/elakshi/pkg/infoextract/spotify"
	ytinfo "github.com/gieseladev/elakshi/pkg/infoextract/youtube"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"google.golang.org/api/youtube/v3"
	"log"
	"os"
	"sync"
)

// TODO move service libraries to separate services package!
//		/services/youtube
//		/services/spotify
// 		These can then implement all interfaces without having to do the awkward
//		client sharing.

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

var ytMux sync.Mutex
var ytClient *youtube.Service

func getYoutubeClient() *youtube.Service {
	ytMux.Lock()
	defer ytMux.Unlock()

	if ytClient == nil {
		youtubeClient, err := ytinfo.NewClient(context.Background(), os.Getenv("YOUTUBE_API_KEY"))
		if err != nil {
			panic(err)
		}
		ytClient = youtubeClient
	}

	return ytClient
}

func getExtractorPool(db *gorm.DB) *infoextract.ExtractorPool {
	pool := &infoextract.ExtractorPool{}

	pool.AddExtractors(ytinfo.NewExtractor(db, getYoutubeClient()))

	spotifyClient, err := spotify.NewClient(context.Background(), os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		panic(err)
	}
	pool.AddExtractors(spotify.NewExtractor(db, spotifyClient))

	return pool
}

func getFinder(db *gorm.DB) *audiosrc.Finder {
	ytSearcher := ytsearch.New(getYoutubeClient())

	return audiosrc.NewFinder(db, ytSearcher)
}

func getCore() *api.Core {
	lyricsSearcher := &search.Google{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	}

	db := getDB()

	return &api.Core{
		DB:             db,
		LyricsSearcher: lyricsSearcher,

		ExtractorPool:     getExtractorPool(db),
		TrackSourceFinder: getFinder(db),
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
