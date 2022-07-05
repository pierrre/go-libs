package syncutil

import (
	"testing"
)

func TestPool(t *testing.T) {
	p := &Pool[string]{
		New: func() string {
			return "test"
		},
	}
	s, ok := p.Get()
	if !ok {
		t.Fatal("not ok")
	}
	if s != "test" {
		t.Fatalf("unexpected result: got %q, want %q", s, "test")
	}
	p.Put(s)
}

func TestPoolWithoutNew(t *testing.T) {
	p := &Pool[string]{}
	_, ok := p.Get()
	if ok {
		t.Fatal("ok")
	}
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
			b.Fatal("not ok")
		}
		p.Put(sp)
	}
}
