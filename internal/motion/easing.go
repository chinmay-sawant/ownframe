package motion

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Bezier handles for ease, ease-in, ease-out and ease-in-out.
const (
	easeX1, easeY1, easeX2 = 0.25, 0.1, 0.25
	easeInX1               = 0.42
	easeOutX2              = 0.58
	bezierHandleCount      = 4
)

// Ease samples a CSS timing-function string at progress t. It parses value on
// every call; per-frame callers use compileEase once and sample the result.
func Ease(value string, t float64) float64 {
	return cachedEase(value)(t)
}

// easeCacheLimit bounds the compiled-sampler cache; timing functions come
// from author CSS, so the set of distinct strings is small in practice.
const easeCacheLimit = 512

var (
	easeCache     sync.Map
	easeCacheSize atomic.Int32
)

// cachedEase returns the compiled sampler for value, compiling it on first
// use so repeated string-keyed callers parse each timing function once.
func cachedEase(value string) easingFn {
	if fn, ok := easeCache.Load(value); ok {
		if sampler, isFn := fn.(easingFn); isFn {
			return sampler
		}
	}
	fn := compileEase(value)
	if easeCacheSize.Load() < easeCacheLimit {
		if _, loaded := easeCache.LoadOrStore(value, fn); !loaded {
			easeCacheSize.Add(1)
		}
	}
	return fn
}

// compileEase parses a CSS timing-function string once into a sampler. The
// returned function clamps progress to [0,1] and never touches the string
// again, so sampling allocates nothing.
func compileEase(value string) easingFn {
	sample := parseEase(value)
	return func(t float64) float64 { return sample(clamp(t)) }
}

func identityEase(t float64) float64 { return t }

func bezierEase(x1, y1, x2, y2 float64) easingFn {
	return func(t float64) float64 { return cubic(t, x1, y1, x2, y2) }
}

func parseEase(value string) easingFn {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "linear":
		return identityEase
	case "ease":
		return bezierEase(easeX1, easeY1, easeX2, 1)
	case "ease-in":
		return bezierEase(easeInX1, 0, 1, 1)
	case "ease-out":
		return bezierEase(0, 0, easeOutX2, 1)
	case "ease-in-out":
		return bezierEase(easeInX1, 0, easeOutX2, 1)
	case "step-start":
		return func(float64) float64 { return 1 }
	case "step-end":
		return func(t float64) float64 { return steps(t, 1, "end") }
	}
	name, body, ok := easingParts(value)
	if !ok {
		return identityEase
	}
	switch name {
	case "cubic-bezier":
		v := commaNumbers(body)
		if len(v) == bezierHandleCount {
			return bezierEase(v[0], v[1], v[2], v[3])
		}
	case "steps":
		v := strings.Split(body, ",")
		n, err := strconv.Atoi(strings.TrimSpace(v[0]))
		if err == nil && n > 0 {
			pos := "end"
			if len(v) > 1 {
				pos = strings.TrimSpace(v[1])
			}
			return func(t float64) float64 { return clamp(steps(t, n, pos)) }
		}
	case "linear":
		return parseLinearStops(body)
	}
	return identityEase
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

// parseLinearStops resolves a linear() stop list once. A malformed list
// samples as identity.
func parseLinearStops(body string) easingFn {
	parts := strings.Split(body, ",")
	values, positions, set := make([]float64, len(parts)), make([]float64, len(parts)), make([]bool, len(parts))
	for i, part := range parts {
		fields := strings.Fields(part)
		if len(fields) < 1 || len(fields) > maxLinearFields {
			return identityEase
		}
		value, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return identityEase
		}
		values[i] = value
		if len(fields) == maxLinearFields {
			position, err := strconv.ParseFloat(strings.TrimSuffix(fields[1], "%"), 64)
			if err != nil {
				return identityEase
			}
			positions[i], set[i] = position/percentDivisor, true
		}
	}
	if len(parts) < minLinearStops {
		return identityEase
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
			return identityEase
		}
		step := (positions[j] - positions[i-1]) / float64(j-i+1)
		positions[i], set[i] = positions[i-1]+step, true
	}
	return func(t float64) float64 {
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
}
