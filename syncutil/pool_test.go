package syncutil_test

import (
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/syncutil"
)

func TestPool(t *testing.T) {
	var p Pool[*[]byte]
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

func TestValuePool(t *testing.T) {
	p := &ValuePool[[]byte]{
		New: func() []byte {
			return make([]byte, 10)
		},
	}
	b := p.Get()
	assert.SliceLen(t, b, 10)
	p.Put(b)
}

func TestValuePoolNoValue(t *testing.T) {
	p := &ValuePool[[]byte]{}
	b := p.Get()
	assert.SliceNil(t, b)
}

func TestValuePoolAllocs(t *testing.T) {
	p := &ValuePool[[]byte]{
		New: func() []byte {
			return make([]byte, 10)
		},
	}
	assert.AllocsPerRun(t, 100, func() {
		b := p.Get()
		p.Put(b)
	}, 0)
}

func BenchmarkValuePool(b *testing.B) {
	p := &ValuePool[[]byte]{
		New: func() []byte {
			return make([]byte, 10)
		},
	}
	for b.Loop() {
		b := p.Get()
		p.Put(b)
	}
}
