package bidi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitBracketGroupContent(t *testing.T) {
	a := assert.New(t)

	a.Equal([]string{"Hello World"}, SplitBracketGroupContent("Hello World"))
	a.Equal([]string{"Binary Star  - moll", "feat. Uru"}, SplitBracketGroupContent("Binary Star (feat. Uru) - moll"))

	a.Equal([]string{"((b", "a", "("}, SplitBracketGroupContent("(((a[(])b"))
	a.Equal([]string{" f", " b  d ", "a", "c", "e"}, SplitBracketGroupContent("((a) b (c) d (e)) f"))

	a.Equal(
		[]string{"agh]l", "", "bc", "de", "f", "i[jk"},
		SplitBracketGroupContent("a[(bc[de]{f})]gh(i[jk)]l"),
	)

	a.Equal(
		[]string{"5 * ", "2 *  + ", "1 + 2", "3 / 5"},
		SplitBracketGroupContent("5 * [2 * (1 + 2) + (3 / 5)]"),
	)
}
