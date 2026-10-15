package web

import (
	"errors"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// ErrNilContext means Serve got a nil context.
var ErrNilContext = fetch.ErrNilContext

// ErrNilApp means Serve got a nil screen.
var ErrNilApp = errors.New("ownframe: nil screen")
