package bracket

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitBracketGroupContent(t *testing.T) {
	a := assert.New(t)

	a.Equal([]string{"Hello World"}, ExtractContents("Hello World"))
	a.Equal([]string{"Binary Star  - moll", "feat. Uru"}, ExtractContents("Binary Star (feat. Uru) - moll"))

	a.Equal([]string{"((b", "a", "("}, ExtractContents("(((a[(])b"))
	a.Equal([]string{" f", " b  d ", "a", "c", "e"}, ExtractContents("((a) b (c) d (e)) f"))

	a.Equal(
		[]string{"agh]l", "", "bc", "de", "f", "i[jk"},
		ExtractContents("a[(bc[de]{f})]gh(i[jk)]l"),
	)

	a.Equal(
		[]string{"5 * ", "2 *  + ", "1 + 2", "3 / 5"},
		ExtractContents("5 * [2 * (1 + 2) + (3 / 5)]"),
	)

	a.Equal([]string{"これはです", "まれ"}, ExtractContents("これは⦓まれ⦔です"))
}

func BenchmarkSplitBracketGroupContent(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ExtractContents("5 * [2 * (1 + 2) + (3 / 5)]")
	}
}
