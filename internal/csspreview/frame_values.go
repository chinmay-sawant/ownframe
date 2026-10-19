package csspreview

type stop struct{ offset, value float64 }

func keyframeX(f []keyframe, p float64, ease string) float64 {
	return interpolateStops(stops(f, func(k keyframe) (float64, bool) { return k.x, k.move }, 0), p, ease)
}
func keyframeY(f []keyframe, p float64, ease string) float64 {
	return interpolateStops(stops(f, func(k keyframe) (float64, bool) { return k.y, k.move }, 0), p, ease)
}

func sampleFrames(f []keyframe, p float64, base paint, fill bool, ease string) paint {
	if len(f) == 0 {
		return base
	}
	base.opacity = interpolateStops(stops(f, func(k keyframe) (float64, bool) { return k.opacity, k.opacitySet }, base.opacity), p, ease)
	if fill {
		base.r = interpolateStops(stops(f, func(k keyframe) (float64, bool) { return k.r, k.colorSet }, base.r), p, ease)
		base.g = interpolateStops(stops(f, func(k keyframe) (float64, bool) { return k.g, k.colorSet }, base.g), p, ease)
		base.b = interpolateStops(stops(f, func(k keyframe) (float64, bool) { return k.b, k.colorSet }, base.b), p, ease)
	}
	return base
}

func stops(f []keyframe, value func(keyframe) (float64, bool), fallback float64) []stop {
	out := []stop{}
	for _, k := range f {
		if v, ok := value(k); ok {
			out = append(out, stop{k.offset, v})
		}
	}
	if len(out) == 0 {
		return []stop{{0, fallback}, {1, fallback}}
	}
	if out[0].offset > 0 {
		out = append([]stop{{0, fallback}}, out...)
	}
	if out[len(out)-1].offset < 1 {
		out = append(out, stop{1, fallback})
	}
	return out
}

func interpolateStops(s []stop, p float64, ease string) float64 {
	if p <= s[0].offset {
		return s[0].value
	}
	for i := 1; i < len(s); i++ {
		if p <= s[i].offset {
			a, b := s[i-1], s[i]
			part := (p - a.offset) / (b.offset - a.offset)
			part = easing(ease, part)
			return a.value + (b.value-a.value)*part
		}
	}
	return s[len(s)-1].value
}
