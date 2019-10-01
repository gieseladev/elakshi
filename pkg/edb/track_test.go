package edb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrack_AllArtists(t *testing.T) {
	a := Artist{Name: "artist A"}
	b := Artist{Name: "artist B"}
	c := Artist{Name: "artist C"}

	track := Track{
		Artist:            a,
		AdditionalArtists: []Artist{b, c},
	}

	assert.Equal(t, []Artist{a, b, c}, track.AllArtists())
}
