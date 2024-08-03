// Package randstr allows to build random string.
package randstr

import (
	"math/rand"

	"github.com/pierrre/go-libs/bufpool"
)

// Generate generates a random string from the given length and characters set.
//
// If the provided [rand.Rand] is nil, the default global instance is used.
func Generate(n int, cs []rune, r *rand.Rand) string {
	intn := rand.Intn
	if r != nil {
		intn = r.Intn
	}
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	for range n {
		buf.WriteRune(cs[intn(len(cs))])
	}
	return buf.String()
}

var bufPool = bufpool.Pool{}
