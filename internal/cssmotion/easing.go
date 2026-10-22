package cssmotion

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
