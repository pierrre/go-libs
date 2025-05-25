package weakutil_test

import (
	"runtime"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/weakutil"
)

func TestSimpleMapLoad(t *testing.T) {
	m := new(SimpleMap[string, [64]byte])
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	p2, ok := m.Load("test")
	assert.True(t, ok)
	assert.Equal(t, p2, p1)
}

func TestSimpleMapLoadGCDelete(t *testing.T) {
	m := new(SimpleMap[string, [64]byte])
	v := [64]byte{}
	m.Store("test", &v)
	runtime.GC()
	p, ok := m.Load("test")
	assert.False(t, ok)
	assert.Zero(t, p)
}

func BenchmarkSimpleMapStoreSame(b *testing.B) {
	m := new(SimpleMap[string, [64]byte])
	m.AutoCleanInterval = -1
	v := [64]byte{}
	p := &v
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("test", p)
		}
	})
}

func BenchmarkSimpleMapStoreNil(b *testing.B) {
	m := new(SimpleMap[string, [64]byte])
	m.AutoCleanInterval = -1
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store("test", nil)
		}
	})
}

func BenchmarkSimpleMapStoreDifferent(b *testing.B) {
	m := new(SimpleMap[string, [64]byte])
	m.AutoCleanInterval = -1
	v := [64]byte{}
	p := &v
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			var pp *[64]byte
			if i%2 == 0 {
				pp = p
			}
			m.Store("test", pp)
		}
	})
}

func BenchmarkSimpleMapLoad(b *testing.B) {
	m := new(SimpleMap[string, [64]byte])
	m.AutoCleanInterval = -1
	v := [64]byte{}
	p1 := &v
	m.Store("test", p1)
	var res *[64]byte
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p2, ok := m.Load("test")
			assert.True(b, ok)
			assert.Equal(b, p2, p1)
			res = p2
		}
	})
	benchRes = res
}
