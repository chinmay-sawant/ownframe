package cssmotion

import (
	"strconv"
	"strings"
)

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
