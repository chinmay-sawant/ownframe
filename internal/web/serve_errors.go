package web

import (
	"errors"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// ErrNilContext means Serve got a nil context.
var ErrNilContext = host.ErrNilContext

// ErrNilApp means Serve got a nil screen.
var ErrNilApp = errors.New("ownframe: nil screen")
