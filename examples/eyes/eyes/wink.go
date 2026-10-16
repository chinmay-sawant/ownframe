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
	phase := elapsed % (2400 * time.Millisecond)
	switch {
	case phase < 240*time.Millisecond:
		p := float64(phase) / float64(240*time.Millisecond)
		return (1 - math.Cos(math.Pi*p)) / 2
	case phase < 420*time.Millisecond:
		return 1
	case phase < 660*time.Millisecond:
		p := float64(phase-420*time.Millisecond) / float64(240*time.Millisecond)
		return (1 + math.Cos(math.Pi*p)) / 2
	default:
		return 0
	}
}
