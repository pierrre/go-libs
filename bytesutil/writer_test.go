package bytesutil_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/bytesutil"
)

func ExampleWriter() {
	w := new(Writer)
	w.Append([]byte("a"))
	w.AppendString("b")
	w.AppendByte('c')
	w.AppendRune('d')
	fmt.Println(w.String())
	// Output: abcd
}

func TestWriterAppend(t *testing.T) {
	w := new(Writer)
	w.Append([]byte("abc"))
	assert.BytesEqual(t, *w, []byte("abc"))
}

func BenchmarkWriterAppend(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		w.Append([]byte("abc"))
		w.Reset()
	}
}

func TestWriterWrite(t *testing.T) {
	w := new(Writer)
	n, err := w.Write([]byte("abc")) //nolint:gocritic // Don't want to rewrite to WriteString.
	assert.NoError(t, err)
	assert.Equal(t, n, 3)
	assert.BytesEqual(t, *w, []byte("abc"))
}

func BenchmarkWriterWrite(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		_, _ = w.Write([]byte("abc")) //nolint:gocritic // Don't want to rewrite to WriteString.
		w.Reset()
	}
}

func TestWriterAppendString(t *testing.T) {
	w := new(Writer)
	w.AppendString("abc")
	assert.BytesEqual(t, *w, []byte("abc"))
}

func BenchmarkWriterAppendString(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		w.AppendString("abc")
		w.Reset()
	}
}

func TestWriterWriteString(t *testing.T) {
	w := new(Writer)
	n, err := w.WriteString("abc")
	assert.NoError(t, err)
	assert.Equal(t, n, 3)
	assert.BytesEqual(t, *w, []byte("abc"))
}

func BenchmarkWriterWriteString(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		_, _ = w.WriteString("abc")
		w.Reset()
	}
}

func TestWriterAppendByte(t *testing.T) {
	w := new(Writer)
	w.AppendByte('a')
	assert.BytesEqual(t, *w, []byte("a"))
}

func BenchmarkWriterAppendByte(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		w.AppendByte('a')
		w.Reset()
	}
}

func TestWriterWriteByte(t *testing.T) {
	w := new(Writer)
	err := w.WriteByte('a')
	assert.NoError(t, err)
	assert.BytesEqual(t, *w, []byte("a"))
}

func BenchmarkWriterWriteByte(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		_ = w.WriteByte('a')
		w.Reset()
	}
}

func TestWriterAppendRune(t *testing.T) {
	w := new(Writer)
	w.AppendRune('é')
	assert.BytesEqual(t, *w, []byte("é"))
}

func BenchmarkWriterAppendRune(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		w.AppendRune('é')
		w.Reset()
	}
}

func TestWriterWriteRuneSimple(t *testing.T) {
	w := new(Writer)
	n, err := w.WriteRune('a') //nolint:gocritic // Don't want to rewrite to WriteByte.
	assert.NoError(t, err)
	assert.Equal(t, n, 1)
	assert.BytesEqual(t, *w, []byte("a"))
}

func TestWriterWriteRuneMulti(t *testing.T) {
	w := new(Writer)
	n, err := w.WriteRune('é')
	assert.NoError(t, err)
	assert.Equal(t, n, 2)
	assert.BytesEqual(t, *w, []byte("é"))
}

func BenchmarkWriterWriteRuneSimple(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		_, _ = w.WriteRune('a') //nolint:gocritic // Don't want to rewrite to WriteByte.
		w.Reset()
	}
}

func BenchmarkWriterWriteRuneMulti(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		_, _ = w.WriteRune('é')
		w.Reset()
	}
}

func TestWriterReadFrom(t *testing.T) {
	w := new(Writer)
	n, err := w.ReadFrom(bytes.NewReader([]byte("abc")))
	assert.NoError(t, err)
	assert.Equal(t, n, 3)
	assert.BytesEqual(t, *w, []byte("abc"))
}

func TestWriterReadFromError(t *testing.T) {
	w := new(Writer)
	r := readerFunc(func(p []byte) (n int, err error) {
		return 0, errors.New("error")
	})
	n, err := w.ReadFrom(r)
	assert.Error(t, err)
	assert.Equal(t, n, 0)
	assert.SliceEmpty(t, *w)
}

func TestWriterReadFromPanicNegative(t *testing.T) {
	w := new(Writer)
	r := readerFunc(func(p []byte) (n int, err error) {
		return -1, nil
	})
	assert.Panics(t, func() {
		_, _ = w.ReadFrom(r)
	})
	assert.SliceEmpty(t, *w)
}

type readerFunc func(p []byte) (n int, err error)

func (f readerFunc) Read(p []byte) (n int, err error) {
	return f(p)
}

func BenchmarkWriterReadFrom(b *testing.B) {
	w := new(Writer)
	r := new(bytes.Reader)
	for b.Loop() {
		r.Reset([]byte("abc"))
		_, _ = w.ReadFrom(r)
		w.Reset()
	}
}

func TestWriterReset(t *testing.T) {
	w := new(Writer("abc"))
	w.Reset()
	assert.SliceEmpty(t, *w)
	assert.BytesEqual(t, []byte(*w)[:cap(*w)], []byte("abc"))
}

func TestWriterClear(t *testing.T) {
	w := new(Writer("abc"))
	w.Clear()
	assert.SliceEmpty(t, *w)
	assert.SliceEqual(t, []byte(*w)[:cap(*w)], make([]byte, 3))
}

func TestWriterGrow(t *testing.T) {
	w := new(Writer)
	w.Grow(3)
	assert.GreaterOrEqual(t, cap([]byte(*w)), 3)
}

func TestWriterLen(t *testing.T) {
	w := new(Writer("abc"))
	assert.Equal(t, w.Len(), 3)
}

func TestWriterCap(t *testing.T) {
	w := new(Writer("abc"))
	assert.Equal(t, w.Cap(), 3)
}

func TestWriterAvailable(t *testing.T) {
	w := new(Writer(make([]byte, 1, 3)))
	assert.Equal(t, w.Available(), 2)
}

func TestWriterAvailableBuffer(t *testing.T) {
	w := new(Writer(make([]byte, 1, 3)))
	buf := w.AvailableBuffer()
	assert.SliceEmpty(t, buf)
	assert.Equal(t, cap(buf), 2)
}

func TestWriterBytes(t *testing.T) {
	w := new(Writer("abc"))
	b := w.Bytes()
	assert.SliceEqual(t, b, []byte("abc"))
	assert.Equal(t, &b[0], &(*w)[0])
}

func TestWriterCloneBytes(t *testing.T) {
	w := new(Writer("abc"))
	clone := w.CloneBytes()
	assert.SliceEqual(t, clone, []byte("abc"))
	assert.NotEqual(t, &clone[0], &(*w)[0])
}

func TestWriterString(t *testing.T) {
	w := new(Writer("abc"))
	s := w.String()
	assert.Equal(t, s, "abc")
}
