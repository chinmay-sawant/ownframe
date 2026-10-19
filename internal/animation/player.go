package animation

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Start sets the animation's clock origin. Step starts it on its first call
// when Start was not called explicitly.
func (p *Player) Start(now time.Time) {
	p.start, p.started = now, true
	p.active = p.timing.Duration > 0 && p.timing.Play != Paused
}

// Step samples and writes the frame at now, then reports whether it is active.
func (p *Player) Step(now time.Time) bool {
	if p.timing.Duration <= 0 || p.timing.Play == Paused {
		p.active = false
		return false
	}
	if !p.started {
		p.Start(now)
	}
	progress, active, apply := p.position(now)
	for _, target := range p.targets {
		value := target.From
		if apply {
			value.X += p.sample(progress, 0, moveX)
			value.Y += p.sample(progress, 0, moveY)
			value.Opacity = p.sample(progress, value.Opacity, opacity)
			if target.Op.Kind == ownframe.DisplayOpFillRect {
				value.R = p.sample(progress, value.R, red)
				value.G = p.sample(progress, value.G, green)
				value.B = p.sample(progress, value.B, blue)
			}
		}
		cssmotion.Write(target.Op, value)
	}
	p.active = active
	return active
}

// Active reports whether the latest Step left work to do.
func (p *Player) Active() bool { return p != nil && p.active }
