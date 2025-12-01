package syncutil

import (
	"github.com/go4org/hashtriemap"
)

// Map is a typed implementation of [sync.Map].
type Map[K comparable, V any] = hashtriemap.HashTrieMap[K, V]
