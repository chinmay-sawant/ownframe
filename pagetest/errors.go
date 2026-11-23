package pagetest

import (
	"errors"

	"github.com/chinmay-sawant/ownframe"
)

// ErrNoBox means no hit box matched the id or action.
var ErrNoBox = errors.New("pagetest: no box")

// ErrNilPage means a helper received a nil page.
var ErrNilPage = ownframe.ErrNilPage
