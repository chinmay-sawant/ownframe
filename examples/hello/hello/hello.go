// Package hello is the hello world example.
// A single template prints one message. ownframe opens the window.
package hello

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed hello.html
var helloHTML string

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 480
	DefaultHeight = 320
)

// View is the message the template prints.
type View struct {
	Message string
}

// App is the hello world screen.
type App struct {
	page *ownframe.Page
	view View
}
