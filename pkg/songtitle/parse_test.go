package songtitle

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type parseTestCase struct {
	Input string

	BaselineParts    []string
	OtherParts       []string
	GuestAppearances []string
	ContentLabels    []string

	NameOverwrite string
}

func (c *parseTestCase) Name() string {
	if c.NameOverwrite == "" {
		c.NameOverwrite = c.Input
	}

	return c.NameOverwrite
}

func assertSlice(t *testing.T, expected, actual interface{}) bool {
	if reflect.ValueOf(expected).Len() > 0 {
		return assert.Equal(t, expected, actual)
	} else {
		return assert.Empty(t, actual)
	}
}

func (c *parseTestCase) Run(t *testing.T) {
	actual := ParseTitle(c.Input)

	assertSlice(t, c.BaselineParts, actual.BaselineParts)
	assertSlice(t, c.GuestAppearances, actual.GuestAppearances)
	assertSlice(t, c.OtherParts, actual.OtherParts)
	assertSlice(t, c.ContentLabels, actual.ContentLabels)
}

func TestParseTitle(t *testing.T) {
	cases := []parseTestCase{
		{
			Input:            "Ed Sheeran - South of the Border (feat. Camila Cabello & Cardi B) [Official Video]",
			BaselineParts:    []string{"Ed Sheeran", "South of the Border"},
			GuestAppearances: []string{"Camila Cabello", "Cardi B"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name(), testCase.Run)
	}
}
