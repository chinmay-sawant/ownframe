package motion

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func timing(a layout.Animation) Timing {
	return Timing{Duration: ms(a.DurationMS), Delay: ms(a.DelayStartMS),
		EndDelay: ms(a.DelayEndMS), Iterations: a.IterationCount, Infinite: a.IterationInfinite,
		Direction: direction(a.Direction), Fill: fill(a.FillMode), Play: parsePlayState(a.PlayState),
		TimingFunction: a.TimingFunction, DurationAuto: a.DurationAuto,
		Composition: composition(a.Composition), Timeline: a.Timeline,
		RangeStart: a.RangeStart, RangeEnd: a.RangeEnd, RangeCenter: a.RangeCenter}
}

func composition(value string) Composition {
	switch value {
	case "add":
		return Add
	case "accumulate":
		return Accumulate
	default:
		return Replace
	}
}

func ms(value float64) time.Duration { return time.Duration(value * float64(time.Millisecond)) }

func direction(value string) Direction {
	switch value {
	case "reverse":
		return Reverse
	case "alternate":
		return Alternate
	case "alternate-reverse":
		return AlternateReverse
	default:
		return Normal
	}
}

func fill(value string) FillMode {
	switch value {
	case "forwards":
		return FillForwards
	case "backwards":
		return FillBackwards
	case "both":
		return FillBoth
	default:
		return FillNone
	}
}

func parsePlayState(value string) PlayState {
	if value == "paused" {
		return Paused
	}
	return Running
}
