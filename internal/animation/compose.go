package animation

import (
	"math"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func compose(base, current cssmotion.Paint, e Effect, at float64) cssmotion.Paint {
	mode := e.Timing.Composition
	fields := []struct {
		get  valueOf
		kind int
	}{
		{moveX, 0}, {moveY, 1}, {opacity, 2}, {red, 3}, {green, 4}, {blue, 5},
	}
	fallbacks := missingFallbacks(base, mode)
	for i, item := range fields {
		v := e.sample(at, fallbacks[i], item.get)
		if !hasValue(e.Frames, item.get) {
			continue
		}
		apply(&current, base, item.kind, v, mode)
	}
	return current
}

// missingFallbacks returns the value a keyframe that omits a property
// contributes. A replacing composition shows the underlying paint, so an
// omitted property falls back to the base value. An additive composition
// sums the keyframe onto the paint, so an omitted property must add
// nothing, or the paint is added to itself and saturates.
func missingFallbacks(base cssmotion.Paint, mode Composition) [6]float64 {
	if mode != Replace {
		return [6]float64{}
	}
	return [6]float64{0, 0, base.Opacity, base.R, base.G, base.B}
}

func apply(p *cssmotion.Paint, base cssmotion.Paint, field int, v float64, mode Composition) {
	add := mode != Replace
	switch field {
	case 0:
		if add {
			p.X += v
		} else {
			p.X = base.X + v
		}
	case 1:
		if add {
			p.Y += v
		} else {
			p.Y = base.Y + v
		}
	case 2:
		if add {
			p.Opacity += v
		} else {
			p.Opacity = v
		}
	case 3:
		p.R = color(p.R, v, add)
	case 4:
		p.G = color(p.G, v, add)
	case 5:
		p.B = color(p.B, v, add)
	}
}

func color(current, value float64, add bool) float64 {
	if add {
		current += value
	} else {
		current = value
	}
	return math.Max(0, math.Min(1, current))
}

func hasValue(frames []Keyframe, field valueOf) bool {
	for _, frame := range frames {
		if _, ok := field(frame); ok {
			return true
		}
	}
	return false
}
