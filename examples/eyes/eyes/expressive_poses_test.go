package eyes

import (
	"math"
	"testing"
)

func TestAddedExpressionsAnimate(t *testing.T) {
	names := []string{"sad", "embarrassed", "scared", "frustrated", "suspicious",
		"smug", "mischievous", "disgusted", "proud", "pleading"}
	for _, name := range names {
		pose := poseFor(poseIndex(name))
		if pose.motion == motionNone {
			t.Errorf("%s has no animation", name)
			continue
		}
		leftA, rightA := animatePose(pose, .2)
		leftB, rightB := animatePose(pose, 1.1)
		if styleDistance(leftA, leftB)+styleDistance(rightA, rightB) < .01 {
			t.Errorf("%s does not change over time", name)
		}
	}
}

func styleDistance(a, b eyeStyle) float64 {
	return math.Abs(a.open-b.open) + math.Abs(a.height-b.height) +
		math.Abs(a.top-b.top) + math.Abs(a.bottom-b.bottom) +
		math.Abs(a.slant-b.slant) + math.Abs(a.gazeX-b.gazeX) +
		math.Abs(a.gazeY-b.gazeY)
}
