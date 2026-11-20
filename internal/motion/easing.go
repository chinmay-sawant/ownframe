package motion

import (
	"strconv"
	"strings"
)

// Bezier handles for ease, ease-in, ease-out and ease-in-out.
const (
	easeX1, easeY1, easeX2 = 0.25, 0.1, 0.25
	easeInX1               = 0.42
	easeOutX2              = 0.58
	bezierHandleCount      = 4
)

// Ease samples a CSS timing-function string at progress t.
func Ease(value string, t float64) float64 {
	t = clamp(t)
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "linear":
		return t
	case "ease":
		return cubic(t, easeX1, easeY1, easeX2, 1)
	case "ease-in":
		return cubic(t, easeInX1, 0, 1, 1)
	case "ease-out":
		return cubic(t, 0, 0, easeOutX2, 1)
	case "ease-in-out":
		return cubic(t, easeInX1, 0, easeOutX2, 1)
	case "step-start":
		return 1
	case "step-end":
		return steps(t, 1, "end")
	}
	name, body, ok := easingParts(value)
	if !ok {
		return t
	}
	switch name {
	case "cubic-bezier":
		v := commaNumbers(body)
		if len(v) == bezierHandleCount {
			return cubic(t, v[0], v[1], v[2], v[3])
		}
	case "steps":
		v := strings.Split(body, ",")
		n, err := strconv.Atoi(strings.TrimSpace(v[0]))
		if err == nil && n > 0 {
			pos := "end"
			if len(v) > 1 {
				pos = strings.TrimSpace(v[1])
			}
			return clamp(steps(t, n, pos))
		}
	case "linear":
		return linearStops(body, t)
	}
	return t
}

func easingParts(s string) (string, string, bool) {
	s = strings.TrimSpace(s)
	i := strings.IndexByte(s, '(')
	if i < 1 || !strings.HasSuffix(strings.TrimSpace(s), ")") {
		return "", "", false
	}
	return strings.ToLower(strings.TrimSpace(s[:i])),
		strings.TrimSuffix(strings.TrimSpace(s[i+1:]), ")"), true
}

func commaNumbers(s string) []float64 {
	parts := strings.Split(s, ",")
	values := make([]float64, len(parts))
	for i, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil
		}
		values[i] = value
	}
	return values
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// halfDivisor splits a bisection range into two halves.
const halfDivisor = 2

func cubic(t, x1, y1, x2, y2 float64) float64 {
	lo, hi, u := 0.0, 1.0, t
	for i := 0; i < 24; i++ {
		x := bezier(u, x1, x2)
		if x < t {
			lo = u
		} else {
			hi = u
		}
		u = (lo + hi) / halfDivisor
	}
	return bezier(u, y1, y2)
}

func bezier(t, a, b float64) float64 {
	u := 1 - t
	return 3*u*u*t*a + 3*u*t*t*b + t*t*t
}

func steps(t float64, count int, position string) float64 {
	if count <= 1 && position == "jump-none" {
		return t
	}
	switch position {
	case "start", "jump-start":
		return float64(int(t*float64(count))+1) / float64(count)
	case "jump-both":
		return float64(int(t*float64(count))+1) / float64(count+1)
	case "jump-none":
		return float64(int(t*float64(count))) / float64(count-1)
	default:
		return float64(int(t*float64(count))) / float64(count)
	}
}

const (
	// maxLinearFields is "value" plus an optional stop position.
	maxLinearFields = 2
	// minLinearStops is the smallest usable linear() stop list.
	minLinearStops = 2
	// percentDivisor converts a percent stop position to a fraction.
	percentDivisor = 100
)

func linearStops(body string, t float64) float64 {
	parts := strings.Split(body, ",")
	values, positions, set := make([]float64, len(parts)), make([]float64, len(parts)), make([]bool, len(parts))
	for i, part := range parts {
		fields := strings.Fields(part)
		if len(fields) < 1 || len(fields) > maxLinearFields {
			return t
		}
		value, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return t
		}
		values[i] = value
		if len(fields) == maxLinearFields {
			position, err := strconv.ParseFloat(strings.TrimSuffix(fields[1], "%"), 64)
			if err != nil {
				return t
			}
			positions[i], set[i] = position/percentDivisor, true
		}
	}
	if len(parts) < minLinearStops {
		return t
	}
	set[0] = true
	if !set[len(parts)-1] {
		positions[len(parts)-1], set[len(parts)-1] = 1, true
	}
	positions[0] = 0
	for i := 1; i < len(parts); i++ {
		if set[i] {
			continue
		}
		j := i + 1
		for j < len(parts) && !set[j] {
			j++
		}
		if j == len(parts) {
			return t
		}
		step := (positions[j] - positions[i-1]) / float64(j-i+1)
		positions[i], set[i] = positions[i-1]+step, true
	}
	for i := 1; i < len(parts); i++ {
		if t <= positions[i] {
			span := positions[i] - positions[i-1]
			if span <= 0 {
				return values[i]
			}
			part := (t - positions[i-1]) / span
			return values[i-1] + (values[i]-values[i-1])*part
		}
	}
	return values[len(values)-1]
}
