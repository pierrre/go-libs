package syncutil_test

import (
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/syncutil"
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

func TestPoolFor(t *testing.T) {
	var p PoolFor[[]byte]
	bp := p.Get()
	assert.Zero(t, bp)
	p.Put(new([]byte))
	bp = p.Get()
	assert.NotZero(t, bp)
	p.New = func() *[]byte {
		return new([]byte)
	}
	bp = p.Get()
	assert.NotZero(t, bp)
}
