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
	p := &Pool[string]{
		New: func() string {
			return "test"
		},
	}
	s, ok := p.Get()
	assert.True(t, ok)
	assert.Equal(t, s, "test")
	p.Put(s)
}

func TestPoolWithoutNew(t *testing.T) {
	p := &Pool[string]{}
	_, ok := p.Get()
	assert.False(t, ok)
}

func BenchmarkPool(b *testing.B) {
	p := &Pool[*string]{
		New: func() *string {
			s := "test"
			return &s
		},
	}
	for i := 0; i < b.N; i++ {
		sp, ok := p.Get()
		if !ok {
			assert.True(b, ok)
		}
		p.Put(sp)
	}
}
