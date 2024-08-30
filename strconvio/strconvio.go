// Package strconvio provides utilities to write values string representation to a writer.
package strconvio

import (
	"io"
	"strconv"

	"github.com/pierrre/go-libs/syncutil"
	"github.com/pierrre/go-libs/unsafeio"
)

// WriteBool writes the string representation of the bool to the writer.
func WriteBool(w io.Writer, b bool) (int, error) {
	return unsafeio.WriteString(w, strconv.FormatBool(b)) //nolint:wrapcheck // It's fine.
}

// WriteFloat writes the string representation of the float to the writer.
func WriteFloat(w io.Writer, f float64, fmt byte, prec, bitSize int) (int, error) {
	bp := getFloatBytes(f, fmt, prec, bitSize)
	n, err := w.Write(*bp)
	bytesPool.Put(bp)
	return n, err //nolint:wrapcheck // It's fine.
}

func getFloatBytes(f float64, fmt byte, prec, bitSize int) *[]byte {
	bp := bytesPool.Get()
	*bp = strconv.AppendFloat((*bp)[:0], f, fmt, prec, bitSize)
	return bp
}

// WriteComplex writes the string representation of the complex to the writer.
func WriteComplex(w io.Writer, c complex128, fmt byte, prec, bitSize int) (int, error) {
	bp := getComplexBytes(c, fmt, prec, bitSize)
	n, err := w.Write(*bp)
	bytesPool.Put(bp)
	return n, err //nolint:wrapcheck // It's fine.
}

func getComplexBytes(c complex128, fmt byte, prec, bitSize int) *[]byte {
	bp := bytesPool.Get()
	*bp = appendComplex((*bp)[:0], c, fmt, prec, bitSize)
	return bp
}

func appendComplex(dst []byte, c complex128, fmt byte, prec, bitSize int) []byte {
	if bitSize != 64 && bitSize != 128 {
		panic("invalid bitSize")
	}
	bitSize >>= 1 // complex64 uses float32 internally
	bpReal := getFloatBytes(real(c), fmt, prec, bitSize)
	bpImag := getFloatBytes(imag(c), fmt, prec, bitSize)
	dst = append(dst, '(')
	dst = append(dst, *bpReal...)
	// Check if imaginary part has a sign. If not, add one.
	if (*bpImag)[0] != '+' && (*bpImag)[0] != '-' {
		dst = append(dst, '+')
	}
	dst = append(dst, *bpImag...)
	dst = append(dst, "i)"...)
	bytesPool.Put(bpReal)
	bytesPool.Put(bpImag)
	return dst
}

// WriteInt writes the string representation of the signed integer to the writer.
func WriteInt(w io.Writer, i int64, base int) (int, error) {
	if 0 <= i && i < 100 && base == 10 {
		return unsafeio.WriteString(w, strconv.FormatInt(i, base)) //nolint:wrapcheck // It's fine.
	}
	bp := bytesPool.Get()
	*bp = strconv.AppendInt((*bp)[:0], i, base)
	n, err := w.Write(*bp)
	bytesPool.Put(bp)
	return n, err //nolint:wrapcheck // It's fine.
}

// WriteUint writes the string representation of the unsigned integer to the writer.
func WriteUint(w io.Writer, i uint64, base int) (int, error) {
	if i < 100 && base == 10 {
		return unsafeio.WriteString(w, strconv.FormatUint(i, base)) //nolint:wrapcheck // It's fine.
	}
	bp := bytesPool.Get()
	*bp = strconv.AppendUint((*bp)[:0], i, base)
	n, err := w.Write(*bp)
	bytesPool.Put(bp)
	return n, err //nolint:wrapcheck // It's fine.
}

// WriteQuote writes the quoted string to the writer.
func WriteQuote(w io.Writer, s string) (int, error) {
	if s == "" {
		return w.Write(emptyQuotes) //nolint:wrapcheck // It's fine.
	}
	bp := bytesPool.Get()
	*bp = strconv.AppendQuote((*bp)[:0], s)
	n, err := w.Write(*bp)
	bytesPool.Put(bp)
	return n, err //nolint:wrapcheck // It's fine.
}

var emptyQuotes = []byte(`""`)

var bytesPool = syncutil.PoolFor[*[]byte]{
	New: func() *[]byte {
		return new([]byte)
	},
}
