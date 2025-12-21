// Package bufpool provides a [sync.Pool] of [bytes.Buffer].
package bufpool

import (
	"bytes"

	"github.com/pierrre/go-libs/syncutil"
)

const maxCapDefault = 1 << 16 // 64 KiB

// Pool is a [sync.Pool] of [bytes.Buffer].
//
// Buffers are automatically reset.
type Pool struct {
	pool syncutil.Pool[*bytes.Buffer]

	// MaxCap defines the maximum capacity accepted for recycled buffer.
	// If Put() is called with a buffer larger than this value, it's discarded.
	// See https://github.com/golang/go/issues/23199 .
	// 0 (default) means 64 KiB.
	// A negative value means no limit.
	MaxCap int

	// Clear indicates whether to clear the buffer before putting it back to the pool.
	// It prevents leaking sensitive data, but has a small performance cost.
	Clear bool
}

// Get returns a [bytes.Buffer] from the Pool.
func (p *Pool) Get() *bytes.Buffer {
	buf := p.pool.Get()
	if buf == nil {
		return new(bytes.Buffer)
	}
	return buf
}

// Put puts the [bytes.Buffer] to the Pool.
// WARNING: the caller MUST NOT reuse the buffer's content after this call.
func (p *Pool) Put(buf *bytes.Buffer) {
	maxCap := p.MaxCap
	if maxCap == 0 {
		maxCap = maxCapDefault
	}
	if maxCap < 0 || buf.Cap() <= maxCap {
		buf.Reset()
		if p.Clear {
			clear(buf.AvailableBuffer())
		}
		p.pool.Put(buf)
	}
}
