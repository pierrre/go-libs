// Package runtimeutil provides runtime utilities.
package runtimeutil

import (
	"runtime"

	"github.com/pierrre/go-libs/syncutil"
)

// GetCallers returns the callers.
// See [runtime.Callers].
func GetCallers(skip int) []uintptr {
	skip += 2 // Skip [GetCallers] and [runtime.Callers].
	pc := callersPool.Get()
	for {
		n := runtime.Callers(skip, pc)
		if n < len(pc) {
			res := make([]uintptr, n)
			copy(res, pc[:n])
			callersPool.Put(pc)
			return res
		}
		pc = make([]uintptr, 2*len(pc))
	}
}

var callersPool = syncutil.ValuePool[[]uintptr]{
	New: func() []uintptr {
		return make([]uintptr, 1024)
	},
}
