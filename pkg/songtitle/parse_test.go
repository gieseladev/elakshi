package songtitle

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTitle(t *testing.T) {
	a := assert.New(t)

	title := ParseTitle("Ed Sheeran - South of the Border (feat. Camila Cabello & Cardi B) [Official Video]")
	a.Equal("Ed Sheeran - South of the Border (feat. Camila Cabello & Cardi B) [Official Video]", title.Raw)
	a.Equal([]string{"Ed Sheeran", "South of the Border"}, title.BaselineParts)
	a.Empty(title.OtherParts)
	a.Equal([]string{"Camila Cabello", "Cardi B"}, title.GuestAppearances)
}
