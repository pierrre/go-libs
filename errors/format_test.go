package errors

import (
	"fmt"
	"io"
	"testing"
)

func TestFormat(t *testing.T) {
	err := newBase("error")
	err = withTestFormat(err)
	for _, tc := range []struct {
		ft       string
		expected string
	}{
		{
			ft:       "%v",
			expected: "test: error",
		},
		{
			ft:       "%+v",
			expected: "test verbose\nerror",
		},
		{
			ft:       "%s",
			expected: "test: error",
		},
		{
			ft:       "%q",
			expected: "\"test: error\"",
		},
	} {
		t.Run(tc.ft, func(t *testing.T) {
			s := fmt.Sprintf(tc.ft, err)
			if s != tc.expected {
				t.Fatalf("unexpected message: got %q, want %q", s, tc.expected)
			}
		})
	}
}

func BenchmarkFormat(b *testing.B) {
	for wrapCount := 1; wrapCount <= 64; wrapCount *= 2 {
		b.Run(fmt.Sprintf("WrapCount_%d", wrapCount), func(b *testing.B) {
			err := newBase("error")
			for i := 0; i < wrapCount; i++ {
				err = withTestFormat(err)
			}
			for _, ft := range []string{"%v", "%+v", "%q"} {
				b.Run(ft, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_, _ = fmt.Fprintf(io.Discard, ft, err)
					}
				})
			}
		})
	}
}

type testFormatError struct {
	err error
}

func withTestFormat(err error) error {
	if err == nil {
		return nil
	}
	return &testFormatError{
		err: err,
	}
}

func (err *testFormatError) WriteErrorMessage(w io.Writer, verbose bool) bool {
	if verbose {
		_, _ = io.WriteString(w, "test verbose")
	} else {
		_, _ = io.WriteString(w, "test")
	}
	return true
}

func (err *testFormatError) Error() string                 { return Error(err) }
func (err *testFormatError) Format(s fmt.State, verb rune) { Format(err, s, verb) }
func (err *testFormatError) Unwrap() error                 { return err.err }
