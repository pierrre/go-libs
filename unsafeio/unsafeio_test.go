package unsafeio_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/pierrre/assert"
	. "github.com/pierrre/go-libs/unsafeio"
)

func TestWriteString(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteString(buf, "test")
	assert.NoError(t, err)
	assert.Equal(t, n, 4)
	assert.Equal(t, buf.String(), "test")
}

func TestWriteStringAllocs(t *testing.T) {
	assert.AllocsPerRun(t, 100, func() {
		_, _ = WriteString(io.Discard, "test")
	}, 0)
}

func BenchmarkWriteString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = WriteString(io.Discard, "test")
	}
}

func BenchmarkIOWriteString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = io.WriteString(io.Discard, "test")
	}
}

func BenchmarkIOWriteStringConvert(b *testing.B) {
	var w io.Writer = &testWriter{}
	for i := 0; i < b.N; i++ {
		_, _ = io.WriteString(w, "test")
	}
}

type testWriter struct{}

func (w *testWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
