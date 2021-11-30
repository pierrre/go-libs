// Package goroutine helps to manage goroutines safely.
//
// It recovers panic with panichandle.
package goroutine

import (
	"sync"

	"github.com/pierrre/go-libs/panichandle"
)

// Go executes a function in a new goroutine.
// It returns a function that blocks until the goroutine is terminated.
// The caller must call this function.
func Go(f func()) (wait func()) {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		defer panichandle.Recover()
		f()
	}()
	return func() {
		<-ch
	}
}

// WaitGroup executes a function in a new goroutine with a WaitGroup.
// It calls WaitGroup.Add() before starting it, and WaitGroup.Done() when the goroutine is terminated.
func WaitGroup(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer panichandle.Recover()
		f()
	}()
}

// RunN runs a function with multiple goroutines.
// It blocks until all goroutines are terminated.
func RunN(n int, f func()) {
	wg := new(sync.WaitGroup)
	for i := 0; i < n; i++ {
		WaitGroup(wg, f)
	}
	wg.Wait()
}
