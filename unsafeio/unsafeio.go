// Package unsafeio provides unsafe IO operations.
package unsafeio

import (
	"io"
	"unsafe" //nolint:depguard // The current package is unsafe.
)

// WriteString writes a string to a [io.Writer].
func WriteString(w io.Writer, s string) (int, error) {
	return w.Write(unsafe.Slice(unsafe.StringData(s), len(s))) //nolint:wrapcheck // The error is not wrapped.
}
