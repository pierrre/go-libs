package syncutil

import (
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/go-libs/internal/golibstest"
)

func init() {
	golibstest.Configure()
}

func TestPool(t *testing.T) {
	p := &Pool[[]byte]{
		New: func() *[]byte {
			b := make([]byte, 1)
			return &b
		},
	}
	for i := 0; i < 10; i++ {
		bp := p.Get()
		assert.NotZero(t, bp)
		assert.SliceLen(t, *bp, 1)
		p.Put(bp)
	}
}

func TestPoolWithoutNew(t *testing.T) {
	p := &Pool[[]byte]{}
	bp := p.Get()
	assert.Zero(t, bp)
}

func BenchmarkPool(b *testing.B) {
	p := &Pool[[]byte]{
		New: func() *[]byte {
			b := make([]byte, 1)
			return &b
		},
	}
	for i := 0; i < b.N; i++ {
		bp := p.Get()
		if bp == nil {
			assert.NotZero(b, bp)
		}
		p.Put(bp)
	}
}
