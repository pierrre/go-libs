package testutil

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/pierrre/compare"
)

// Compare is an alias for CompareFatal.
func Compare(tb testing.TB, msg string, got, want any) {
	tb.Helper()
	CompareFatal(tb, msg, got, want)
}

// CompareFatal compares 2 values and fails immediately the test if there is a difference.
func CompareFatal(tb testing.TB, msg string, got, want any) {
	tb.Helper()
	comparef(tb, msg, got, want, tb.Fatalf)
}

// CompareError compares 2 values and flags the test as errored if there is a difference.
// It is useful for tests that run in a separate goroutine.
// You can get the "failed" status of the current test with tb.Failed().
func CompareError(tb testing.TB, msg string, got, want any) {
	tb.Helper()
	comparef(tb, msg, got, want, tb.Errorf)
}

func comparef(tb testing.TB, msg string, got, want any, f func(format string, args ...any)) {
	tb.Helper()
	if tb.Failed() {
		return
	}
	diff := compare.Compare(got, want)
	if len(diff) != 0 {
		f("%s:\ngot:\n%s\nwant:\n%s\ndiff:\n%+v", msg, spew.Sdump(got), spew.Sdump(want), diff)
	}
}
