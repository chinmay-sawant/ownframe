// Package motion drives the CSS keyframe animations a page declares. A page
// binds the runtime after it publishes a display list and steps it once per
// frame; the runtime resolves one player per animated element, samples the
// players, and reports the rectangles the page must repaint.
//
// The package does not import internal/page, internal/window, or
// internal/replay. The page owns the display list, and motion only edits the
// operations it was handed.
package motion

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Display is the retained display list a runtime samples. It is the same
// layout type the page publishes, so a caller passes its list unchanged.
type Display = layout.Display

// Runtime binds the keyframe animations of one page and samples them.
// It is not safe for concurrent use: the page loop owns it.
type Runtime struct {
	// display is the list the entries were resolved against. A publish
	// installs a new list pointer, so a mismatch in Step means the page
	// replaced the list without rebinding.
	display *Display

	entries []entry
	stats   Stats

	// restamp is the per-element re-projection requests the last Step
	// collected. TakeRestamp hands them to the page and clears them, so one
	// tick applies one pass.
	restamp []Request

	// now is the clock a newly bound effect starts at. Tests replace it.
	now func() time.Time
}

// New returns an idle runtime with no players bound.
func New() *Runtime {
	return &Runtime{now: time.Now}
}

// Active reports whether the runtime needs another frame: a bound player has
// not been sampled yet, or its latest sample left work to do. A finished,
// paused, or triggered player is not active, so the page stops asking for
// frames.
func (r *Runtime) Active() bool {
	if r == nil {
		return false
	}

	for i := range r.entries {
		if !r.entries[i].sampled && !r.entries[i].idle {
			return true
		}

		if r.entries[i].player.Active() {
			return true
		}
	}

	return false
}

// Stats returns a snapshot of the runtime counters.
func (r *Runtime) Stats() Stats {
	if r == nil {
		return Stats{}
	}

	return r.stats
}

// clock returns the time a newly bound effect starts at.
func (r *Runtime) clock() time.Time {
	if r.now == nil {
		return time.Now()
	}

	return r.now()
}
