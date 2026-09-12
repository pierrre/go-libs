package weakutil

import (
	"sync"

	"github.com/pierrre/go-libs/syncutil"
)

type lazyValue[T any] struct {
	once  sync.Once
	value T
}

func (l *lazyValue[T]) get(compute func() T) T {
	l.once.Do(func() {
		l.value = compute()
	})
	return l.value
}

func clearMap[K comparable, V any](m *syncutil.Map[K, V], del func(K, V) bool) {
	for range 10 {
		var count int64
		m.Range(func(k K, v V) bool {
			count++
			del(k, v)
			return true
		})
		if count == 0 {
			return
		}
	}
}
