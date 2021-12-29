package errors

import (
	"fmt"
	"io"
	"regexp"
	"testing"
)

func TestStack(t *testing.T) {
	err := newBase("error")
	err = WithStack(err)
	sfs := StackFrames(err)
	if len(sfs) != 1 {
		t.Fatalf("unexpected length: got %d, want %d", len(sfs), 1)
	}
	sf := sfs[0]
	if sf == nil {
		t.Fatal("no stack frames")
	}
	f, _ := sf.Next()
	expectedFunction := "github.com/pierrre/go-libs/errors.TestStack"
	if f.Function != expectedFunction {
		t.Fatalf("unexpected function: got %q, want %q", f.Function, expectedFunction)
	}
}

func TestStackNil(t *testing.T) {
	err := WithStack(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStackError(t *testing.T) {
	err := newBase("error")
	err = WithStack(err)
	s := err.Error()
	expected := "error"
	if s != expected {
		t.Fatalf("unexpected message: got %q, want %q", s, expected)
	}
}

func TestStackFormat(t *testing.T) {
	err := newBase("error")
	err = WithStack(err)
	s := fmt.Sprintf("%+v", err)
	expectedRegexp := regexp.MustCompile(`^stack(\n\t.+ .+:\d+)+\nerror$`)
	if !expectedRegexp.MatchString(s) {
		t.Fatalf("unexpected formatted message:\ngot: %q\nwant match: %q", s, expectedRegexp)
	}
}

func BenchmarkStackFormat(b *testing.B) {
	err := newBase("error")
	err = WithStack(err)
	for i := 0; i < b.N; i++ {
		_, _ = fmt.Fprintf(io.Discard, "%+v", err)
	}
}
