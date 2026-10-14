package runtimebehavior

import "errors"

// IsPromiseErrorValue reports whether err is a PromiseError.
func IsPromiseErrorValue(err error) bool {
	return IsPromiseError(err)
}

// Unwrap returns the stored error text as an error.
func (p PromiseError) Unwrap() error {
	if p.Err == "" {
		return nil
	}
	return &promiseTextError{text: p.Err}
}

type promiseTextError struct{ text string }

func (e *promiseTextError) Error() string { return e.text }

// Is reports target matching by string.
func (p PromiseError) Is(target error) bool {
	return errors.Is(&promiseTextError{text: p.Err}, target)
}
