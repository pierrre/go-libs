package bytesutil

import (
	"cmp"
	"math/bits"
	"slices"
	"sync/atomic"
	"unsafe" //nolint:depguard // Required for unsafe.Sizeof.

	"github.com/pierrre/go-libs/syncutil"
)

const (
	writerPoolMinBitSize              = 6
	writerPoolSteps                   = 20
	writerPoolMinSize                 = 1 << writerPoolMinBitSize
	writerPoolCalibrateCallsThreshold = 42000
	writerPoolMaxPercentile           = 0.95
)

// WriterPool is a [syncutil.Pool] of [Writer].
//
// Writers are automatically reset.
//
// The pool calibrates itself with the sizes of the writers put back into it.
// The capacity of the new writers converges to the most frequently used size, and writers with a capacity above the calibrated maximum size are discarded.
// The maximum size is computed so that at least 95% of the calls to [WriterPool.Put] are accepted, which limits the memory retained by the pool (see https://github.com/golang/go/issues/23199).
// There is no limit until the first calibration.
//
// Setting [WriterPool.DefaultSize] to a positive value and [WriterPool.MaxSize] to a non-zero value fixes both sizes and disables the calibration.
// When [WriterPool.DefaultSize] is calibrated automatically and [WriterPool.MaxSize] is positive, the calibrated default size is clamped to [WriterPool.MaxSize].
// When [WriterPool.DefaultSize] is positive and [WriterPool.MaxSize] is calibrated automatically, the calibrated maximum size is raised to at least [WriterPool.DefaultSize].
// The exported fields must not be modified after the first use.
//
// The calibration algorithm is adapted from https://github.com/valyala/bytebufferpool.
type WriterPool struct {
	pool syncutil.Pool[*Writer]

	// DefaultSize is the capacity preallocated for the new writers returned by [WriterPool.Get].
	// 0 (default) means that it's calibrated automatically, a positive value fixes it, and a negative value behaves as 0.
	// It must not be larger than a positive [WriterPool.MaxSize], otherwise all the new writers are discarded.
	DefaultSize int

	// MaxSize is the maximum capacity accepted for recycled writers, larger writers are discarded.
	// 0 (default) means that it's calibrated automatically, a positive value fixes it, and a negative value means no limit.
	MaxSize int

	// Clear indicates whether to clear the [Writer] before putting it back to the pool.
	// It prevents leaking sensitive data, but has a small performance cost.
	Clear bool

	calls                 [writerPoolSteps]writerPoolCounter
	calibrating           atomic.Bool
	calibratedDefaultSize atomic.Uint64
	calibratedMaxSize     atomic.Uint64
}

type writerPoolCounter struct {
	n atomic.Uint64

	// The counter is padded to 128 bytes, a multiple of the CPU cache line size on widespread platforms, to limit false sharing between the counters.
	_ [128 - unsafe.Sizeof(atomic.Uint64{})]byte
}

// Get returns a [Writer] from the pool.
func (p *WriterPool) Get() *Writer {
	w := p.pool.Get()
	if w == nil {
		w = new(Writer)
		if size := p.defaultSize(); size > 0 {
			*w = make(Writer, 0, size)
		}
	}
	return w
}

func (p *WriterPool) defaultSize() int {
	if p.DefaultSize > 0 {
		return p.DefaultSize
	}
	return int(p.calibratedDefaultSize.Load())
}

// Put puts the [Writer] back to the pool.
// WARNING: the caller MUST NOT reuse the writer's content after this call.
func (p *WriterPool) Put(w *Writer) {
	p.countCall(w)
	maxSize := p.maxSize()
	if maxSize < 0 || w.Cap() <= maxSize {
		if p.Clear {
			w.Clear()
		} else {
			w.Reset()
		}
		p.pool.Put(w)
	}
}

func (p *WriterPool) maxSize() int {
	if p.MaxSize != 0 {
		return p.MaxSize
	}
	maxSize := int(p.calibratedMaxSize.Load())
	if maxSize == 0 {
		return -1
	}
	return max(maxSize, p.DefaultSize)
}

func (p *WriterPool) countCall(w *Writer) {
	if p.DefaultSize > 0 && p.MaxSize != 0 {
		return
	}
	n := w.Len()
	if n == 0 {
		n = w.Cap()
	}
	if p.calls[writerPoolIndex(n)].n.Add(1) > writerPoolCalibrateCallsThreshold {
		p.tryCalibrate()
	}
}

func writerPoolIndex(n int) int {
	if n <= writerPoolMinSize {
		return 0
	}
	return min(bits.Len(uint(n-1))-writerPoolMinBitSize, writerPoolSteps-1)
}

func (p *WriterPool) tryCalibrate() {
	if !p.calibrating.CompareAndSwap(false, true) {
		return
	}
	defer p.calibrating.Store(false)
	p.calibrate()
}

func (p *WriterPool) calibrate() {
	sizes := make([]writerPoolCallSize, writerPoolSteps)
	var callsSum uint64
	for i := range writerPoolSteps {
		calls := p.calls[i].n.Swap(0)
		callsSum += calls
		sizes[i] = writerPoolCallSize{
			calls: calls,
			size:  writerPoolMinSize << i,
		}
	}
	if callsSum == 0 {
		return
	}
	slices.SortFunc(sizes, func(a, b writerPoolCallSize) int {
		return -cmp.Compare(a.calls, b.calls)
	})
	defaultSize := sizes[0].size
	maxSize := defaultSize
	maxSum := uint64(float64(callsSum) * writerPoolMaxPercentile)
	callsSum = 0
	for _, s := range sizes {
		if callsSum > maxSum {
			break
		}
		callsSum += s.calls
		if s.size > maxSize {
			maxSize = s.size
		}
	}
	if p.MaxSize > 0 {
		defaultSize = min(defaultSize, p.MaxSize)
	}
	p.calibratedDefaultSize.Store(uint64(defaultSize))
	p.calibratedMaxSize.Store(uint64(maxSize))
}

type writerPoolCallSize struct {
	calls uint64
	size  int
}
