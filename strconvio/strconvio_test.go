package strconvio_test

import (
	"bytes"
	"io"
	"math"
	"strconv"
	"testing"

	"github.com/pierrre/assert"
	"github.com/pierrre/assert/assertauto"
	. "github.com/pierrre/go-libs/strconvio"
)

var writeBoolTestCases = []struct {
	b        bool
	expected string
}{
	{
		b:        false,
		expected: "false",
	},
	{
		b:        true,
		expected: "true",
	},
}

var testWriteBoolValues = []bool{false, true}

func TestWriteBool(t *testing.T) {
	for _, v := range testWriteBoolValues {
		buf := new(bytes.Buffer)
		n, err := WriteBool(buf, v)
		assert.NoError(t, err)
		s := buf.String()
		assertauto.Equal(t, s)
		assert.StringLen(t, s, n)
		assert.AllocsPerRun(t, 100, func() {
			_, _ = WriteBool(io.Discard, v)
		}, 0)
	}
}

func BenchmarkWriteBool(b *testing.B) {
	for _, tc := range writeBoolTestCases {
		b.Run(strconv.FormatBool(tc.b), func(b *testing.B) {
			for range b.N {
				_, _ = WriteBool(io.Discard, tc.b)
			}
		})
	}
}

var testFloatValues = []float64{
	0,
	1,
	12.34,
	-1,
	math.Inf(1),
	math.Inf(-1),
	math.NaN(),
}

func TestWriteFloat(t *testing.T) {
	for _, v := range testFloatValues {
		buf := new(bytes.Buffer)
		n, err := WriteFloat(buf, v, 'f', -1, 64)
		assert.NoError(t, err)
		s := buf.String()
		assertauto.Equal(t, s)
		assert.StringLen(t, s, n)
		assert.AllocsPerRun(t, 100, func() {
			_, _ = WriteFloat(io.Discard, v, 'f', -1, 64)
		}, 0)
	}
}

func BenchmarkWriteFloat(b *testing.B) {
	for _, v := range testFloatValues {
		b.Run(strconv.FormatFloat(v, 'f', -1, 64), func(b *testing.B) {
			for range b.N {
				_, _ = WriteFloat(io.Discard, v, 'f', -1, 64)
			}
		})
	}
}

var testWriteComplexValues = []complex128{}

func init() {
	for _, real := range testFloatValues {
		for _, imag := range testFloatValues {
			testWriteComplexValues = append(testWriteComplexValues, complex(real, imag))
		}
	}
}

func TestWriteComplex(t *testing.T) {
	for _, v := range testWriteComplexValues {
		buf := new(bytes.Buffer)
		n, err := WriteComplex(buf, v, 'f', -1, 128)
		assert.NoError(t, err)
		s := buf.String()
		assertauto.Equal(t, s)
		assert.StringLen(t, s, n)
		assert.AllocsPerRun(t, 100, func() {
			_, _ = WriteComplex(io.Discard, v, 'f', -1, 128)
		}, 0)
	}
}

func TestWriteComplexPanicBitSize(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = WriteComplex(io.Discard, 0, 'f', -1, 0)
	})
}

func BenchmarkWriteComplex(b *testing.B) {
	for _, v := range testWriteComplexValues {
		b.Run(strconv.FormatComplex(v, 'f', -1, 128), func(b *testing.B) {
			for range b.N {
				_, _ = WriteComplex(io.Discard, v, 'f', -1, 128)
			}
		})
	}
}

var testWriteIntValues = []int64{
	0,
	1,
	-1,
	1234567890,
	-1234567890,
}

func TestWriteInt(t *testing.T) {
	for _, v := range testWriteIntValues {
		buf := new(bytes.Buffer)
		n, err := WriteInt(buf, v, 10)
		assert.NoError(t, err)
		s := buf.String()
		assertauto.Equal(t, s)
		assert.StringLen(t, s, n)
		assert.AllocsPerRun(t, 100, func() {
			_, _ = WriteInt(io.Discard, v, 10)
		}, 0)
	}
}

func BenchmarkWriteInt(b *testing.B) {
	for _, v := range testWriteIntValues {
		b.Run(strconv.FormatInt(v, 10), func(b *testing.B) {
			for range b.N {
				_, _ = WriteInt(io.Discard, v, 10)
			}
		})
	}
}

var testWriteUintValues = []uint64{
	0,
	1,
	1234567890,
}

func TestWriteUint(t *testing.T) {
	for _, v := range testWriteUintValues {
		buf := new(bytes.Buffer)
		n, err := WriteUint(buf, v, 10)
		assert.NoError(t, err)
		s := buf.String()
		assertauto.Equal(t, s)
		assert.StringLen(t, s, n)
		assert.AllocsPerRun(t, 100, func() {
			_, _ = WriteUint(io.Discard, v, 10)
		}, 0)
	}
}

func BenchmarkWriteUint(b *testing.B) {
	for _, v := range testWriteUintValues {
		b.Run(strconv.FormatUint(v, 10), func(b *testing.B) {
			for range b.N {
				_, _ = WriteUint(io.Discard, v, 10)
			}
		})
	}
}

var testWriteQuoteValues = []string{
	"",
	"test",
	"\"",
}

func TestWriteQuote(t *testing.T) {
	for _, v := range testWriteQuoteValues {
		buf := new(bytes.Buffer)
		n, err := WriteQuote(buf, v)
		assert.NoError(t, err)
		s := buf.String()
		assertauto.Equal(t, s)
		assert.StringLen(t, s, n)
		assert.AllocsPerRun(t, 100, func() {
			_, _ = WriteQuote(io.Discard, v)
		}, 0)
	}
}

func BenchmarkWriteQuote(b *testing.B) {
	for _, v := range testWriteQuoteValues {
		b.Run(strconv.Quote(v), func(b *testing.B) {
			for range b.N {
				_, _ = WriteQuote(io.Discard, v)
			}
		})
	}
}
