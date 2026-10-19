package csspreview

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

type pair struct {
	op       *ownframe.DisplayOp
	from, to paint
}

type motion struct {
	pairs                               []pair
	frames                              []keyframe
	start                               time.Time
	duration, delay, endDelay           time.Duration
	iterations                          float64
	infinite                            bool
	kind, direction, fill, play, easing string
	pointsPerPixel                      float64
}

func (m *motion) step(now time.Time) bool {
	if m.duration <= 0 || m.play == "paused" {
		return false
	}
	progress, active, apply := m.position(now)
	for _, q := range m.pairs {
		v := q.from
		if apply {
			if m.kind == "animation" && len(m.frames) > 0 {
				v = sampleFrames(m.frames, progress, q.from, q.op.Kind == ownframe.DisplayOpFillRect, m.easing)
				v.x += keyframeX(m.frames, progress, m.easing) * m.pointsPerPixel
				v.y += keyframeY(m.frames, progress, m.easing) * m.pointsPerPixel
			} else {
				v = blend(q.from, q.to, easing(m.easing, progress))
			}
		}
		write(q.op, v)
	}
	return active
}
