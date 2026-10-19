package cssanimation

import (
	"sort"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func keyframes(rules []layout.KeyframesRule, name string, pointsPerPixel float64) []animation.Keyframe {
	var source []layout.Keyframe
	for _, rule := range rules {
		if rule.Name == name {
			source = rule.Keyframes
		}
	}
	frames := make([]animation.Keyframe, len(source))
	for i, item := range source {
		frame := animation.Keyframe{Offset: item.Offset}
		for _, declaration := range item.Declarations {
			switch declaration.Property {
			case "transform":
				frame.X, frame.Y, frame.Move = transformShift(declaration.Value)
				frame.X *= pointsPerPixel
				frame.Y *= pointsPerPixel
			case "opacity":
				frame.Opacity, frame.OpacitySet = number(declaration.Value)
			case "background-color":
				frame.R, frame.G, frame.B, frame.ColorSet = color(declaration.Value)
			}
		}
		frames[i] = frame
	}
	sort.SliceStable(frames, func(i, j int) bool { return frames[i].Offset < frames[j].Offset })
	return frames
}
