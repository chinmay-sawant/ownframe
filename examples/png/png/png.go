// Package png is the on-demand PNG example.
// The page carries a stamp. PNG encodes the last draw once and caches the
// bytes until the next Redraw. Redraw changes the stamp and Save PNG writes
// a.png. ownframe opens the window. This package does not.
package png

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed png.html
var pngHTML string

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 560
	DefaultHeight = 480
	// savedPNGPerm keeps saved pictures owner-only.
	savedPNGPerm = 0o600
)

// View is the stamp and the status line the template prints.
type View struct {
	Stamp  int
	Status string
}

// App is the PNG screen.
type App struct {
	page *ownframe.Page
	view View
}
