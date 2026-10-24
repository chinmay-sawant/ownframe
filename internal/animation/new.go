package animation

import (
	"fmt"
	"math"
	"sort"
)

// New validates and copies the targets and keyframes for one animation.
func New(targets []Target, frames []Keyframe, timing Timing) (*Player, error) {
	return NewComposite(targets, []Effect{{Frames: frames, Timing: timing}})
}

// NewComposite builds one player that composes several effects per frame.
func NewComposite(targets []Target, effects []Effect) (*Player, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("animation needs at least one target")
	}
	if len(effects) == 0 {
		return nil, fmt.Errorf("animation needs at least one effect")
	}
	effects = append([]Effect(nil), effects...)
	for i, target := range targets {
		if target.Op == nil {
			return nil, fmt.Errorf("animation target %d has no operation", i)
		}
	}
	for i := range effects {
		frames := append([]Keyframe(nil), effects[i].Frames...)
		for j, frame := range frames {
			if math.IsNaN(frame.Offset) || frame.Offset < 0 || frame.Offset > 1 {
				return nil, fmt.Errorf("effect %d keyframe %d offset must be between 0 and 1", i, j)
			}
		}
		sort.SliceStable(frames, func(a, b int) bool { return frames[a].Offset < frames[b].Offset })
		for j := range frames {
			if frames[j].Err != nil {
				return nil, fmt.Errorf("effect %d keyframe %d: %w", i, j, frames[j].Err)
			}
		}
		effects[i].Frames = frames
	}
	targets = append([]Target(nil), targets...)
	states := make([]effectState, len(effects))
	for i := range effects {
		states[i].timing = effects[i].Timing
		states[i].rate = 1
	}
	return &Player{targets: targets, effects: effects, states: states}, nil
}
