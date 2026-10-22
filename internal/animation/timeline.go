package animation

import (
	"strconv"
	"strings"
)

// percentDivisor scales a percent value to a fraction.
const percentDivisor = 100

func timelineProgress(t Timing, in FrameInput) (float64, bool) {
	name := strings.TrimSpace(t.Timeline)
	switch {
	case name == "" || name == "auto":
		return 0, true
	case name == "none":
		return 0, false
	case strings.HasPrefix(name, "scroll("):
		return unit(in.ScrollProgress), true
	case strings.HasPrefix(name, "view("):
		return unit(in.ViewProgress), true
	}
	v, ok := in.Timelines[name]
	return unit(v), ok
}

func timelineRange(t Timing, in FrameInput) (float64, float64, bool) {
	start, ok := edge(t.RangeStart, false, in)
	if !ok {
		return 0, 0, false
	}
	end, ok := edge(t.RangeEnd, true, in)
	return start, end, ok
}

func edge(value string, end bool, in FrameInput) (float64, bool) {
	value = strings.TrimSpace(value)
	if v, ok := in.RangeEdges[value]; ok {
		return v, true
	}
	if value == "" || value == "normal" {
		if end {
			return 1, true
		}
		return 0, true
	}
	parts := strings.Fields(value)
	if len(parts) == 1 {
		if v, ok := percent(parts[0]); ok {
			return v, true
		}
	}
	r, ok := in.Ranges[parts[0]]
	if !ok {
		return 0, false
	}
	if len(parts) == 1 {
		if end {
			return r.End, true
		}
		return r.Start, true
	}
	p, ok := percent(parts[len(parts)-1])
	if !ok {
		return 0, false
	}
	return r.Start + (r.End-r.Start)*p, true
}

func percent(s string) (float64, bool) {
	if !strings.HasSuffix(s, "%") {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
	return v / percentDivisor, err == nil
}

func unit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
