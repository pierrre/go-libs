package randutil

import (
	"testing"
	"unicode/utf8"

	"github.com/pierrre/assert"
)

func TestString(t *testing.T) {
	n := 15
	cs := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	s := String(n, cs, nil)
	resN := utf8.RuneCountInString(s)
	assert.Equal(t, resN, n)
}
