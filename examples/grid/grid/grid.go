// Package grid is a small CSS grid example.
package grid

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed grid.html
var gridHTML string

const (
	DefaultWidth  = 480
	DefaultHeight = 720
)

// Tile is one cell in the grid.
type Tile struct {
	ID    string
	Label string
	Class string
	On    bool
}

// View is the data the template reads.
type View struct {
	Title string
	Tiles []Tile
}

// App is the grid screen.
type App struct {
	page *ownframe.Page
	view View
}
