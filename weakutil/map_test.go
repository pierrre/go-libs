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
