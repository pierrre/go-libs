package bytesutil

import (
	"slices"
	"unicode/utf8"
)

// Writer appends data to a byte slice.
//
// The zero value for Writer is ready to use.
type Writer struct {
	// Bytes holds the accumulated data.
	Bytes []byte
}

// Append appends p to w.Bytes.
func (w *Writer) Append(p []byte) {
	w.Bytes = append(w.Bytes, p...)
}

// Write appends p to w.Bytes.
//
// Write always returns len(p), nil.
func (w *Writer) Write(p []byte) (n int, err error) {
	w.Append(p)
	return len(p), nil
}

// AppendString appends s to w.Bytes.
func (w *Writer) AppendString(s string) {
	w.Bytes = append(w.Bytes, s...)
}

// WriteString appends s to w.Bytes.
//
// WriteString always returns len(s), nil.
func (w *Writer) WriteString(s string) (n int, err error) {
	w.AppendString(s)
	return len(s), nil
}

// AppendByte appends c to w.Bytes.
func (w *Writer) AppendByte(c byte) {
	w.Bytes = append(w.Bytes, c)
}

// WriteByte appends c to w.Bytes.
//
// WriteByte always returns nil.
func (w *Writer) WriteByte(c byte) error {
	w.AppendByte(c)
	return nil
}

// AppendRune appends the UTF-8 encoding of r to w.Bytes.
func (w *Writer) AppendRune(r rune) {
	w.Bytes = utf8.AppendRune(w.Bytes, r)
}

// WriteRune appends the UTF-8 encoding of r to w.Bytes.
//
// It returns the number of bytes written and a nil error.
func (w *Writer) WriteRune(r rune) (n int, err error) {
	l := len(w.Bytes)
	w.AppendRune(r)
	return len(w.Bytes) - l, nil
}

// Reset resets w.Bytes to be empty, while keeping the underlying storage.
func (w *Writer) Reset() {
	w.Bytes = w.Bytes[:0]
}

// Clear resets w.Bytes to be empty and clears the underlying storage.
func (w *Writer) Clear() {
	w.Reset()
	clear(w.Bytes[:cap(w.Bytes)])
}

// Grow grows w.Bytes's capacity, if necessary, to guarantee space for another n bytes.
//
// After Grow(n), at least n bytes can be appended to w.Bytes without another allocation.
func (w *Writer) Grow(n int) {
	w.Bytes = slices.Grow(w.Bytes, n)
}

// Len returns the number of bytes currently stored in w.Bytes.
func (w *Writer) Len() int {
	return len(w.Bytes)
}

// Cap returns the capacity of w.Bytes.
func (w *Writer) Cap() int {
	return cap(w.Bytes)
}

// Available returns how many unused bytes remain in w.Bytes's capacity.
func (w *Writer) Available() int {
	return cap(w.Bytes) - len(w.Bytes)
}

// AvailableBuffer returns an empty slice backed by w.Bytes's unused capacity.
//
// The returned slice has length 0 and capacity Available(). It is intended for immediate use with append.
func (w *Writer) AvailableBuffer() []byte {
	return w.Bytes[len(w.Bytes):]
}

// CloneBytes returns a copy of w.Bytes.
func (w *Writer) CloneBytes() []byte {
	return slices.Clone(w.Bytes)
}

// String returns the contents of w.Bytes as a string.
func (w *Writer) String() string {
	return string(w.Bytes)
}
