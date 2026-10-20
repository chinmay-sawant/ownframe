package cssanimation

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func timing(a layout.Animation) animation.Timing {
	return animation.Timing{Duration: ms(a.DurationMS), Delay: ms(a.DelayStartMS),
		EndDelay: ms(a.DelayEndMS), Iterations: a.IterationCount, Infinite: a.IterationInfinite,
		Direction: direction(a.Direction), Fill: fill(a.FillMode), Play: play(a.PlayState),
		TimingFunction: a.TimingFunction, DurationAuto: a.DurationAuto,
		Composition: composition(a.Composition), Timeline: a.Timeline,
		RangeStart: a.RangeStart, RangeEnd: a.RangeEnd, RangeCenter: a.RangeCenter}
}

func composition(value string) animation.Composition {
	switch value {
	case "add":
		return animation.Add
	case "accumulate":
		return animation.Accumulate
	default:
		return animation.Replace
	}
}

func ms(value float64) time.Duration { return time.Duration(value * float64(time.Millisecond)) }

func direction(value string) animation.Direction {
	switch value {
	case "reverse":
		return animation.Reverse
	case "alternate":
		return animation.Alternate
	case "alternate-reverse":
		return animation.AlternateReverse
	default:
		return animation.Normal
	}
}

func fill(value string) animation.FillMode {
	switch value {
	case "forwards":
		return animation.FillForwards
	case "backwards":
		return animation.FillBackwards
	case "both":
		return animation.FillBoth
	default:
		return animation.FillNone
	}
}

func play(value string) animation.PlayState {
	if value == "paused" {
		return animation.Paused
	}
	return animation.Running
}
