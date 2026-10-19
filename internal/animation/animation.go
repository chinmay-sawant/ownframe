// Package animation samples CSS-style keyframes onto retained display ops.
package animation

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Easing maps progress within one keyframe interval to eased progress.
type Easing func(float64) float64

// Target pairs a display operation with its captured starting paint.
type Target struct {
	Op   *ownframe.DisplayOp
	From cssmotion.Paint
}

// Keyframe stores values declared at Offset. X and Y are display-list points.
type Keyframe struct {
	Offset                     float64
	X, Y                       float64
	Opacity                    float64
	R, G, B                    float64
	Move, OpacitySet, ColorSet bool
}

// Direction controls the order in which iterations play.
type Direction uint8

const (
	Normal Direction = iota
	Reverse
	Alternate
	AlternateReverse
)

// FillMode controls paint before and after the active interval.
type FillMode uint8

const (
	FillNone FillMode = iota
	FillForwards
	FillBackwards
	FillBoth
)

// PlayState controls whether Step applies the keyframes.
type PlayState uint8

const (
	Running PlayState = iota
	Paused
)

// Timing describes the animation clock and its iteration behavior.
type Timing struct {
	Duration, Delay, EndDelay time.Duration
	Iterations                float64
	Infinite                  bool
	Direction                 Direction
	Fill                      FillMode
	Play                      PlayState
	TimingFunction            string
	Easing                    Easing
}

// Player applies one animation to its targets. Construct with New, call Start
// when playback begins, then call Step before drawing each frame.
type Player struct {
	targets []Target
	frames  []Keyframe
	timing  Timing
	start   time.Time
	started bool
	active  bool
}
