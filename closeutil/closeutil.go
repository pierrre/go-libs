// Package closeutil provides close related utilities.
package closeutil

import "fmt"

// F represents a function that closes something.
type F func()

// Nil returns a new F that checks if the F referenced by the method pointer receiver is not nil, and if so calls it.
func (f *F) Nil() F {
	return func() {
		if *f != nil {
			(*f)()
		}
	}
}

// Err represents a function that closes something and returns an error.
type Err func() error

// Convert converts the Err to a WithOnErr.
func (f Err) Convert() WithOnErr {
	return func(oe OnErr) {
		err := f()
		if err != nil {
			oe(err)
		}
	}
}

// Nil returns a new Err that checks if the Err referenced by the method pointer receiver is not nil, and if so calls it.
func (f *Err) Nil() Err {
	return func() error {
		if *f != nil {
			return (*f)()
		}
		return nil
	}
}

// OnErr represents a function that is called when an error occurs.
type OnErr func(error)

// Wrap wraps the OnErr with a message.
func (f OnErr) Wrap(msg string) OnErr {
	return func(err error) {
		f(fmt.Errorf("%s: %w", msg, err))
	}
}

// WithOnErr represents a function that closes something and calls OnErr if an error occurs.
type WithOnErr func(OnErr)

// Wrap wraps the WithOnErr with a message.
func (f WithOnErr) Wrap(msg string) WithOnErr {
	return func(oe OnErr) {
		f(oe.Wrap(msg))
	}
}

// Nil returns a new WithOnErr that checks if the WithOnErr referenced by the method pointer receiver is not nil, and if so calls it.
func (f *WithOnErr) Nil() WithOnErr {
	return func(oe OnErr) {
		if *f != nil {
			(*f)(oe)
		}
	}
}
