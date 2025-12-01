package syncutil_test

import (
	"sync"
	"testing"

	. "github.com/pierrre/go-libs/syncutil"
)

func TestMap(t *testing.T) {
	var m Map[string, int]
	m.Clear()
	m.CompareAndDelete("key", 1)
	m.CompareAndSwap("key", 1, 2)
	m.Delete("key")
	m.Load("key")
	m.LoadAndDelete("key")
	m.LoadOrStore("key", 1)
	m.Range(func(key string, value int) bool {
		return true
	})
	m.Store("key", 1)
	m.Swap("key", 1)
}

func BenchmarkMapStore(b *testing.B) {
	var m Map[string, int]
	for b.Loop() {
		m.Store("key", 1)
	}
}

func BenchmarkSyncMapStore(b *testing.B) {
	var m sync.Map
	for b.Loop() {
		m.Store("key", 1)
	}
}

func BenchmarkMapLoad(b *testing.B) {
	var m Map[string, int]
	m.Store("key", 1)
	for b.Loop() {
		m.Load("key")
	}
}

func BenchmarkSyncMapLoad(b *testing.B) {
	var m sync.Map
	m.Store("key", 1)
	for b.Loop() {
		m.Load("key")
	}
}
