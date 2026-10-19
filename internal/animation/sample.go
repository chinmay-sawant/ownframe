package animation

import "github.com/chinmay-sawant/ownframe/internal/cssmotion"

type valueOf func(Keyframe) (float64, bool)
type stop struct{ at, value float64 }

func moveX(k Keyframe) (float64, bool)   { return k.X, k.Move }
func moveY(k Keyframe) (float64, bool)   { return k.Y, k.Move }
func opacity(k Keyframe) (float64, bool) { return k.Opacity, k.OpacitySet }
func red(k Keyframe) (float64, bool)     { return k.R, k.ColorSet }
func green(k Keyframe) (float64, bool)   { return k.G, k.ColorSet }
func blue(k Keyframe) (float64, bool)    { return k.B, k.ColorSet }

func (p *Player) sample(at, fallback float64, value valueOf) float64 {
	stops := make([]stop, 0, len(p.frames)+2)
	for _, frame := range p.frames {
		if v, ok := value(frame); ok {
			stops = append(stops, stop{frame.Offset, v})
		}
	}
	if len(stops) == 0 {
		return fallback
	}
	if stops[0].at > 0 {
		stops = append([]stop{{0, fallback}}, stops...)
	}
	if stops[len(stops)-1].at < 1 {
		stops = append(stops, stop{1, fallback})
	}
	if at <= stops[0].at {
		return stops[0].value
	}
	for i := 1; i < len(stops); i++ {
		if at > stops[i].at {
			continue
		}
		prev, next := stops[i-1], stops[i]
		span := next.at - prev.at
		if span <= 0 {
			return next.value
		}
		part := (at - prev.at) / span
		if p.timing.Easing != nil {
			part = p.timing.Easing(part)
		} else {
			part = cssmotion.Ease(p.timing.TimingFunction, part)
		}
		return prev.value + (next.value-prev.value)*part
	}
	return stops[len(stops)-1].value
}
