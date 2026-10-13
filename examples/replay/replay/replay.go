// Package replay is the display-list replay example.
// The replay page is drawn from its display list. The fallback page has the
// same content inside a non-separable mix-blend-mode group, so ownframe keeps
// a bitmap and the window shows the fallback badge. data-action buttons
// switch routes.
// ownframe opens the window. This package does not.
package replay

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed replay.html
var replayHTML string

//go:embed fallback.html
var fallbackHTML string

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 700
	DefaultHeight = 600
)

// View is the status line the templates print.
type View struct {
	Status string
}

// ShowingReplay records the replay mode and returns the status text.
func (v *View) ShowingReplay() string {
	v.Status = "display replay"

	return v.Status
}

// ShowingFallback records the fallback mode and returns the status text.
func (v *View) ShowingFallback() string {
	v.Status = "bitmap fallback"

	return v.Status
}

// App is the replay screen.
type App struct {
	page *ownframe.Page
	view View
}
