// Package bufpool provides a sync.Pool of bytes.Buffer.
package bufpool

import (
	"bytes"
	"sync"

	"github.com/pierrre/go-libs/syncutil"
)

const maxCapDefault = 1 << 16 // 64 KiB

// Pool is a pool of *bytes.Buffer.
type Pool struct {
	once sync.Once
	pool syncutil.Pool[*bytes.Buffer]

	// MaxCap defines the maximum capacity accepted for recycled buffer.
	// If Put() is called with a buffer larger than this value, it's discarded.
	// See https://github.com/golang/go/issues/23199 .
	// 0 (default) means 64 KiB.
	// A negative value means no limit.
	MaxCap int
}

func (p *Pool) ensureInit() {
	p.once.Do(p.init)
}

func (p *Pool) init() {
	p.pool.New = func() *bytes.Buffer {
		return new(bytes.Buffer)
	}
	if p.MaxCap == 0 {
		p.MaxCap = maxCapDefault
	}
}

// Get gets a buffer from the Pool, resets it and returns it.
func (p *Pool) Get() *bytes.Buffer {
	p.ensureInit()
	buf, _ := p.pool.Get()
	buf.Reset()
	return buf
}

// Put puts the buffer to the Pool.
// WARNING: the call MUST NOT reuse the buffer's content after this call.
func (p *Pool) Put(buf *bytes.Buffer) {
	p.ensureInit()
	if p.MaxCap < 0 || buf.Cap() <= p.MaxCap {
		p.pool.Put(buf)
	}
}
