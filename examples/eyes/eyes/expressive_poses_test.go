package eyes

import (
	"math"
	"testing"
)

func TestAddedExpressionsAnimate(t *testing.T) {
	for _, pose := range poses {
		if pose.name == "spiral" {
			continue
		}
		if pose.motion == motionNone {
			t.Errorf("%s has no animation", pose.name)
			continue
		}
		leftA, rightA := animatePose(pose, .2)
		leftB, rightB := animatePose(pose, 3)
		if styleDistance(leftA, leftB)+styleDistance(rightA, rightB) < .01 {
			t.Errorf("%s does not change over time", pose.name)
		}
	}
}

func TestExpressionsReadThroughTheirEyeCues(t *testing.T) {
	if poseFor(poseIndex("scared")).left.height < 1.1 {
		t.Fatal("scared eyes are not wide enough")
	}
	frustrated := poseFor(poseIndex("frustrated"))
	angry := poseFor(poseIndex("angry"))
	if frustrated.left.open >= angry.left.open || frustrated.left.gazeY <= angry.left.gazeY {
		t.Fatal("frustrated should be narrower and look lower than angry")
	}
}

func styleDistance(a, b eyeStyle) float64 {
	return math.Abs(a.open-b.open) + math.Abs(a.height-b.height) +
		math.Abs(a.top-b.top) + math.Abs(a.bottom-b.bottom) +
		math.Abs(a.slant-b.slant) +
		math.Abs(a.gazeX-b.gazeX) + math.Abs(a.gazeY-b.gazeY)
}
