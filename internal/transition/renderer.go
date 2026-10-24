// Package transition samples timed changes into OwnFrame display operations.
package transition

import (
	"errors"
	"fmt"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Track describes one display operation's transition.
type Track struct {
	Op              *layout.DisplayOp
	From, To        cssmotion.Paint
	Duration, Delay time.Duration
	Easing          string
	// Discrete switches endpoints at 50% of eased progress.
	Discrete bool
}

// Renderer samples tracks from their start time. Track operations must remain
// valid until Step reports that every track has finished.
type Renderer struct {
	tracks  []Track
	start   time.Time
	started bool
}

// New validates and copies the tracks. Negative delays are allowed.
func New(tracks []Track) (*Renderer, error) {
	if len(tracks) == 0 {
		return nil, errors.New("transition: at least one track is required")
	}
	copyOfTracks := append([]Track(nil), tracks...)
	for i, track := range copyOfTracks {
		if track.Op == nil {
			return nil, fmt.Errorf("transition: track %d has no display operation", i)
		}
		if track.Duration < 0 {
			return nil, fmt.Errorf("transition: track %d has negative duration", i)
		}
	}
	return &Renderer{tracks: copyOfTracks}, nil
}

// Start sets the shared start time. Calling it again restarts every track.
func (r *Renderer) Start(now time.Time) {
	if r == nil {
		return
	}
	r.start, r.started = now, true
}

// Retarget points every track at the operation at the same position in ops.
// A display commit rebuilds the operation list, so a renderer that keeps its
// clock and endpoints must write through the new list. It refuses a different
// track count, because a renderer built for one target set must not half-write
// another.
func (r *Renderer) Retarget(ops []*layout.DisplayOp) error {
	if r == nil {
		return errors.New("transition: renderer is nil")
	}
	if len(ops) != len(r.tracks) {
		return fmt.Errorf("transition: retarget has %d operations, want %d",
			len(ops), len(r.tracks))
	}
	for i := range ops {
		if ops[i] == nil {
			return fmt.Errorf("transition: retarget operation %d is nil", i)
		}
		r.tracks[i].Op = ops[i]
	}
	return nil
}
