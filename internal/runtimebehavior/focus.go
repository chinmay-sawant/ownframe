package runtimebehavior

import "errors"

// FocusRestorer remembers focus across partial DOM updates.
type FocusRestorer struct {
	ElementID string
	Selection [2]int
	Caret     int
	Restored  bool
}

// Save records current focus state.
func (f *FocusRestorer) Save(id string, caret int, sel [2]int) {
	if f == nil {
		return
	}
	f.ElementID = id
	f.Caret = caret
	f.Selection = sel
	f.Restored = false
}

// Restore applies saved focus state to target if IDs match.
func (f *FocusRestorer) Restore(id string) bool {
	if f == nil || f.ElementID != id {
		return false
	}
	f.Restored = true
	return true
}

// Clear drops saved focus state.
func (f *FocusRestorer) Clear() {
	if f == nil {
		return
	}
	f.ElementID = ""
	f.Caret = 0
	f.Selection = [2]int{}
	f.Restored = false
}

// FocusState returns saved state for callers to apply.
func (f *FocusRestorer) FocusState() (string, int, [2]int, bool) {
	if f == nil {
		return "", 0, [2]int{}, false
	}
	return f.ElementID, f.Caret, f.Selection, f.Restored
}

// PromiseError surfaces an unhandled state error as a page error.
type PromiseError struct {
	Source string
	Err    string
}

// Error implements error.
func (p PromiseError) Error() string {
	if p.Source == "" {
		return p.Err
	}
	return p.Source + ": " + p.Err
}

// Wrap builds a PromiseError from an error.
func WrapPromiseError(source string, err error) PromiseError {
	if err == nil {
		return PromiseError{Source: source}
	}
	return PromiseError{Source: source, Err: err.Error()}
}

// IsPromiseError reports whether err is a PromiseError.
func IsPromiseError(err error) bool {
	var p PromiseError
	return errors.As(err, &p)
}
