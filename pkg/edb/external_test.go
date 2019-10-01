package edb

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func getDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	require.NoError(t, AutoMigrate(db))
	return db
}

func TestGetModelByExternalRef(t *testing.T) {
	assert := assert.New(t)

	db := getDB(t)

	track := Track{
		Name: "B12",
		ExternalReferences: []ExternalRef{{
			Service:    "spotify",
			Identifier: "abc",
		}},
	}
	require.NoError(t, db.Create(&track).Error)

	var tr Track
	err := db.Scopes(GetModelByExternalRef(tr, "spotify", "abc")).Scan(&tr).Error
	assert.NoError(err)
	assert.Equal(tr.Name, "B12")
}
