package reflectutil_test

import (
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

func TestGetSortedMap(t *testing.T) {
	es := GetSortedMap(testMapValue)
	for i, e := range es {
		assert.Equal(t, int(e.Key.Int()), i)
		assert.Equal(t, int(e.Value.Int()), i)
	}
	es.Release()
}

func TestGetSortedMapAllocs(t *testing.T) {
	assert.AllocsPerRun(t, 100, func() {
		es := GetSortedMap(testMapValue)
		es.Release()
	}, 0)
}
