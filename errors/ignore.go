package errors

import (
	"fmt"
	"io"
)

// Ignore flags an error as ignored.
func Ignore(err error) error {
	if err == nil {
		return nil
	}
	return &ignore{
		err: err,
	}
}

type ignore struct {
	err error
}

func (err *ignore) Ignored() bool {
	return true
}

func (err *ignore) WriteErrorMessage(w io.Writer, verbose bool) bool {
	_, _ = io.WriteString(w, "ignored")
	return true
}

func (err *ignore) Error() string                 { return Error(err) }
func (err *ignore) Format(s fmt.State, verb rune) { Format(err, s, verb) }
func (err *ignore) Unwrap() error                 { return err.err }

// IsIgnored returns true if an error is ignored.
func IsIgnored(err error) bool {
	var werr *ignore
	ok := As(err, &werr)
	if ok {
		return werr.Ignored()
	}
	return false
}
