// Package panichandle handles panic.
package panichandle

// Handler handles panic.
//
// By default there is no handler.
var Handler func(r any)

// Recover recovers panic and call Handler.
//
// If there is no handler, it doesn't recover.
//
// It should be called in defer.
func Recover() {
	if Handler != nil {
		r := recover()
		if r != nil {
			Handler(r)
		}
	}
}
