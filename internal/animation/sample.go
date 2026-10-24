package animation

type valueOf func(Keyframe) (float64, bool)

func moveX(k Keyframe) (float64, bool)   { return k.X, k.Move }
func moveY(k Keyframe) (float64, bool)   { return k.Y, k.Move }
func opacity(k Keyframe) (float64, bool) { return k.Opacity, k.OpacitySet }
func red(k Keyframe) (float64, bool)     { return k.R, k.ColorSet }
func green(k Keyframe) (float64, bool)   { return k.G, k.ColorSet }
func blue(k Keyframe) (float64, bool)    { return k.B, k.ColorSet }

// sample resolves one field at at. The live path uses the precomputed
// effectTable; compose keeps this scan for callers that still hold a bare
// Effect. It walks the frames once and allocates nothing.
func (e Effect) sample(at, fallback float64, value valueOf) float64 {
	var prev, last stop
	havePrev, first := false, true
	for _, frame := range e.Frames {
		v, ok := value(frame)
		if !ok {
			continue
		}
		cur := stop{at: frame.Offset, value: v, ease: frame.ease}
		if first {
			first = false
			if cur.at > 0 {
				if at <= 0 {
					return fallback
				}
				prev, havePrev = stop{at: 0, fromBase: 1}, true
			}
		}
		if havePrev && at <= cur.at {
			return e.mix(prev, cur, at, fallback)
		}
		if !havePrev && at <= cur.at {
			return cur.value
		}
		prev, havePrev, last = cur, true, cur
	}
	if first {
		return fallback
	}
	if last.at < 1 {
		if at > 1 {
			return fallback
		}
		return e.mix(last, stop{at: 1, fromBase: 1}, at, fallback)
	}
	return last.value
}

// mix interpolates between two stops. A base-dependent sentinel is resolved
// against fallback, the target paint for this sample. The easing is the one
// declared on the keyframe that starts the interval, or the effect's own
// function when that keyframe declares none.
func (e Effect) mix(prev, next stop, at, fallback float64) float64 {
	span := next.at - prev.at
	if span <= 0 {
		return next.value
	}
	part := (at - prev.at) / span
	ease := prev.ease
	if ease == nil {
		ease = effectEase(e.Timing)
	}
	part = ease(part)
	return prev.value + (next.value-prev.value)*part +
		(prev.fromBase+(next.fromBase-prev.fromBase)*part)*fallback
}
