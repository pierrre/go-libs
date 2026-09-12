package weakutil

import (
	"sync"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

type commonMap[K comparable, V any] struct {
	m syncutil.Map[K, V]
}

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

func loadWeak[M interface{ deleteEntry(key K, e E) bool }, T any, K comparable, E any](m M, key K, e E, wp weak.Pointer[T],
) (*T, bool) {
	value, alive := loadPointer(wp)
	if !alive {
		m.deleteEntry(key, e)
	}
	return value, alive
}
