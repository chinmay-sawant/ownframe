package eyes

import (
	"math"
	"time"
)

type winkSide uint8

const (
	winkNone winkSide = iota
	winkScreenLeft
	winkScreenRight
)

func winkPulse(elapsed time.Duration) float64 {
	if elapsed < 0 {
		return 0
	}
	phase := elapsed % winkCycle
	switch {
	case phase < winkEdge:
		p := float64(phase) / float64(winkEdge)
		return (1 - math.Cos(math.Pi*p)) / halfDivisor
	case phase < winkHold:
		return 1
	case phase < winkHold+winkEdge:
		p := float64(phase-winkHold) / float64(winkEdge)
		return (1 + math.Cos(math.Pi*p)) / halfDivisor
	default:
		return 0
	}
}
