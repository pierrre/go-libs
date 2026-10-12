package bytesutil_test

import (
	"math"
	"sync"
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

func TestWriterPoolIndex(t *testing.T) {
	for _, tc := range []struct {
		n    int
		want int
	}{
		{n: 0, want: 0},
		{n: 1, want: 0},
		{n: 64, want: 0},
		{n: 65, want: 1},
		{n: 128, want: 1},
		{n: 129, want: 2},
		{n: 4096, want: 6},
		{n: 4097, want: 7},
		{n: 1 << 24, want: 18},
		{n: 1 << 25, want: 19},
		{n: (1 << 25) + 1, want: 19},
		{n: math.MaxInt, want: 19},
	} {
		assert.Equal(t, WriterPoolIndexForTest(tc.n), tc.want)
	}
}

func TestWriterPoolCalibrate(t *testing.T) {
	p := &WriterPool{}
	p.SetCallsForTest(1024, 100)
	p.SetCallsForTest(65536, 10)
	p.CalibrateForTest()
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(1024))
	assert.Equal(t, p.CalibratedMaxSizeForTest(), uint64(65536))
	assert.True(t, p.NoCallsForTest())
}

func TestWriterPoolCalibrateSingleSize(t *testing.T) {
	p := &WriterPool{}
	p.SetCallsForTest(4096, 5000)
	p.CalibrateForTest()
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(4096))
	assert.Equal(t, p.CalibratedMaxSizeForTest(), uint64(4096))
}

func TestWriterPoolCalibrateZeroCalls(t *testing.T) {
	p := &WriterPool{}
	p.SetCalibratedDefaultSizeForTest(1024)
	p.SetCalibratedMaxSizeForTest(65536)
	p.CalibrateForTest()
	// The previous sizes are kept when there is nothing to calibrate from.
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(1024))
	assert.Equal(t, p.CalibratedMaxSizeForTest(), uint64(65536))
}

func TestWriterPoolTryCalibrate(t *testing.T) {
	p := &WriterPool{}
	p.SetCallsForTest(1024, 100)
	p.TryCalibrateForTest()
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(1024)) // The calibration has run.
	assert.False(t, p.CalibratingForTest())                         // The flag has been released.
}

func TestWriterPoolTryCalibrateRunning(t *testing.T) {
	p := &WriterPool{}
	p.SetCallsForTest(1024, 100)
	p.SetCalibratingForTest(true) // Simulate a calibration in progress.
	p.TryCalibrateForTest()
	// Nothing has been modified.
	assert.Equal(t, p.CallsForTest(1024), uint64(100))
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(0))
	assert.True(t, p.CalibratingForTest()) // The flag is still owned by the running calibration.
}

func TestWriterPoolGetPreallocated(t *testing.T) {
	p := &WriterPool{}
	p.SetCalibratedDefaultSizeForTest(2048)
	w := p.Get()
	assert.Equal(t, w.Cap(), 2048)
	assert.Equal(t, w.Len(), 0)
}

func TestWriterPoolPutCount(t *testing.T) {
	p := &WriterPool{}
	w := p.Get()
	w.AppendString("foo")
	p.Put(w)
	assert.Equal(t, p.CallsForTest(64), uint64(1))
}

func TestWriterPoolPutEmptyUsesCapacity(t *testing.T) {
	p := &WriterPool{}
	w := p.Get()
	w.Grow(4096)
	w.Reset()
	p.Put(w)
	assert.Equal(t, p.CallsForTest(4096), uint64(1))
	assert.Equal(t, p.CallsForTest(64), uint64(0))
}

func TestWriterPoolPutDiscardOverMaxSize(t *testing.T) {
	p := &WriterPool{}
	p.SetCalibratedMaxSizeForTest(64)
	w := p.Get()
	w.Grow(1024)
	p.Put(w)
	w = p.Get()
	assert.Equal(t, w.Cap(), 0) // The oversized writer has been discarded, so the pool is empty.
}

func TestWriterPoolNoLimitBeforeCalibration(t *testing.T) {
	p := &WriterPool{}
	w := p.Get()
	assert.Equal(t, w.Cap(), 0) // No preallocation before the first calibration.
	w.Grow(1 << 20)
	w.AppendString("foo")
	p.Put(w)
	assert.Equal(t, p.CallsForTest(64), uint64(1)) // The call has been counted.
	assert.Equal(t, w.Len(), 0)                    // There is no limit until the first calibration, so the writer has been kept and reset.
}

