package csspreview

import (
	"sort"

	"github.com/chinmay-sawant/blinkless/layout"
)

type keyframe struct {
	offset, x, y, opacity      float64
	move, opacitySet, colorSet bool
	r, g, b                    float64
}

func namedKeyframes(rules []layout.KeyframesRule, name string) []keyframe {
	var selected []layout.Keyframe
	for _, rule := range rules {
		if rule.Name == name {
			selected = rule.Keyframes
		}
	}
	frames := make([]keyframe, len(selected))
	for i, source := range selected {
		f := keyframe{offset: source.Offset}
		for _, decl := range source.Declarations {
			switch decl.Property {
			case "transform":
				f.x, f.y, f.move = transformShift(decl.Value)
			case "opacity":
				f.opacity, f.opacitySet = number(decl.Value)
			case "background-color":
				f.r, f.g, f.b, f.colorSet = color(decl.Value)
			}
		}
		frames[i] = f
	}
	sort.SliceStable(frames, func(i, j int) bool { return frames[i].offset < frames[j].offset })
	return frames
}
