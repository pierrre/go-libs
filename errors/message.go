package errors

import (
	"fmt"
	"io"
)

// WithMessage adds a message to an error.
func WithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	if msg == "" {
		return err
	}
	return &message{
		err: err,
		msg: msg,
	}
}

// WithMessagef adds a formatted message to an error.
func WithMessagef(err error, format string, args ...interface{}) error {
	return WithMessage(err, fmt.Sprintf(format, args...))
}

type message struct {
	err error
	msg string
}

func (err *message) WriteErrorMessage(w io.Writer, verbose bool) bool {
	_, _ = io.WriteString(w, err.msg)
	return true
}

func (err *message) Error() string                 { return Error(err) }
func (err *message) Format(s fmt.State, verb rune) { Format(err, s, verb) }
func (err *message) Unwrap() error                 { return err.err }

// Wrap adds a message to an error, and optionnally a stack if it doesn't have one.
func Wrap(err error, msg string) error {
	err = WithMessage(err, msg)
	err = ensureStack(err, 2)
	return err
}

// Wrapf adds a formatted message to an error, and optionnally a stack if it doesn't have one.
func Wrapf(err error, format string, args ...interface{}) error {
	err = WithMessagef(err, format, args...)
	err = ensureStack(err, 2)
	return err
}
