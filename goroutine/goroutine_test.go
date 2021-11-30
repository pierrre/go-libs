package goroutine

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestGo(t *testing.T) {
	var called int64
	wait := Go(func() {
		atomic.AddInt64(&called, 1)
	})
	wait()
	if called == 0 {
		t.Fatal("not called")
	}
}

func TestWaitGroup(t *testing.T) {
	wg := new(sync.WaitGroup)
	var called int64
	WaitGroup(wg, func() {
		atomic.AddInt64(&called, 1)
	})
	wg.Wait()
	if called == 0 {
		t.Fatal("not called")
	}
}

func TestRunN(t *testing.T) {
	var called int64
	RunN(10, func() {
		atomic.AddInt64(&called, 1)
	})
	if called != 10 {
		t.Fatalf("unexpected called: got %d, want %d", called, 10)
	}
}
