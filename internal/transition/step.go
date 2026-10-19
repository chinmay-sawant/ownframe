package transition

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Step writes the current sample and reports whether any track remains active.
func (r *Renderer) Step(now time.Time) bool {
	if r == nil || !r.started {
		return false
	}
	active := false
	for _, track := range r.tracks {
		if track.Duration == 0 {
			cssmotion.Write(track.Op, track.To)
			continue
		}
		progress := float64(now.Sub(r.start)-track.Delay) / float64(track.Duration)
		if progress < 0 {
			cssmotion.Write(track.Op, track.From)
			active = true
			continue
		}
		if progress >= 1 {
			cssmotion.Write(track.Op, track.To)
			continue
		}
		value := cssmotion.Blend(track.From, track.To, easing(track.Easing, progress))
		cssmotion.Write(track.Op, value)
		active = true
	}
	return active
}
