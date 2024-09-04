package reflectutil_test

import (
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

func TestGetMapKeys(t *testing.T) {
	ks := GetMapKeys(testMapValue)
	assert.SliceLen(t, ks, 10)
	for _, k := range ks {
		assert.True(t, k.CanInt())
	}
	ks.Release()
}

func TestGetMapKeysEmpty(t *testing.T) {
	ks := GetMapKeys(reflect.ValueOf(map[int]int(nil)))
	assert.SliceNil(t, ks)
	ks.Release()
}

func TestGetMapKeysAllocs(t *testing.T) {
	assert.AllocsPerRun(t, 100, func() {
		ks := GetMapKeys(testMapValue)
		ks.Release()
	}, 0)
}

func BenchmarkGetMapKeys(b *testing.B) {
	for range b.N {
		ks := GetMapKeys(testMapValue)
		ks.Release()
	}
}

func TestGetMapKeysUnexported(t *testing.T) {
	ks := GetMapKeys(testMapUnexportedValue)
	assert.SliceLen(t, ks, 10)
	for _, k := range ks {
		assert.True(t, k.CanInt())
	}
	ks.Release()
}

func TestGetMapKeysUnexportedAllocs(t *testing.T) {
	assert.AllocsPerRun(t, 100, func() {
		ks := GetMapKeys(testMapUnexportedValue)
		ks.Release()
	}, 11)
}

func BenchmarkGetMapKeysUnexported(b *testing.B) {
	for range b.N {
		ks := GetMapKeys(testMapUnexportedValue)
		ks.Release()
	}
}
