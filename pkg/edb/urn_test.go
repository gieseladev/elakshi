package edb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseURN(t *testing.T) {
	assert := assert.New(t)

	_, err := ParseURN("anything really")
	assert.Error(ErrInvalidURN, err)

	_, err = ParseURN("spotify:test:1")
	assert.Error(ErrInvalidNID, err)

	u, err := ParseURN("elakshi:track:AE")
	if assert.NoError(err) {
		assert.Equal(NewElakshiURN("track", "AE"), u)

		id, err := u.DecodeEID()
		if assert.NoError(err) {
			assert.Equal(id, uint64(1))
		}
	}
}

func TestURNFromParts(t *testing.T) {
	model := Track{
		DBModel: DBModel{
			ID: 5552,
		},
	}

	u := URNFromParts(model)
	assert.Equal(t, u.URN(), "elakshi:track:WAKQAAAAAAAAA")
}
