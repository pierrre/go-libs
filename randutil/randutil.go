// Package randutil provides random related utilities.
//
// Importing this package initializes the global math/rand seed with a truly random value.
package randutil

import (
	crypto_rand "crypto/rand"
	"math/rand"
)

func init() {
	src := &ReaderSource{
		crypto_rand.Reader,
	}
	seed := int64(src.Uint64())
	rand.Seed(seed)
}
