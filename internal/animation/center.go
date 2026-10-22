package animation

import "strings"

// centerHalf splits progress around a view-timeline center.
const centerHalf = 0.5

func centerProgress(p float64, value string, in FrameInput, lo, hi float64) float64 {
	value = strings.TrimSpace(value)
	if value == "" || value == "normal" {
		return p
	}
	center, ok := in.RangeEdges[value]
	parts := strings.Fields(value)
	if !ok && len(parts) > 0 {
		base := 0.0
		if parts[0] == "target" {
			base = in.TargetProgress
		}
		if parts[0] == "source" {
			base = in.SourceProgress
		}
		center, ok = base, parts[0] == "target" || parts[0] == "source"
		if len(parts) > 1 {
			if p, valid := percent(parts[len(parts)-1]); valid {
				center = base + (in.TargetProgress-in.SourceProgress)*p
				ok = true
			}
		}
	}
	if !ok && len(parts) == 1 {
		if fraction, valid := percent(parts[0]); valid {
			center, ok = lo+(hi-lo)*fraction, true
		}
	}
	if !ok {
		return p
	}
	ratio := (center - lo) / (hi - lo)
	if ratio <= 0 {
		return centerHalf + centerHalf*p
	}
	if ratio >= 1 {
		return centerHalf * p
	}
	if p <= ratio {
		return centerHalf * p / ratio
	}
	return centerHalf + centerHalf*(p-ratio)/(1-ratio)
}
