// Package unsafeio provides unsafe IO operations.
//
//nolint:gosec // It uses unsafe.
package unsafeio

import (
	"io"
	"unsafe" //nolint:depguard // The current package is unsafe.
)

// WriteString writes a string to a [io.Writer].
func WriteString(w io.Writer, s string) (int, error) {
	return w.Write( //nolint:gosec,wrapcheck // The error is not wrapped.
		unsafe.Slice(
			unsafe.StringData(s),
			len(s),
		),
	)
}
