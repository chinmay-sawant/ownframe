package pagetest

import "errors"

// ErrNoBox means no hit box matched the id or action.
var ErrNoBox = errors.New("pagetest: no box")
