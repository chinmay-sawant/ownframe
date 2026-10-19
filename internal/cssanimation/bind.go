package cssanimation

import (
	"errors"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Bind builds an animation player from one box's Blinkless CSS metadata.
func Bind(d *ownframe.Display, box ownframe.Box, indices []int, initial []cssmotion.Paint,
	start time.Time) (*animation.Player, error) {
	if d == nil || len(box.Animations) == 0 {
		return nil, errors.New("cssanimation: box has no animation metadata")
	}
	a := box.Animations[0]
	ops := cssmotion.Resolve(d, indices)
	if len(ops) != len(indices) {
		return nil, errors.New("cssanimation: captured operations are missing from display")
	}
	if len(ops) == 0 || len(initial) != len(ops) {
		return nil, errors.New("cssanimation: target operation paints do not match")
	}
	targets := make([]animation.Target, len(ops))
	for i := range targets {
		targets[i] = animation.Target{Op: ops[i], From: initial[i]}
	}
	player, err := animation.New(targets, keyframes(d.Keyframes, a.Name, d.PointsPerPixel), timing(a))
	if err != nil {
		return nil, err
	}
	player.Start(start)
	return player, nil
}
