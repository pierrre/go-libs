// Package bufpool provides a sync.Pool of bytes.Buffer.
package bufpool

import (
	"bytes"
	"sync"
)

const maxCapDefault = 1 << 16 // 64 KiB

// Pool is a pool of *bytes.Buffer.
//
// Nuffers are automatically reset.
type Pool struct {
	pool sync.Pool

	// MaxCap defines the maximum capacity accepted for recycled buffer.
	// If Put() is called with a buffer larger than this value, it's discarded.
	// See https://github.com/golang/go/issues/23199 .
	// 0 (default) means 64 KiB.
	// A negative value means no limit.
	MaxCap int
}

// Get returns a buffer from the Pool.
func (p *Pool) Get() *bytes.Buffer {
	bufItf := p.pool.Get()
	if bufItf != nil {
		return bufItf.(*bytes.Buffer) //nolint:forcetypeassert // The pool only contains *bytes.Buffer.
	}
	return new(bytes.Buffer)
}

// Put puts the buffer to the Pool.
// WARNING: the call MUST NOT reuse the buffer's content after this call.
func (p *Pool) Put(buf *bytes.Buffer) {
	maxCap := p.MaxCap
	if maxCap == 0 {
		maxCap = maxCapDefault
	}
	if maxCap < 0 || buf.Cap() <= maxCap {
		buf.Reset()
		p.pool.Put(buf)
	}
}
