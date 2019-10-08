// +build integration

package edb

import (
	"flag"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/require"
	"testing"
)

var posgresDSN = flag.String("postgres-dsn", "user=postgres sslmode=disable", "Postgres connection string")

func createDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open("postgres", *posgresDSN)
	require.NoError(t, err)

	require.NoError(t, AutoMigrate(db))

	return db
}
