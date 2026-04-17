package bytesutil

import (
	"github.com/pierrre/go-libs/syncutil"
)

const writerPoolMaxCapDefault = 1 << 16 // 64 KiB

// WriterPool is a [syncutil.Pool] of [Writer].
//
// Writers are automatically reset.
type WriterPool struct {
	pool syncutil.Pool[*Writer]

	// MaxCap defines the maximum capacity accepted for recycled writers.
	// If Put() is called with a writer larger than this value, it's discarded.
	// See https://github.com/golang/go/issues/23199.
	// 0 (default) means 64 KiB.
	// A negative value means no limit.
	MaxCap int

	// Clear indicates whether to clear the writer before putting it back to the pool.
	// It prevents leaking sensitive data, but has a small performance cost.
	Clear bool
}

// Get returns a [Writer] from the pool.
func (p *WriterPool) Get() *Writer {
	w := p.pool.Get()
	if w == nil {
		return new(Writer)
	}
	return w
}

// Put puts the [Writer] to the Pool.
// WARNING: the caller MUST NOT reuse the writer's content after this call.
func (p *WriterPool) Put(w *Writer) {
	maxCap := p.MaxCap
	if maxCap == 0 {
		maxCap = writerPoolMaxCapDefault
	}
	if maxCap < 0 || w.Cap() <= maxCap {
		if p.Clear {
			w.Clear()
		} else {
			w.Reset()
		}
		p.pool.Put(w)
	}
}
