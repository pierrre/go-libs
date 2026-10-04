package weakutil

import (
	"runtime"
	"sync"
	"sync/atomic"
	"weak"

	"github.com/pierrre/go-libs/syncutil"
)

type mapEntry interface {
	stopCleanup()
}

type commonMap[K comparable, E mapEntry] struct {
	// m is a separate allocation, not an embedded value: the cleanups attached to stored keys and values hold a reference to it, and embedding it would pin the outer map for as long as a stored key or value is alive.
	// When the outer map is collected, commonMapCleanup stops the cleanups of the stored entries, so that the keys and values no longer hold a reference to m, which can then be collected.
	m               *syncutil.Map[K, E]
	initialized     sync.Once
	cleanupEnabled  atomic.Bool
	sweepWriteCount atomic.Uint64
	writeCount      atomic.Uint64
	sweeping        atomic.Bool
}

func (m *commonMap[K, E]) ensureInit() {
	m.initialized.Do(m.initialize)
}

func (m *commonMap[K, E]) initialize() {
	m.m = &syncutil.Map[K, E]{}
	m.cleanupEnabled.Store(DefaultMapCleanupEnabled.Load())
	m.sweepWriteCount.Store(DefaultMapSweepWriteCount.Load())
	runtime.AddCleanup(m, commonMapCleanup, commonMapCleanupArg[K, E]{
		m: m.m,
	})
}

type commonMapCleanupArg[K comparable, E any] struct {
	m *syncutil.Map[K, E]
}

func commonMapCleanup[K comparable, E mapEntry](arg commonMapCleanupArg[K, E]) {
	arg.m.Range(func(_ K, e E) bool {
		e.stopCleanup()
		return true
	})
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
func (m *commonMap[K, E]) IsCleanupEnabled() bool {
	m.ensureInit()
	return m.cleanupEnabled.Load()
}

// SetCleanupEnabled configures whether the cleanup with [runtime.Cleanup] is enabled.
func (m *commonMap[K, E]) SetCleanupEnabled(enabled bool) {
	m.ensureInit()
	m.cleanupEnabled.Store(enabled)
}

// GetSweepWriteCount returns the number of writes between two sweeps.
// The sweep only runs when cleanup is disabled with [SetCleanupEnabled].
// A value of 0 disables the sweep.
func (m *commonMap[K, E]) GetSweepWriteCount() uint64 {
	m.ensureInit()
	return m.sweepWriteCount.Load()
}

// SetSweepWriteCount configures the number of writes between two sweeps.
// The sweep only runs when cleanup is disabled with [SetCleanupEnabled].
// A value of 0 disables the sweep.
func (m *commonMap[K, E]) SetSweepWriteCount(count uint64) {
	m.ensureInit()
	m.sweepWriteCount.Store(count)
}

func (m *commonMap[K, E]) maybeSweep(sweep func()) {
	m.ensureInit()
	if m.cleanupEnabled.Load() {
		return
	}
	threshold := m.sweepWriteCount.Load()
	if threshold == 0 {
		return
	}
	if m.writeCount.Add(1) < threshold {
		return
	}
	if !m.sweeping.CompareAndSwap(false, true) {
		return
	}
	defer func() {
		m.writeCount.Store(0)
		m.sweeping.Store(false)
	}()
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

func clearMap[K comparable, E any](m *syncutil.Map[K, E], del func(K, E) bool) {
	for range 10 {
		var count int64
		m.Range(func(k K, e E) bool {
			count++
			del(k, e)
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
