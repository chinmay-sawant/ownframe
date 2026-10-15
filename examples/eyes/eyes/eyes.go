// Package eyes shows two cute robot eyes on one HTML page.
// The frame tick moves the retained display ops, so the eyes blink,
// look around, and cycle moods without another Redraw.
package eyes

import (
	"math/rand/v2"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	DefaultWidth  = 960
	DefaultHeight = 540
	MinWidth      = 480
	MinHeight     = 320
	MaxWidth      = 1920
	MaxHeight     = 1080
)

// App is the eyes screen.
type App struct {
	page *ownframe.Page
	rng  *rand.Rand
	now  func() time.Time
	last time.Time
	st   state
	pm   parts
	bnd  uint64
	tch  bool
}
