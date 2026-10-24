package transition

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// easeMidpoint is the eased progress where a discrete track flips.
const easeMidpoint = 0.5

// ProgressFor returns the eased progress of one transition record at now and
// whether the run is still active. It is the clock Step and Progress share,
// for callers that carry a record rather than a renderer: a geometry or
// discrete transition has no display operation to bind, but it follows the
// same duration, delay, and timing function.
//
// A run inside a positive delay reports zero progress and stays active. A
// finished or zero-duration run reports one and inactive, so the caller can
// install the destination once and drop the entry.
func ProgressFor(record layout.Transition, start, now time.Time) (float64, bool) {
	if record.DurationMS <= 0 {
		return 1, false
	}

	elapsed := float64(now.Sub(start)) - record.DelayMS*float64(time.Millisecond)
	duration := record.DurationMS * float64(time.Millisecond)
	if elapsed < 0 {
		return 0, true
	}
	if elapsed >= duration {
		return 1, false
	}

	return easing(record.TimingFunction, elapsed/duration), true
}

// Progress returns the timing-function output of the run at now, on the same
// clock Step samples. A run inside a positive delay reports zero, and a
// finished or zero-duration run reports one. A normal bind shares one record
// across every track, so the first track with a duration speaks for the run.
//
// The value is the eased progress, not the raw time fraction, because CSS
// Transitions defines the reversing shortening factor as the portion of the
// value range a run has crossed. A caller that reverses a run uses this value
// to shorten the reverse by that portion.
func (r *Renderer) Progress(now time.Time) float64 {
	if r == nil || !r.started {
		return 0
	}
	for _, track := range r.tracks {
		if track.Duration <= 0 {
			continue
		}
		progress := float64(now.Sub(r.start)-track.Delay) / float64(track.Duration)
		if progress <= 0 {
			return 0
		}
		if progress >= 1 {
			return 1
		}
		return easing(track.Easing, progress)
	}

	return 1
}

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
		if track.Discrete {
			if easing(track.Easing, progress) < easeMidpoint {
				cssmotion.Write(track.Op, track.From)
			} else {
				cssmotion.Write(track.Op, track.To)
			}
			active = true
			continue
		}
		value := cssmotion.Blend(track.From, track.To, easing(track.Easing, progress))
		cssmotion.Write(track.Op, value)
		active = true
	}
	return active
}
