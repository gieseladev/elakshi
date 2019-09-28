package main

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/api/http"
	"github.com/gieseladev/elakshi/pkg/edb"
	"github.com/gieseladev/glyrics/v3/pkg/search"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
)

func main() {
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
