package weakutil

import (
	"runtime/debug"
	"testing"
)

func disableGC(tb testing.TB) {
	tb.Helper()
	old := debug.SetGCPercent(-1)
	tb.Cleanup(func() { debug.SetGCPercent(old) })
}
