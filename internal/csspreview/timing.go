package csspreview

import "github.com/chinmay-sawant/ownframe"

type schedule struct {
	duration, delay, endDelay, iterations float64
	infinite                              bool
	direction, fill, play, easing, name   string
}

func timing(d *ownframe.Display, b ownframe.Box, t target) schedule {
	if d == nil {
		return schedule{}
	}
	if t.kind == "transition" {
		for _, r := range b.Transitions {
			if r.Property == t.property || r.Property == "all" {
				return schedule{duration: r.DurationMS, delay: r.DelayMS, iterations: 1, fill: "forwards", play: "running", easing: r.TimingFunction}
			}
		}
		return schedule{}
	}
	if len(b.Animations) == 0 {
		return schedule{}
	}
	a := b.Animations[0]
	return schedule{duration: a.DurationMS, delay: a.DelayStartMS, endDelay: a.DelayEndMS,
		iterations: a.IterationCount, infinite: a.IterationInfinite, direction: a.Direction,
		fill: a.FillMode, play: a.PlayState, easing: a.TimingFunction, name: a.Name}
}
