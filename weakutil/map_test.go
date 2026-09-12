package weakutil

import (
	"testing"

	"github.com/pierrre/assert"
)

func TestCommonMapCleanupEnabled(t *testing.T) {
	m := new(commonMap[int, int])
	assert.True(t, m.IsCleanupEnabled())
	m.SetCleanupEnabled(false)
	assert.False(t, m.IsCleanupEnabled())
}

func TestCommonMapSweepWriteCount(t *testing.T) {
	m := new(commonMap[int, int])
	assert.Equal(t, m.GetSweepWriteCount(), uint64(10000))
	m.SetSweepWriteCount(123)
	assert.Equal(t, m.GetSweepWriteCount(), uint64(123))
	m.SetSweepWriteCount(0)
	assert.Equal(t, m.GetSweepWriteCount(), uint64(0))
}

func TestCommonMapSweepWriteCountDefault(t *testing.T) {
	old := DefaultMapSweepWriteCount.Load()
	DefaultMapSweepWriteCount.Store(123)
	t.Cleanup(func() { DefaultMapSweepWriteCount.Store(old) })
	m := new(commonMap[int, int])
	assert.Equal(t, m.GetSweepWriteCount(), uint64(123))
}

func TestCommonMapSweepSkipsWhenSweeping(t *testing.T) {
	m := new(commonMap[int, int])
	m.SetCleanupEnabled(false)
	m.SetSweepWriteCount(1)
	sweepStarted := make(chan struct{})
	releaseSweep := make(chan struct{})
	sweepA := func() {
		close(sweepStarted)
		<-releaseSweep
	}
	sweepBCalled := false
	sweepB := func() {
		sweepBCalled = true
	}
	done := make(chan struct{})
	go func() {
		m.maybeSweep(sweepA)
		close(done)
	}()
	<-sweepStarted // sweep A is in progress
	m.maybeSweep(sweepB)
	close(releaseSweep)
	<-done
	assert.False(t, sweepBCalled)
	assert.False(t, m.sweeping.Load())
}

func TestCommonMapSweepDisabled(t *testing.T) {
	m := new(commonMap[int, int])
	m.SetCleanupEnabled(false)
	m.SetSweepWriteCount(0)
	m.maybeSweep(func() { t.Fatal("sweep must not run when disabled") })
	assert.Equal(t, m.writeCount.Load(), uint64(0))
}

func TestCommonMapSweepSkippedWhenCleanupEnabled(t *testing.T) {
	m := new(commonMap[int, int])
	m.SetSweepWriteCount(1)
	m.maybeSweep(func() { t.Fatal("sweep must not run when cleanup is enabled") })
	assert.Equal(t, m.writeCount.Load(), uint64(0))
}

func TestCommonMapSweepNotTriggeredBeforeThreshold(t *testing.T) {
	m := new(commonMap[int, int])
	m.SetCleanupEnabled(false)
	m.SetSweepWriteCount(2)
	sweepCalled := false
	m.maybeSweep(func() { sweepCalled = true })
	assert.False(t, sweepCalled)
	assert.Equal(t, m.writeCount.Load(), uint64(1))
	m.maybeSweep(func() { sweepCalled = true })
	assert.True(t, sweepCalled)
	assert.Equal(t, m.writeCount.Load(), uint64(0))
}

func getMapLen[M interface{ Range(f func(K, V) bool) }, K any, V any](m M) int {
	count := 0
	for range m.Range {
		count++
	}
	return count
}

func assertDeadEntry[K any](tb testing.TB, raw func(K) (present, alive bool), id K) {
	tb.Helper()
	present, alive := raw(id)
	assert.True(tb, present)
	assert.False(tb, alive)
}

func assertNoEntry[K any](tb testing.TB, raw func(K) (present, alive bool), id K) {
	tb.Helper()
	present, _ := raw(id)
	assert.False(tb, present)
}
