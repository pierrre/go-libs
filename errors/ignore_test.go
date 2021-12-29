package errors

import (
	"fmt"
	"io"
	"testing"
)

func TestIgnore(t *testing.T) {
	err := newBase("error")
	err = Ignore(err)
	ignored := IsIgnored(err)
	if !ignored {
		t.Fatalf("unexpected ignored: got %t, want %t", ignored, true)
	}
}

func TestIgnoreNil(t *testing.T) {
	err := Ignore(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsIgnoredFalse(t *testing.T) {
	err := newBase("error")
	ignored := IsIgnored(err)
	if ignored {
		t.Fatalf("unexpected ignored: got %t, want %t", ignored, false)
	}
}

func TestIgnoreError(t *testing.T) {
	err := newBase("error")
	err = Ignore(err)
	s := err.Error()
	expected := "ignored: error"
	if s != expected {
		t.Fatalf("unexpected message: got %q, want %q", s, expected)
	}
}

func TestIgnoreFormat(t *testing.T) {
	err := newBase("error")
	err = Ignore(err)
	s := fmt.Sprint(err) //nolint:gocritic // We want to test the Format method.
	expected := "ignored: error"
	if s != expected {
		t.Fatalf("unexpected message: got %q, want %q", s, expected)
	}
}

func BenchmarkIgnoreFormat(b *testing.B) {
	err := newBase("error")
	err = Ignore(err)
	for i := 0; i < b.N; i++ {
		_, _ = fmt.Fprint(io.Discard, err)
	}
}
