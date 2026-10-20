package csstransition

import (
	"errors"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

// BindTo builds a transition renderer with an explicit destination paint.
// A nil final reads the destination from the display, which is what a first
// bind wants. A caller that rebinds later passes the paints it captured on
// the first bind, so a redraw in the middle of the transition does not make
// the destination the value the transition had reached.
func BindTo(d *ownframe.Display, box ownframe.Box, property string,
	indices []int, initial, final []cssmotion.Paint, start time.Time) (*transition.Renderer, error) {
	if d == nil {
		return nil, errors.New("csstransition: display is nil")
	}
	index := -1
	for i := range box.Transitions {
		item := box.Transitions[i]
		if item.Property == property || item.Property == "all" {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, errors.New("csstransition: box has no matching transition metadata")
	}
	record := box.Transitions[index]
	ops := cssmotion.Resolve(d, indices)
	if len(ops) != len(indices) {
		return nil, errors.New("csstransition: captured operations are missing from display")
	}
	if len(ops) == 0 || len(initial) != len(ops) {
		return nil, errors.New("csstransition: target operation paints do not match")
	}
	if final != nil && len(final) != len(ops) {
		return nil, errors.New("csstransition: destination paints do not match")
	}
	tracks := make([]transition.Track, len(ops))
	discrete := discretePaintProperty(property)
	for i := range tracks {
		to := cssmotion.Read(ops[i])
		if final != nil {
			to = final[i]
		}
		tracks[i] = transition.Track{Op: ops[i], From: initial[i], To: to,
			Duration: millis(record.DurationMS), Delay: millis(record.DelayMS), Easing: record.TimingFunction,
			Discrete: discrete && record.Behavior == "allow-discrete"}
		if discrete && record.Behavior != "allow-discrete" {
			tracks[i].Duration, tracks[i].Delay = 0, 0
		}
	}
	renderer, err := transition.New(tracks)
	if err != nil {
		return nil, err
	}
	renderer.Start(start)
	return renderer, nil
}

func millis(value float64) time.Duration { return time.Duration(value * float64(time.Millisecond)) }
