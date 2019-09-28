package youtube

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseYTDuration(t *testing.T) {
	assert := assert.New(t)

	d, err := parseYTDuration("PT5M3S")
	if assert.NoError(err) {
		assert.Equal(d, 5*time.Minute+3*time.Second)
	}
}
