package randutil

import (
	"math/rand"
	"strings"
)

// String generates a random string from the given length and characters set.
//
// If the provided *rand.Rand is nil, the default global instance is used.
func String(n int, cs []rune, r *rand.Rand) string {
	intn := rand.Intn
	if r != nil {
		intn = r.Intn
	}
	var sb strings.Builder
	sb.Grow(n) // Pre-allocate memory, more efficient for single byte characters.
	for i := 0; i < n; i++ {
		sb.WriteRune(cs[intn(len(cs))])
	}
	return sb.String()
}
