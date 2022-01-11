package randutil

import (
	"testing"
	"unicode/utf8"
)

func TestString(t *testing.T) {
	n := 15
	cs := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	s := String(n, cs, nil)
	resN := utf8.RuneCountInString(s)
	if resN != n {
		t.Fatalf("unexpected characters count: got %d, want %d", resN, n)
	}
}
