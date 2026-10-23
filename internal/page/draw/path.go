package draw

import (
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Path is how a display list is published.
type Path uint8

const (
	PathNone Path = iota
	PathList
	PathBitmap
)

// Choose picks the display-list or bitmap path.
func Choose(d *layout.Display) Path {
	if d == nil {
		return PathNone
	}
	if render.Replayable(d) {
		return PathList
	}

	return PathBitmap
}
