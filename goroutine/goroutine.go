// Package goroutine helps to manage goroutines safely.
//
// It recovers panic with [panichandle].
package goroutine

import (
	"sync"

	"github.com/pierrre/go-libs/panichandle"
)

// Go executes a function in a new goroutine.
func Go(f func()) {
	go func() {
		defer panichandle.Recover()
		f()
	}()
}

// GoWait executes a function in a new goroutine.
//
// It returns a function that blocks until the goroutine is terminated.
// The caller must call this function.
func GoWait(f func()) (wait func()) {
	ch := make(chan struct{})
	Go(func() {
		defer close(ch)
		f()
	})
	return func() {
		<-ch
	}
}

// WaitGroup executes a function in a new goroutine with a [sync.WaitGroup].
// It calls [sync.WaitGroup.Add] before starting it, and [sync.WaitGroup.Done] when the goroutine is terminated.
func WaitGroup(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	Go(func() {
		defer wg.Done()
		f()
	})
}

// N executes a function with multiple goroutines.
// It blocks until all goroutines are terminated.
func N(n int, f func(i int)) {
	wg := new(sync.WaitGroup)
	for i := 0; i < n; i++ {
		i := i
		WaitGroup(wg, func() {
			f(i)
		})
	}
	wg.Wait()
}

// Slice executes a function with a different goroutine for each element of the slice.
// It blocks until all goroutines are terminated.
func Slice[E any](s []E, f func(i int, e E)) {
	wg := new(sync.WaitGroup)
	for i, e := range s {
		i, e := i, e
		WaitGroup(wg, func() {
			f(i, e)
		})
	}
	wg.Wait()
}

// Map executes a function with a different goroutine for each element of the map.
// It blocks until all goroutines are terminated.
func Map[M ~map[K]V, K comparable, V any](m M, f func(k K, v V)) {
	wg := new(sync.WaitGroup)
	for k, v := range m {
		k, v := k, v
		WaitGroup(wg, func() {
			f(k, v)
		})
	}
	wg.Wait()
}
