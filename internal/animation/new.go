package animation

import (
	"fmt"
	"math"
	"sort"
)

// New validates and copies the targets and keyframes for an animation.
func New(targets []Target, frames []Keyframe, timing Timing) (*Player, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("animation needs at least one target")
	}
	for i, target := range targets {
		if target.Op == nil {
			return nil, fmt.Errorf("animation target %d has no operation", i)
		}
	}
	frames = append([]Keyframe(nil), frames...)
	for i, frame := range frames {
		if math.IsNaN(frame.Offset) || frame.Offset < 0 || frame.Offset > 1 {
			return nil, fmt.Errorf("keyframe %d offset must be between 0 and 1", i)
		}
	}
	sort.SliceStable(frames, func(i, j int) bool {
		return frames[i].Offset < frames[j].Offset
	})
	targets = append([]Target(nil), targets...)
	return &Player{targets: targets, frames: frames, timing: timing}, nil
}
