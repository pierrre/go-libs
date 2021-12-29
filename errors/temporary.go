package errors

import (
	"fmt"
	"io"
)

// WithTemporary flags an errors as temporary.
func WithTemporary(err error, tmp bool) error {
	if err == nil {
		return nil
	}
	return &temporary{
		err: err,
		tmp: tmp,
	}
}

type temporary struct {
	err error
	tmp bool
}

func (err *temporary) Temporary() bool {
	return err.tmp
}

func (err *temporary) WriteErrorMessage(w io.Writer, verbose bool) bool {
	_, _ = fmt.Fprintf(w, "temporary %t", err.tmp)
	return true
}

func (err *temporary) Error() string                 { return Error(err) }
func (err *temporary) Format(s fmt.State, verb rune) { Format(err, s, verb) }
func (err *temporary) Unwrap() error                 { return err.err }

// IsTemporary returns true if an error is temporary, false otherwise.
// By default, an error is temporary.
func IsTemporary(err error) bool {
	var werr *temporary
	ok := As(err, &werr)
	if ok {
		return werr.Temporary()
	}
	return true
}
