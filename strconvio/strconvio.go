// Package strconvio provides utilities to write values string representation to a writer.
package strconvio

import (
	"io"
	"strconv"
	"sync"
)

// WriteBool writes the string representation of the bool to the writer.
func WriteBool(w io.Writer, b bool) (int, error) {
	return io.WriteString(w, strconv.FormatBool(b)) //nolint:wrapcheck // It's fine.
}

// WriteFloat writes the string representation of the float to the writer.
func WriteFloat(w io.Writer, f float64, fmt byte, prec, bitSize int) (int, error) {
	bp := bytesPool.Get().(*[]byte) //nolint:forcetypeassert // The pool only contains *[]byte.
	defer bytesPool.Put(bp)
	*bp = strconv.AppendFloat((*bp)[:0], f, fmt, prec, bitSize)
	return w.Write(*bp) //nolint:wrapcheck // It's fine.
}

// WriteInt writes the string representation of the integer to the writer.
func WriteInt(w io.Writer, i int64, base int) (int, error) {
	if 0 <= i && i < 100 && base == 10 {
		return io.WriteString(w, strconv.FormatInt(i, base)) //nolint:wrapcheck // It's fine.
	}
	bp := bytesPool.Get().(*[]byte) //nolint:forcetypeassert // The pool only contains *[]byte.
	defer bytesPool.Put(bp)
	*bp = strconv.AppendInt((*bp)[:0], i, base)
	return w.Write(*bp) //nolint:wrapcheck // It's fine.
}

// WriteUint writes the string representation of the unsigned integer to the writer.
func WriteUint(w io.Writer, i uint64, base int) (int, error) {
	if i < 100 && base == 10 {
		return io.WriteString(w, strconv.FormatUint(i, base)) //nolint:wrapcheck // It's fine.
	}
	bp := bytesPool.Get().(*[]byte) //nolint:forcetypeassert // The pool only contains *[]byte.
	defer bytesPool.Put(bp)
	*bp = strconv.AppendUint((*bp)[:0], i, base)
	return w.Write(*bp) //nolint:wrapcheck // It's fine.
}

// WriteQuote writes the quoted string to the writer.
func WriteQuote(w io.Writer, s string) (int, error) {
	bp := bytesPool.Get().(*[]byte) //nolint:forcetypeassert // The pool only contains *[]byte.
	defer bytesPool.Put(bp)
	*bp = strconv.AppendQuote((*bp)[:0], s)
	return w.Write(*bp) //nolint:wrapcheck // It's fine.
}

var bytesPool = sync.Pool{
	New: func() any {
		var b []byte
		return &b
	},
}
