package stringcmp

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

func TestSplitParts(t *testing.T) {
	a := assert.New(t)

	a.Equal([]string{"Binary Star", "moll", "feat. Uru"},
		SplitParts("Binary Star (feat. Uru) - moll"))
}

func BenchmarkSplitParts(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = SplitParts("Binary Star (feat. Uru) - moll")
	}
}
