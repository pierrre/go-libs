// Package panichandle handles panic.
package panichandle

// Handler handles panic.
var Handler = DefaultHandler

// DefaultHandler is the default Handler.
// It panics.
func DefaultHandler(r interface{}) {
	panic(r)
}

// Recover recovers panic and call Handler.
func Recover() {
	r := recover()
	if r != nil {
		Handler(r)
	}
}
