// Package login is the sign-in example.
// The screen is an HTML template. Clicks and keystrokes are Go functions.
// ownframe opens the window. This package does not.
package login

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed login.html
var loginHTML string

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 480
	DefaultHeight = 640

	// MinWidth and MinHeight are the smallest frame the screen will draw.
	MinWidth  = 320
	MinHeight = 400

	undoLimit = 64
)

// View is the login screen data.
// The template prints Error and Status, and grays the button until Ready.
// Email and Password bind through data-bind.
type View struct {
	Error    string
	Status   string
	Ready    bool
	Email    string
	Password ownframe.Secret
}

// App is the sign-in screen.
type App struct {
	page *ownframe.Page
	view View
	undo []snap
	redo []snap
}
