package strconvio

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/pierrre/assert"
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

func TestWriteBool(t *testing.T) {
	for _, tc := range writeBoolTestCases {
		t.Run(strconv.FormatBool(tc.b), func(t *testing.T) {
			buf := new(bytes.Buffer)
			n, err := WriteBool(buf, tc.b)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected), n)
			assert.Equal(t, tc.expected, buf.String())
			assert.AllocsPerRun(t, 100, func() {
				buf.Reset()
				_, _ = WriteBool(buf, tc.b)
			}, 0)
		})
	}
}

func BenchmarkWriteBool(b *testing.B) {
	for _, tc := range writeBoolTestCases {
		b.Run(strconv.FormatBool(tc.b), func(b *testing.B) {
			buf := new(bytes.Buffer)
			for n := 0; n < b.N; n++ {
				buf.Reset()
				_, _ = WriteBool(buf, tc.b)
			}
		})
	}
}

var writeFloatTestCases = []struct {
	f        float64
	expected string
}{
	{
		f:        0,
		expected: "0",
	},
	{
		f:        1,
		expected: "1",
	},
	{
		f:        12.34,
		expected: "12.34",
	},
}

func TestWriteFloat(t *testing.T) {
	for _, tc := range writeFloatTestCases {
		t.Run(strconv.FormatFloat(tc.f, 'f', -1, 64), func(t *testing.T) {
			buf := new(bytes.Buffer)
			n, err := WriteFloat(buf, tc.f, 'f', -1, 64)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected), n)
			assert.Equal(t, tc.expected, buf.String())
			assert.AllocsPerRun(t, 100, func() {
				buf.Reset()
				_, _ = WriteFloat(buf, tc.f, 'f', -1, 64)
			}, 0)
		})
	}
}

func BenchmarkWriteFloat(b *testing.B) {
	for _, tc := range writeFloatTestCases {
		b.Run(strconv.FormatFloat(tc.f, 'f', -1, 64), func(b *testing.B) {
			buf := new(bytes.Buffer)
			for n := 0; n < b.N; n++ {
				buf.Reset()
				_, _ = WriteFloat(buf, tc.f, 'f', -1, 64)
			}
		})
	}
}

var writeIntTestCases = []struct {
	i        int64
	expected string
}{
	{
		i:        0,
		expected: "0",
	},
	{
		i:        1,
		expected: "1",
	},
	{
		i:        2,
		expected: "2",
	},

	{
		i:        1234567890,
		expected: "1234567890",
	},
	{
		i:        -1,
		expected: "-1",
	},
	{
		i:        -1234567890,
		expected: "-1234567890",
	},
}

func TestWriteInt(t *testing.T) {
	for _, tc := range writeIntTestCases {
		t.Run(strconv.FormatInt(tc.i, 10), func(t *testing.T) {
			buf := new(bytes.Buffer)
			n, err := WriteInt(buf, tc.i, 10)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected), n)
			assert.Equal(t, tc.expected, buf.String())
			assert.AllocsPerRun(t, 100, func() {
				buf.Reset()
				_, _ = WriteInt(buf, tc.i, 10)
			}, 0)
		})
	}
}

func BenchmarkWriteInt(b *testing.B) {
	for _, tc := range writeIntTestCases {
		b.Run(strconv.FormatInt(tc.i, 10), func(b *testing.B) {
			buf := new(bytes.Buffer)
			for n := 0; n < b.N; n++ {
				buf.Reset()
				_, _ = WriteInt(buf, tc.i, 10)
			}
		})
	}
}

var writeUintTestCases = []struct {
	i        uint64
	expected string
}{
	{
		i:        0,
		expected: "0",
	},
	{
		i:        1,
		expected: "1",
	},
	{
		i:        2,
		expected: "2",
	},

	{
		i:        1234567890,
		expected: "1234567890",
	},
}

func TestWriteUint(t *testing.T) {
	for _, tc := range writeUintTestCases {
		t.Run(strconv.FormatUint(tc.i, 10), func(t *testing.T) {
			buf := new(bytes.Buffer)
			n, err := WriteUint(buf, tc.i, 10)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected), n)
			assert.Equal(t, tc.expected, buf.String())
			assert.AllocsPerRun(t, 100, func() {
				buf.Reset()
				_, _ = WriteUint(buf, tc.i, 10)
			}, 0)
		})
	}
}

func BenchmarkWriteUint(b *testing.B) {
	for _, tc := range writeUintTestCases {
		b.Run(strconv.FormatUint(tc.i, 10), func(b *testing.B) {
			buf := new(bytes.Buffer)
			for n := 0; n < b.N; n++ {
				buf.Reset()
				_, _ = WriteUint(buf, tc.i, 10)
			}
		})
	}
}

var writeQuoteTestCases = []struct {
	name     string
	s        string
	expected string
}{
	{
		name:     "Empty",
		s:        "",
		expected: `""`,
	},
	{
		name:     "Simple",
		s:        "test",
		expected: `"test"`,
	},
	{
		name:     "Quote",
		s:        "\"",
		expected: `"\""`,
	},
}

func TestWriteQuote(t *testing.T) {
	for _, tc := range writeQuoteTestCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			n, err := WriteQuote(buf, tc.s)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.expected), n)
			assert.Equal(t, tc.expected, buf.String())
			assert.AllocsPerRun(t, 100, func() {
				buf.Reset()
				_, _ = WriteQuote(buf, tc.s)
			}, 0)
		})
	}
}

func BenchmarkWriteQuote(b *testing.B) {
	for _, tc := range writeQuoteTestCases {
		b.Run(tc.name, func(b *testing.B) {
			buf := new(bytes.Buffer)
			for n := 0; n < b.N; n++ {
				buf.Reset()
				_, _ = WriteQuote(buf, tc.s)
			}
		})
	}
}
