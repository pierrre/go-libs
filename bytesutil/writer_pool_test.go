package bytesutil_test

import (
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/bytesutil"
)

const testWriterPoolData = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum." //nolint:lll // This is a long text for benchmark.

func TestWriterPool(t *testing.T) {
	p := &WriterPool{}
	for range 10 {
		w := p.Get()
		assert.Equal(t, w.Len(), 0)
		w.AppendString(testWriterPoolData)
		p.Put(w)
	}
}

func BenchmarkWriterPool(b *testing.B) {
	p := &WriterPool{}
	for b.Loop() {
		w := p.Get()
		w.AppendString(testWriterPoolData)
		p.Put(w)
	}
}

func TestWriterPoolClear(t *testing.T) {
	p := &WriterPool{
		Clear: true,
	}
	for range 10 {
		w := p.Get()
		assert.Equal(t, w.Len(), 0)
		avail := w.AvailableBuffer()
		for _, b := range avail[:cap(avail)] {
			assert.Equal(t, b, 0)
		}
		w.AppendString(testWriterPoolData)
		p.Put(w)
	}
}

func BenchmarkWriterPoolClear(b *testing.B) {
	p := &WriterPool{
		Clear: true,
	}
	for b.Loop() {
		w := p.Get()
		w.AppendString(testWriterPoolData)
		p.Put(w)
	}
}
