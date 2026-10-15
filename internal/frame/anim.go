package frame

import (
	"image"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// Marker receives a CSS-pixel damage box. *page.Page satisfies it.
type Marker interface {
	MarkRect(r image.Rectangle)
}

// Damage is the union of the boxes one step changed: where an op was and
// where it landed. Mark Box before the next draw, or the old ink stays on
// screen. A fade changes no box, so Old and New both carry the op's box.
type Damage struct {
	Old, New image.Rectangle
}

// Box is the union of the old and new boxes.
func (d Damage) Box() image.Rectangle { return d.Old.Union(d.New) }

// Empty reports whether the step changed nothing.
func (d Damage) Empty() bool { return d.Old.Empty() && d.New.Empty() }

// Easing maps a 0..1 progress to 0..1. Ease(1) must be 1 or the animation
// would never finish. The easings here satisfy that.
type Easing func(t float64) float64

// Anim is one compositor animation over retained ops. Construct it with
// TranslateTo, FadeTo, ScaleTo, or RotateTo, drive it from the page tick,
// and drop it once Step reports inactive. Constructing a new animation over
// the same ops re-targets from the state on screen, so an interrupted
// animation does not jump.
//
// Animations over one op compose while they patch different fields:
// translate patches X and Y, fade patches Alpha, and scale and rotate patch
// the transform.
//
// An Anim holds pointers into one display. A Redraw builds a new display
// list, so an animation must not outlive one.
type Anim struct {
	d     *page.Display
	ops   []*page.DisplayOp
	kind  animKind
	from  []float64
	to    []float64
	base  []mat6
	start time.Time
	dur   time.Duration
	ease  Easing
	live  bool
}

// Active reports whether Step still has work to do.
func (a *Anim) Active() bool { return a != nil && a.live }
