package goroutine

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/go-libs/internal/golibstest"
)

func init() {
	golibstest.Configure()
}

func TestGo(t *testing.T) {
	var called int64
	wait := Go(func() {
		atomic.AddInt64(&called, 1)
	})
	wait()
	assert.Equal(t, called, 1)
}

func TestWaitGroup(t *testing.T) {
	wg := new(sync.WaitGroup)
	var called int64
	WaitGroup(wg, func() {
		atomic.AddInt64(&called, 1)
	})
	wg.Wait()
	assert.Equal(t, called, 1)
}

func TestRunN(t *testing.T) {
	var called int64
	RunN(10, func() {
		atomic.AddInt64(&called, 1)
	})
	assert.Equal(t, called, 10)
}
