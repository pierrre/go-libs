package weakutil

import (
	"sync"
	"sync/atomic"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

type commonMap[K comparable, V any] struct {
	m               syncutil.Map[K, V]
	initialized     sync.Once
	cleanupEnabled  atomic.Bool
	sweepWriteCount atomic.Uint64
	writeCount      atomic.Uint64
	sweeping        atomic.Bool
}

func (m *commonMap[K, V]) ensureInit() {
	m.initialized.Do(m.initialize)
}

func (m *commonMap[K, V]) initialize() {
	m.cleanupEnabled.Store(DefaultMapCleanupEnabled.Load())
	m.sweepWriteCount.Store(DefaultMapSweepWriteCount.Load())
}

// DefaultMapCleanupEnabled configures the default value of IsCleanupEnabled() for new maps.
// Default: true.
var DefaultMapCleanupEnabled atomic.Bool

// DefaultMapSweepWriteCount configures the default value of GetSweepWriteCount() for new maps.
// Default: 10000.
var DefaultMapSweepWriteCount atomic.Uint64

func init() {
	DefaultMapCleanupEnabled.Store(true)
	DefaultMapSweepWriteCount.Store(10000)
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

// GetSweepWriteCount returns the number of writes between two sweeps.
// The sweep only runs when cleanup is disabled with [SetCleanupEnabled].
// A value of 0 disables the sweep.
func (m *commonMap[K, V]) GetSweepWriteCount() uint64 {
	m.ensureInit()
	return m.sweepWriteCount.Load()
}

// SetSweepWriteCount configures the number of writes between two sweeps.
// The sweep only runs when cleanup is disabled with [SetCleanupEnabled].
// A value of 0 disables the sweep.
func (m *commonMap[K, V]) SetSweepWriteCount(count uint64) {
	m.ensureInit()
	m.sweepWriteCount.Store(count)
}

func (m *commonMap[K, V]) maybeSweep(sweep func()) {
	if m.IsCleanupEnabled() {
		return
	}
	threshold := m.GetSweepWriteCount()
	if threshold == 0 {
		return
	}
	if m.writeCount.Add(1) < threshold {
		return
	}
	if !m.sweeping.CompareAndSwap(false, true) {
		return // another goroutine is already sweeping
	}
	m.writeCount.Store(0)
	defer m.sweeping.Store(false)
	sweep()
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
