package main

import (
	"context"
	"github.com/gieseladev/elakshi/pkg/api"
	"github.com/gieseladev/elakshi/pkg/api/http"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	db, err := gorm.Open("postgres", "")
	if err != nil {
		panic(err)
	}
	defer func() { _ = db.Close() }()

	ctx := api.WithCore(context.Background(), &api.Core{DB: db})

	handler := api.CollectHandlers(
		http.NewHTTPHandler(ctx, ":8800"),
	)

	if err := handler.Start(); err != nil {
		panic(err)
	}

	<-handler.Done()
}
