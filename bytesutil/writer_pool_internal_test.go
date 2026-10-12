package bytesutil

// The helpers in this file expose the internals of WriterPool to the external test package.

// SetCallsForTest sets the number of calls to Put counted for the bucket of the given size.
func (p *WriterPool) SetCallsForTest(size int, n uint64) {
	p.calls[writerPoolIndex(size)].n.Store(n)
}

// CallsForTest returns the number of calls to Put counted for the bucket of the given size.
func (p *WriterPool) CallsForTest(size int) uint64 {
	return p.calls[writerPoolIndex(size)].n.Load()
}

// NoCallsForTest returns true if all the calls counters are zero.
func (p *WriterPool) NoCallsForTest() bool {
	for i := range writerPoolSteps {
		if p.calls[i].n.Load() != 0 {
			return false
		}
	}
	return true
}

// CalibratedDefaultSizeForTest returns the calibrated default size.
func (p *WriterPool) CalibratedDefaultSizeForTest() uint64 {
	return p.calibratedDefaultSize.Load()
}

// SetCalibratedDefaultSizeForTest sets the calibrated default size.
func (p *WriterPool) SetCalibratedDefaultSizeForTest(n uint64) {
	p.calibratedDefaultSize.Store(n)
}

// CalibratedMaxSizeForTest returns the calibrated maximum size.
func (p *WriterPool) CalibratedMaxSizeForTest() uint64 {
	return p.calibratedMaxSize.Load()
}

// SetCalibratedMaxSizeForTest sets the calibrated maximum size.
func (p *WriterPool) SetCalibratedMaxSizeForTest(n uint64) {
	p.calibratedMaxSize.Store(n)
}

// CalibratingForTest returns true if a calibration is running.
func (p *WriterPool) CalibratingForTest() bool {
	return p.calibrating.Load()
}

// SetCalibratingForTest sets the calibrating flag, which simulates a calibration in progress.
func (p *WriterPool) SetCalibratingForTest(v bool) {
	p.calibrating.Store(v)
}

// CalibrateForTest calibrates the pool, without the single-flight protection of tryCalibrate.
func (p *WriterPool) CalibrateForTest() {
	p.calibrate()
}

// TryCalibrateForTest calibrates the pool, unless a calibration is already running.
func (p *WriterPool) TryCalibrateForTest() {
	p.tryCalibrate()
}

// DefaultSizeForTest returns the default size used to preallocate the new writers.
func (p *WriterPool) DefaultSizeForTest() int {
	return p.defaultSize()
}

// WriterPoolIndexForTest returns the bucket index for a size.
func WriterPoolIndexForTest(n int) int {
	return writerPoolIndex(n)
}
