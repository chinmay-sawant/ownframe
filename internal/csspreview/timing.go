package csspreview

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func motionTiming(box ownframe.Box, target target) (time.Duration, time.Duration, time.Duration) {
	if target.kind == "animation" {
		if len(box.Animations) == 0 {
			return 0, 0, 0
		}
		item := box.Animations[0]
		duration := milliseconds(item.DurationMS)
		return duration, milliseconds(item.DelayStartMS), milliseconds(item.DelayEndMS)
	}
	for _, item := range box.Transitions {
		if item.Property == target.property || item.Property == "all" {
			return milliseconds(item.DurationMS), milliseconds(item.DelayMS), 0
		}
	}
	return 0, 0, 0
}

func milliseconds(value float64) time.Duration {
	return time.Duration(value * float64(time.Millisecond))
}
