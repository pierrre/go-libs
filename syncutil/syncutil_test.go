package syncutil

import (
	"testing"
)

func TestMapFor(t *testing.T) {
	var m MapFor[string, int]
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