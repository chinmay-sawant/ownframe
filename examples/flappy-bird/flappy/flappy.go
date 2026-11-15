// Package flappy is the Flappy Bird example. The scene is one HTML page of
// absolute boxes styled by CSS; the frame tick moves the display-list fills
// for the bird, the pipes, the clouds, and the ground marks. A key or a
// click flaps and never draws, so the tick alone paints the game.
package flappy

import (
	"math/rand/v2"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// DefaultWidth and DefaultHeight are the size of a newly opened window.
const (
	DefaultWidth  = 480
	DefaultHeight = 720

	// MinWidth and MinHeight are the smallest frame the screen will draw.
	MinWidth  = 480
	MinHeight = 720

	// MaxWidth and MaxHeight cap a resized frame.
	MaxWidth  = 960
	MaxHeight = 1440
)

// rect is one fill in CSS pixels.
type rect struct {
	x, y, w, h float64
}

// App is the Flappy Bird screen.
type App struct {
	page *ownframe.Page
	game game
	rng  *rand.Rand

	now  func() time.Time
	last time.Time

	clouds  [cloudMax]float64
	stripes [stripeMax]float64

	parts parts
	bound uint64

	// ox and oy are the stage origin in CSS pixels, taken from the scene
	// box at bind. The tick paints game coordinates that start at the
	// stage corner, so every placed fill adds the origin to land where
	// the layout put the element. A full-width window centers the stage
	// and the origin is half the leftover width; at the 480 window it is
	// zero and painting is unchanged.
	ox, oy float64
}
