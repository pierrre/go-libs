package bytesutil

import (
	"io"
	"unicode/utf8"
)

// Writer is a []byte that provides methods for appending data.
//
// The zero value is ready to use.
type Writer []byte

// Append appends p to the writer.
func (w *Writer) Append(p []byte) {
	*w = append(*w, p...)
}

// Write appends p to the writer.
//
// Write always returns len(p), nil.
func (w *Writer) Write(p []byte) (n int, err error) {
	*w = append(*w, p...)
	return len(p), nil
}

// AppendString appends s to the writer.
func (w *Writer) AppendString(s string) {
	*w = append(*w, s...)
}

// WriteString appends s to the writer.
//
// WriteString always returns len(s), nil.
func (w *Writer) WriteString(s string) (n int, err error) {
	*w = append(*w, s...)
	return len(s), nil
}

// AppendByte appends c to the writer.
func (w *Writer) AppendByte(c byte) {
	*w = append(*w, c)
}

// WriteByte appends c to the writer.
//
// WriteByte always returns nil.
func (w *Writer) WriteByte(c byte) error {
	*w = append(*w, c)
	return nil
}

// AppendRune appends the UTF-8 encoding of r to the writer.
func (w *Writer) AppendRune(r rune) {
	*w = utf8.AppendRune(*w, r)
}

// WriteRune appends the UTF-8 encoding of r to the writer.
//
// It returns the number of bytes written and a nil error.
func (w *Writer) WriteRune(r rune) (n int, err error) {
	if uint32(r) < utf8.RuneSelf { // Compare as uint32 to correctly handle negative runes.
		*w = append(*w, byte(r))
		return 1, nil
	}
	l := len(*w)
	*w = utf8.AppendRune(*w, r)
	return len(*w) - l, nil
}

// ReadFrom reads data from r until EOF and appends it to the writer.
// It returns the number of bytes read and any error encountered.
func (w *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		w.grow(4096)
		m, e := r.Read((*w)[len(*w):cap(*w)])
		if m < 0 {
			panic("bytesutil.Writer.ReadFrom: reader returned negative count from Read")
		}
		*w = (*w)[:len(*w)+m]
		n += int64(m)
		if e != nil {
			if e == io.EOF {
				return n, nil
			}
			return n, e //nolint:wrapcheck // Not needed.
		}
	}
}

// Reset resets the writer to be empty, while keeping the underlying storage.
func (w *Writer) Reset() {
	*w = (*w)[:0]
}

// Clear resets the writer to be empty and clears the underlying storage.
func (w *Writer) Clear() {
	*w = (*w)[:0]
	clear((*w)[:cap(*w)])
}

// Grow grows the writer's capacity, if necessary, to guarantee space for another n bytes.
//
// After Grow(n), at least n bytes can be appended to the writer without another allocation.
func (w *Writer) Grow(n int) {
	if n < 0 {
		panic("bytesutil.Writer.Grow: negative count")
	}
	w.grow(n)
}

func (w *Writer) grow(n int) {
	n -= cap(*w) - len(*w)
	if n > 0 {
		*w = append((*w)[:cap(*w)], make(Writer, n)...)[:len(*w)]
	}
}

// Len returns the number of bytes currently stored in the writer.
func (w Writer) Len() int {
	return len(w)
}

// Cap returns the capacity of the writer.
func (w Writer) Cap() int {
	return cap(w)
}

// Available returns how many unused bytes remain in the writer's capacity.
func (w Writer) Available() int {
	return cap(w) - len(w)
}

// AvailableBuffer returns an empty slice backed by the writer's unused capacity.
//
// The returned slice has length 0 and capacity Available(). It is intended for immediate use with append.
func (w Writer) AvailableBuffer() []byte {
	return w[len(w):]
}

// Bytes returns the contents of the writer.
func (w Writer) Bytes() []byte {
	return w
}

// Clone returns a copy of the writer's contents.
func (w Writer) Clone() Writer {
	if w == nil {
		return nil
	}
	return append(Writer{}, w...)
}

// String returns the contents of the writer as a string.
func (w Writer) String() string {
	return string(w)
}
