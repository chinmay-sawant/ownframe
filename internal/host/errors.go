package host

import "errors"

// ErrNilContext means an API entry point received a nil context.
var ErrNilContext = errors.New("ownframe: nil context")
