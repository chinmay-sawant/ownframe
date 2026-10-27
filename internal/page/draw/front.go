package draw

import "github.com/chinmay-sawant/ownframe/internal/host"

// Fronter is the front-half surface every sink needs.
type Fronter = host.Fronter

// Front is one display-else-image snapshot.
type Front = host.Front

// SelectFront prefers the display list, then the picture bounds, then the
// frame size. It calls only Display, Image, Size, and Generation, so the
// window and the browser share one preference and keep only sink delivery.
func SelectFront(s Fronter) Front { return host.SelectFront(s) }
