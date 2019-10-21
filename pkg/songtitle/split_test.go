package songtitle

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitOnDash(t *testing.T) {
	a := assert.New(t)

	a.Equal([]string{"hello ", " world"}, SplitOnDash("hello - world"))

	a.Equal([]string{"hello- world"}, SplitOnDash("hello- world"))
	a.Equal([]string{"hello -world"}, SplitOnDash("hello -world"))

	a.Equal([]string{"hello ", " world-", "-how ", " are you?"},
		SplitOnDash("hello - world---how - are you?"))
}

func TestSplitOnAnyRuneOf(t *testing.T) {
	a := assert.New(t)

	a.Equal([]string{"hello", " world", " What"},
		SplitOnAnyRuneOf("hello, world. What", []rune{',', '.'}))
}
