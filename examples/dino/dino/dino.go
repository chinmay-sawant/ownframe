// Package dino shows the Chrome-style dinosaur game on one HTML page.
// The page is a scene of absolute boxes; the frame tick moves the
// display-list operations that make the dinosaur, the obstacles, and the
// ground marks. A key handler never draws, so the game paints from the
// tick alone.
package dino

import (
	"math/rand/v2"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// DefaultWidth and DefaultHeight are the size of a newly opened window.
const (
	DefaultWidth  = 900
	DefaultHeight = 300

	// MinWidth and MinHeight are the smallest frame the screen will draw.
	MinWidth  = 540
	MinHeight = 180

	// MaxWidth and MaxHeight cap a resized frame.
	MaxWidth  = 1800
	MaxHeight = 600
)

// pcgStream xors the time seed into the PCG stream selector.
const pcgStream = 0x9e3779b97f4a7c15

// point is a scene position in CSS pixels.
type point struct {
	x, y float64
}

// App is the dinosaur game screen.
type App struct {
	page *ownframe.Page
	game game
	rng  *rand.Rand

	now  func() time.Time
	last time.Time
	fps  meter

	clouds  [2]point
	pebbles [6]float64

	parts parts
	bound uint64

	// tapped is a tap jump held until the rise ends; touch switches the
	// overlay hints to tap messages.
	tapped bool
	touch  bool

	// duckUntil ends a swipe duck; the view maps the scene onto a touch
	// page.
	duckUntil time.Time

	view  view
	viewW int
	viewH int
}
