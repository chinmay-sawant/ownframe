package scroll

import (
	"math"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

const (
	sensDefault  = 1
	sensMin      = 0.25
	sensMax      = 3
	decelDefault = 5
	decelMin     = 1
	decelMax     = 12
)

// Touch is the finger-follow multiplier and the fling decay per second.
// Zero reads as the defaults, 1 and 5. SetTouch clamps a positive value
// into range and replaces a non-positive or non-finite value with the default.
type Touch struct {
	Sensitivity  float64
	Deceleration float64
}

// New is the scroll state of a new page: panning on, the configured thumb
// mode, and the default touch tuning. The zero State does not allow panning.
func New(bar host.ScrollbarMode) State {
	return State{
		allow: true,
		bar:   bar,
		touch: Touch{Sensitivity: sensDefault, Deceleration: decelDefault},
	}
}

// SetAllow turns window panning on or off.
func (s *State) SetAllow(on bool) { s.allow = on }

// Allow reports whether the window may pan this page.
func (s *State) Allow() bool { return s.allow }

// Bar is the configured thumb visibility mode.
func (s *State) Bar() host.ScrollbarMode { return s.bar }

// SetPinZ pins layers at or above z to the viewport. Zero disables pinning.
func (s *State) SetPinZ(z int) { s.pinZ = z }

// PinZ is the pinned layer threshold.
func (s *State) PinZ() int { return s.pinZ }

// SetTouch stores clamped touch tuning.
func (s *State) SetTouch(o Touch) {
	if o.Sensitivity <= 0 || badFloat(o.Sensitivity) {
		o.Sensitivity = sensDefault
	}
	if o.Deceleration <= 0 || badFloat(o.Deceleration) {
		o.Deceleration = decelDefault
	}
	s.touch = Touch{
		Sensitivity:  clamp(o.Sensitivity, sensMin, sensMax),
		Deceleration: clamp(o.Deceleration, decelMin, decelMax),
	}
}

// Sensitivity is the finger movement multiplier. A nil or unset State
// reports the default.
func (s *State) Sensitivity() float64 {
	if s == nil || s.touch.Sensitivity == 0 {
		return sensDefault
	}
	return s.touch.Sensitivity
}

// Deceleration is the fling decay per second. A nil or unset State
// reports the default.
func (s *State) Deceleration() float64 {
	if s == nil || s.touch.Deceleration == 0 {
		return decelDefault
	}
	return s.touch.Deceleration
}

func clamp(v, low, high float64) float64 { return max(low, min(v, high)) }

func badFloat(v float64) bool { return math.IsNaN(v) || math.IsInf(v, 0) }
