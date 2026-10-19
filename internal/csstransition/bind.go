package csstransition

import (
	"errors"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

// Bind builds a transition renderer for one CSS property on a box.
func Bind(d *ownframe.Display, box ownframe.Box, property string,
	indices []int, initial []cssmotion.Paint, start time.Time) (*transition.Renderer, error) {
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
	tracks := make([]transition.Track, len(ops))
	for i := range tracks {
		tracks[i] = transition.Track{Op: ops[i], From: initial[i], To: cssmotion.Read(ops[i]),
			Duration: millis(record.DurationMS), Delay: millis(record.DelayMS), Easing: record.TimingFunction}
	}
	renderer, err := transition.New(tracks)
	if err != nil {
		return nil, err
	}
	renderer.Start(start)
	return renderer, nil
}

func millis(value float64) time.Duration { return time.Duration(value * float64(time.Millisecond)) }
