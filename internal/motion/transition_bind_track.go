package motion

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// buildTracks turns the resolved operations and the captured paints into the
// tracks the renderer samples. The captured destination is the paint the
// first bind read, and the endpoints keep the operation's current value for
// every field the layout changed since the capture.
func buildTracks(ops []*layout.DisplayOp, initial, final []Paint,
	record transitionRecord, property string) []Track {
	tracks := make([]Track, len(ops))
	discrete := discretePaintProperty(property)
	for i := range tracks {
		to := Read(ops[i])
		if final != nil {
			to = final[i]
		}
		from, dest := endpoints(ops[i], initial[i], to)
		tracks[i] = Track{Op: ops[i], From: from, To: dest,
			Duration: millis(record.DurationMS), Delay: millis(record.DelayMS),
			Easing:   record.TimingFunction,
			Discrete: discrete && record.Behavior == "allow-discrete"}
		if discrete && record.Behavior != "allow-discrete" {
			tracks[i].Duration, tracks[i].Delay = 0, 0
		}
	}

	return tracks
}

func millis(value float64) time.Duration { return time.Duration(value * float64(time.Millisecond)) }
