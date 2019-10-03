package edb

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetModelByExternalRef(t *testing.T) {
	assert := assert.New(t)

	db := createDB(t)

	track := Track{
		Name: "B12",
		ExternalReferences: []ExternalRef{{
			Service:    "spotify",
			Identifier: "abc",
		}},
	}
	require.NoError(t, db.Create(&track).Error)

	var tr Track
	found, err := GetModelByExternalRef(db, "spotify", "abc", &tr)
	assert.NoError(err)
	assert.True(found)
	assert.Equal(tr.Name, "B12")
}
