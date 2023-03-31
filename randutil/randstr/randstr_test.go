package randstr

import (
	"testing"
	"unicode/utf8"

	"github.com/pierrre/assert"
)

func TestGenerate(t *testing.T) {
	n := 15
	cs := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	s := Generate(n, cs, nil)
	resN := utf8.RuneCountInString(s)
	assert.Equal(t, resN, n)
	assert.AllocsPerRun(t, 100, func() {
		Generate(n, cs, nil)
	}, 1)
}
