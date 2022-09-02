package errors

import (
	"fmt"
	"io"
	"strconv"
)

// WithTag adds a tag to an error.
//
// It should be used for short and simple values, such as identifiers.
func WithTag(err error, key, value string) error {
	if err == nil {
		return nil
	}
	return &tag{
		err:   err,
		key:   key,
		value: value,
	}
}

// WithTagInt is a helper for WithTag with int value.
func WithTagInt(err error, key string, value int) error {
	return WithTag(err, key, strconv.Itoa(value))
}

// WithTagInt64 is a helper for WithTag with int4 value.
func WithTagInt64(err error, key string, value int64) error {
	return WithTag(err, key, strconv.FormatInt(value, 10))
}

// WithTagFloat64 is a helper for WithTag with float64 value.
func WithTagFloat64(err error, key string, value float64) error {
	return WithTag(err, key, strconv.FormatFloat(value, 'g', -1, 64))
}

// WithTagBool is a helper for WithTag with bool value.
func WithTagBool(err error, key string, value bool) error {
	return WithTag(err, key, strconv.FormatBool(value))
}

type tag struct {
	err   error
	key   string
	value string
}

func (err *tag) Tag() (key string, value string) {
	return err.key, err.value
}

func (err *tag) WriteErrorMessage(w io.Writer, verbose bool) bool {
	if !verbose {
		return false
	}
	_, _ = fmt.Fprintf(w, "tag %s = %s", err.key, err.value)
	return true
}

func (err *tag) Error() string                 { return Error(err) }
func (err *tag) Format(s fmt.State, verb rune) { Format(err, s, verb) }
func (err *tag) Unwrap() error                 { return err.err }

// Tags returns the tags associated to an error.
func Tags(err error) map[string]string {
	tags := make(map[string]string)
	for ; err != nil; err = Unwrap(err) {
		err, ok := err.(*tag) //nolint:errorlint // We want to compare the current error.
		if !ok {
			continue
		}
		k, v := err.Tag()
		_, ok = tags[k]
		if ok {
			continue
		}
		tags[k] = v
	}
	return tags
}
