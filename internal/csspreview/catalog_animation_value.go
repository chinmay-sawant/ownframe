package csspreview

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
)

func animationValue(a layout.Animation, name string) string {
	name = strings.TrimPrefix(name, "-webkit-")
	switch name {
	case "animation":
		return fmt.Sprintf("%s %gms %s %g %s %s", a.Name, a.DurationMS, a.TimingFunction, a.IterationCount, a.Direction, a.FillMode)
	case "animation-delay":
		return fmt.Sprintf("start %gms, end %gms", a.DelayStartMS, a.DelayEndMS)
	case "animation-delay-start":
		return fmt.Sprintf("%gms", a.DelayStartMS)
	case "animation-delay-end":
		return fmt.Sprintf("%gms", a.DelayEndMS)
	case "animation-direction":
		return a.Direction
	case "animation-duration":
		return fmt.Sprintf("%gms", a.DurationMS)
	case "animation-fill-mode":
		return a.FillMode
	case "animation-iteration-count":
		if a.IterationInfinite {
			return "infinite"
		}
		return fmt.Sprintf("%g", a.IterationCount)
	case "animation-name":
		return a.Name
	case "animation-play-state":
		return a.PlayState
	case "animation-timing-function":
		return a.TimingFunction
	case "animation-composition":
		return a.Composition
	case "animation-range":
		return a.RangeStart + " to " + a.RangeEnd
	case "animation-range-center":
		return a.RangeCenter
	case "animation-range-start":
		return a.RangeStart
	case "animation-range-end":
		return a.RangeEnd
	case "animation-timeline":
		return a.Timeline
	case "animation-trigger":
		return a.Trigger
	}
	return "not present"
}
