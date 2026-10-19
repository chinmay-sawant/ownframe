package csspreview

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func bind(d *ownframe.Display, b ownframe.Box, t target, base []paint, origin time.Time) motion {
	if d == nil {
		return motion{}
	}
	s := timing(d, b, t)
	m := motion{
		kind: t.kind, duration: time.Duration(s.duration * float64(time.Millisecond)),
		delay:      time.Duration(s.delay * float64(time.Millisecond)),
		endDelay:   time.Duration(s.endDelay * float64(time.Millisecond)),
		iterations: s.iterations, infinite: s.infinite,
		direction: s.direction, fill: s.fill, play: s.play, easing: s.easing,
	}
	if m.duration <= 0 {
		return m
	}
	m.start = origin
	m.pointsPerPixel = d.PointsPerPixel
	if t.kind == "animation" {
		m.frames = namedKeyframes(d.Keyframes, s.name)
	}
	index := 0
	for i := range d.Ops {
		if !inside(d, b, d.Ops[i]) {
			continue
		}
		if index >= len(base) {
			break
		}
		before := base[index]
		index++
		after := read(&d.Ops[i])
		m.pairs = append(m.pairs, pair{&d.Ops[i], before, after})
	}
	return m
}
