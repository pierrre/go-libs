package randutil

import (
	"crypto/rand"
	"testing"
)

func TestReaderSource(t *testing.T) {
	rs := &ReaderSource{
		Reader: rand.Reader,
	}
	for i := 0; i < 100; i++ {
		v := rs.Int63()
		if v < 0 {
			t.Fatalf("negative: %d", v)
		}
	}
}
