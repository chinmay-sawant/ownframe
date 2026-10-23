package page

import "github.com/chinmay-sawant/ownframe/internal/page/draw"

// Fronter is the front-half surface every sink needs.
type Fronter = draw.Fronter

// Front is one display-else-image snapshot.
type Front = draw.Front

// SelectFront prefers the display list, then the picture bounds, then the
// frame size. It calls only Display, Image, Size, and Generation, so the
// window and the browser share one preference and keep only sink delivery.
func SelectFront(s Fronter) Front { return draw.SelectFront(s) }
