package bytesutil_test

import (
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
	assert.SliceEqual(t, w.Bytes, []byte("abc"))
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
	assert.SliceEqual(t, w.Bytes, []byte("abc"))
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
	assert.SliceEqual(t, w.Bytes, []byte("abc"))
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
	assert.SliceEqual(t, w.Bytes, []byte("abc"))
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
	assert.SliceEqual(t, w.Bytes, []byte("a"))
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
	assert.SliceEqual(t, w.Bytes, []byte("a"))
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
	assert.SliceEqual(t, w.Bytes, []byte("é"))
}

func BenchmarkWriterAppendRune(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		w.AppendRune('é')
		w.Reset()
	}
}

func TestWriterWriteRune(t *testing.T) {
	w := new(Writer)
	n, err := w.WriteRune('é')
	assert.NoError(t, err)
	assert.Equal(t, n, 2)
	assert.SliceEqual(t, w.Bytes, []byte("é"))
}

func BenchmarkWriterWriteRune(b *testing.B) {
	w := new(Writer)
	for b.Loop() {
		_, _ = w.WriteRune('é')
		w.Reset()
	}
}

func TestWriterReset(t *testing.T) {
	w := &Writer{Bytes: []byte("abc")}
	w.Reset()
	assert.SliceEmpty(t, w.Bytes)
	assert.SliceEqual(t, w.Bytes[:cap(w.Bytes)], []byte("abc"))
}

func TestWriterClear(t *testing.T) {
	w := &Writer{Bytes: []byte("abc")}
	w.Clear()
	assert.SliceEmpty(t, w.Bytes)
	assert.SliceEqual(t, w.Bytes[:cap(w.Bytes)], make([]byte, 3))
}

func TestWriterGrow(t *testing.T) {
	w := new(Writer)
	w.Grow(3)
	assert.GreaterOrEqual(t, cap(w.Bytes), 3)
}

func TestWriterLen(t *testing.T) {
	w := &Writer{Bytes: []byte("abc")}
	assert.Equal(t, w.Len(), 3)
}

func TestWriterCap(t *testing.T) {
	w := &Writer{Bytes: make([]byte, 0, 3)}
	assert.Equal(t, w.Cap(), 3)
}

func TestWriterAvailable(t *testing.T) {
	w := &Writer{Bytes: make([]byte, 1, 3)}
	assert.Equal(t, w.Available(), 2)
}

func TestWriterAvailableBuffer(t *testing.T) {
	w := &Writer{Bytes: make([]byte, 1, 3)}
	buf := w.AvailableBuffer()
	assert.SliceEmpty(t, buf)
	assert.Equal(t, cap(buf), 2)
}

func TestWriterCloneBytes(t *testing.T) {
	w := &Writer{Bytes: []byte("abc")}
	clone := w.CloneBytes()
	assert.SliceEqual(t, clone, []byte("abc"))
	assert.NotEqual(t, &clone[0], &w.Bytes[0])
}

func TestWriterString(t *testing.T) {
	w := &Writer{Bytes: []byte("abc")}
	s := w.String()
	assert.Equal(t, s, "abc")
}
