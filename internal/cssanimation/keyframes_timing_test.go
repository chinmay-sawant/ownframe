package cssanimation

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// keyframeTimingSource is a three-stop rule whose timing function is declared
// beside a paint property on some stops and alone on another.
func keyframeTimingSource() []layout.Keyframe {
	return []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "opacity", Value: "1"},
			{Property: "animation-timing-function", Value: " ease-out "},
		}},
		{Offset: 0.5, Declarations: []layout.KeyframeDeclaration{
			{Property: "animation-timing-function", Value: "steps(2, jump-none)"},
		}},
		{Offset: 0.75, Declarations: []layout.KeyframeDeclaration{
			{Property: "-webkit-animation-timing-function", Value: "ease-in"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "opacity", Value: "0"},
		}},
	}
}

// keyframeTimingWants is the function each stop of keyframeTimingSource
// declares, in order.
func keyframeTimingWants() []string {
	return []string{"ease-out", "steps(2, jump-none)", "ease-in", ""}
}

// TestKeyframeTimingFunctionSurvivesClassParsing pins that the timing function
// is read before the paint-class filter: it belongs to the keyframe, not to one
// property family, so every class the rule produces carries it.
func TestKeyframeTimingFunctionSurvivesClassParsing(t *testing.T) {
	size := transformSize{width: 200, height: 100, pointsPerPixel: 1, hasBox: true}
	class := paintClass{kind: cssmotion.KindSubtree, mask: cssmotion.MaskOpacity}

	frames := keyframesFor(keyframeTimingSource(), size, class, Color{}, false)
	wants := keyframeTimingWants()
	if len(frames) != len(wants) {
		t.Fatalf("frames = %d, want %d", len(frames), len(wants))
	}
	for i, want := range wants {
		if frames[i].TimingFunction != want {
			t.Fatalf("frame %d timing function = %q, want %q", i, frames[i].TimingFunction, want)
		}
	}
	if !frames[0].OpacitySet || !frames[3].OpacitySet {
		t.Fatal("the paint declarations were lost while keeping the timing function")
	}
	for i := 1; i < 3; i++ {
		if frames[i].OpacitySet {
			t.Fatalf("timing-function-only frame %d invented an opacity", i)
		}
	}
}

// TestKeyframeTimingFunctionIsNotAPaintGap pins that the declaration is not
// mistaken for an unsupported paint property or a deferred phase.
func TestKeyframeTimingFunctionIsNotAPaintGap(t *testing.T) {
	if _, ok := propertyClass("animation-timing-function"); ok {
		t.Fatal("the timing function was classified as a paint property")
	}
	if phase, ok := deferredPhase("animation-timing-function"); ok {
		t.Fatalf("the timing function was deferred to phase %d", phase)
	}
	if !keyframeTimingFunction(layout.KeyframeDeclaration{Property: "animation-timing-function"}) {
		t.Fatal("the keyframe parser did not recognise the timing function")
	}
}

// TestKeyframeTimingFunctionReachesTheEffectTiming checks the end-to-end path a
// stylesheet takes: the rule, the animation record, and the bound effect all
// agree on the segment easing.
func TestKeyframeTimingFunctionReachesTheEffectTiming(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "fade", Keyframes: keyframeTimingSource()}}
	frames := keyframes(rules, "fade", 1, 200, 100)
	wants := keyframeTimingWants()
	if len(frames) != len(wants) {
		t.Fatalf("frames = %d, want %d", len(frames), len(wants))
	}
	for i, want := range wants {
		if frames[i].TimingFunction != want {
			t.Fatalf("frame %d timing function = %q, want %q", i, frames[i].TimingFunction, want)
		}
	}
}
