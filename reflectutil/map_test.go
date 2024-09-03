package reflectutil_test

import (
	"reflect"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/reflectutil"
)

var (
	testMap = func() map[int]int {
		m := make(map[int]int)
		for i := range 10 {
			m[i] = i
		}
		return m
	}()
	testMapValue           = reflect.ValueOf(testMap)
	testMapUnexportedValue = reflect.ValueOf(struct{ v map[int]int }{v: testMap}).FieldByName("v")
)

func TestGetMapEntries(t *testing.T) {
	es := GetMapEntries(testMapValue)
	assert.SliceLen(t, es, 10)
	for _, e := range es {
		assert.Equal(t, e.Key.Int(), e.Value.Int())
	}
	es.Release()
}

func TestMapEntriesEmpty(t *testing.T) {
	es := GetMapEntries(reflect.ValueOf(map[int]int(nil)))
	assert.SliceNil(t, es)
	es.Release()
}

func TestMapEntriesAllocs(t *testing.T) {
	assert.AllocsPerRun(t, 100, func() {
		es := GetMapEntries(testMapValue)
		es.Release()
	}, 0)
}

func BenchmarkGetMapEntries(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		es := GetMapEntries(testMapValue)
		es.Release()
	}
}

func TestGetMapEntriesUnexported(t *testing.T) {
	es := GetMapEntries(testMapUnexportedValue)
	assert.SliceLen(t, es, 10)
	for _, e := range es {
		assert.Equal(t, e.Key.Int(), e.Value.Int())
	}
	es.Release()
}

func TestGetMapEntriesUnexportedAllocs(t *testing.T) {
	assert.AllocsPerRun(t, 100, func() {
		es := GetMapEntries(testMapUnexportedValue)
		es.Release()
	}, 21)
}

func BenchmarkGetMapEntriesUnexported(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		es := GetMapEntries(testMapUnexportedValue)
		es.Release()
	}
}
