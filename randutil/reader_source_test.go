package randutil

import (
	"crypto/rand"
	"testing"

	"github.com/pierrre/assert"
)

func TestReaderSource(t *testing.T) {
	rs := &ReaderSource{
		Reader: rand.Reader,
	}
	for range 100 {
		v := rs.Int63()
		assert.Positive(t, v)
	}
}