func TestWriterPoolStatic(t *testing.T) {
	p := &WriterPool{
		DefaultSize: 2048,
		MaxSize:     4096,
	}
	w := p.Get()
	assert.Equal(t, w.Cap(), 2048) // Static default size.
	w.AppendString("foo")
	p.Put(w)
	assert.True(t, p.NoCallsForTest()) // The calibration is disabled.
	assert.Equal(t, w.Len(), 0)        // The writer has been kept and reset.
	w.Grow(8192)
	w.AppendString("foo")
	p.Put(w)                    // Over MaxSize: discarded.
	assert.Equal(t, w.Len(), 3) // The writer has been discarded, so it has not been reset.
}

func TestWriterPoolMaxSizeNegative(t *testing.T) {
	p := &WriterPool{
		DefaultSize: 64,
		MaxSize:     -1,
	}
	w := p.Get()
	assert.Equal(t, w.Cap(), 64) // Static default size.
	w.Grow(1 << 20)
	w.AppendString("foo")
	p.Put(w)
	assert.True(t, p.NoCallsForTest())
	assert.Equal(t, w.Len(), 0) // There is no limit, so the writer has been kept and reset.
}

func TestWriterPoolPartialStaticDefaultSize(t *testing.T) {
	p := &WriterPool{
		DefaultSize: 2048,
	}
	w := p.Get()
	assert.Equal(t, w.Cap(), 2048) // Static default size.
	w.AppendString("foo")
	p.Put(w)
	assert.Equal(t, p.CallsForTest(64), uint64(1)) // The maximum size is calibrated, so the calls are counted.
	p.CalibrateForTest()
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(64))
	assert.Equal(t, p.DefaultSizeForTest(), 2048) // The static default size wins over the calibrated one.
}

func TestWriterPoolCalibrateClampMaxSize(t *testing.T) {
	p := &WriterPool{
		MaxSize: 4096,
	}
	p.SetCallsForTest(1<<20, 100) // The dominant bucket is above MaxSize.
	p.CalibrateForTest()
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(4096)) // Clamped to MaxSize.
	w := p.Get()
	assert.Equal(t, w.Cap(), 4096) // The preallocated size does not exceed MaxSize.
	w.AppendString("foo")
	p.Put(w)
	assert.Equal(t, p.CallsForTest(64), uint64(1)) // The calls are still counted.
	assert.Equal(t, w.Len(), 0)                    // The writer has been kept and reset, not discarded.
}

func TestWriterPoolMaxSizeFlooredByDefaultSize(t *testing.T) {
	p := &WriterPool{
		DefaultSize: 2048,
	}
	p.SetCallsForTest(64, 42000) // Small traffic dominates, so the calibrated maximum size is 64.
	p.CalibrateForTest()
	w := p.Get()
	assert.Equal(t, w.Cap(), 2048) // The static default size is still honored.
	w.AppendString("foo")
	p.Put(w)
	assert.Equal(t, w.Len(), 0) // The preallocated writer is accepted back, not discarded.
}

func TestWriterPoolConcurrent(t *testing.T) {
	p := &WriterPool{}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 5500 {
				w := p.Get()
				w.AppendString("foo")
				p.Put(w)
			}
		})
	}
	wg.Wait()
	// All the writers are small, so the calibrated sizes are the minimum bucket size.
	assert.Equal(t, p.CalibratedDefaultSizeForTest(), uint64(64))
	assert.Equal(t, p.CalibratedMaxSizeForTest(), uint64(64))
}

func BenchmarkWriterPool(b *testing.B) {
	p := &WriterPool{}
	for b.Loop() {
		w := p.Get()
		w.AppendString(testWriterPoolData)
		p.Put(w)
	}
}

func BenchmarkWriterPoolStatic(b *testing.B) {
	p := &WriterPool{
		DefaultSize: 512,
		MaxSize:     1 << 16,
	}
	for b.Loop() {
		w := p.Get()
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

func BenchmarkWriterPoolParallel(b *testing.B) {
	p := &WriterPool{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := p.Get()
			w.AppendString(testWriterPoolData)
			p.Put(w)
		}
	})
}

func BenchmarkWriterPoolParallelSizes(b *testing.B) {
	p := &WriterPool{}
	sizes := []int{65, 129, 257} // spreads the calls over 3 adjacent buckets, so that concurrent workers contend on distinct counters.
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			w := p.Get()
			w.AppendString(testWriterPoolData[:sizes[i%len(sizes)]])
			p.Put(w)
		}
	})
}
