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
	for i := 0; i < 100; i++ {
		v := rs.Int63()
		assert.Positive(t, v)
	}
}
