package weakutil

import (
	"sync"
	"sync/atomic"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

type commonMap[K comparable, V any] struct {
	m              syncutil.Map[K, V]
	initialized    sync.Once
	cleanupEnabled atomic.Bool
}

func (m *commonMap[K, V]) ensureInit() {
	m.initialized.Do(m.initialize)
}

func (m *commonMap[K, V]) initialize() {
	m.cleanupEnabled.Store(DefaultMapCleanupEnabled.Load())
}

// DefaultMapCleanupEnabled configures the default value of IsCleanupEnabled() for new maps.
// Default: true.
var DefaultMapCleanupEnabled atomic.Bool

func init() {
	DefaultMapCleanupEnabled.Store(true)
}

// IsCleanupEnabled indicates whether the cleanup with [runtime.Cleanup] is enabled.
// The default value is controlled by [DefaultMapCleanupEnabled].
func (m *commonMap[K, V]) IsCleanupEnabled() bool {
	m.ensureInit()
	return m.cleanupEnabled.Load()
}

// SetCleanupEnabled configures whether the cleanup with [runtime.Cleanup] is enabled.
func (m *commonMap[K, V]) SetCleanupEnabled(enabled bool) {
	m.ensureInit()
	m.cleanupEnabled.Store(enabled)
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
