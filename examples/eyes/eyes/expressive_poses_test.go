package eyes

import (
	"math"
	"testing"
)

func TestAddedExpressionsAnimate(t *testing.T) {
	names := []string{"sad", "embarrassed", "scared", "frustrated", "suspicious",
		"smug", "mischievous", "disgusted", "proud", "pleading", "relieved", "excited"}
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

func TestExpressionsReadThroughTheirEyeCues(t *testing.T) {
	if poseFor(poseIndex("scared")).left.height < 1.1 {
		t.Fatal("scared eyes are not wide enough")
	}
	if poseFor(poseIndex("disgusted")).left.gazeY >= -5 {
		t.Fatal("disgusted gaze should look upward and away")
	}
	if poseFor(poseIndex("proud")).left.gazeY >= -10 {
		t.Fatal("proud gaze should look clearly upward")
	}
	pleading := poseFor(poseIndex("pleading"))
	if pleading.left.gazeY > -12 || pleading.left.height < 1.1 || pleading.left.open < .96 ||
		pleading.left.browUp < 20 || pleading.left.browSlant > -15 ||
		pleading.right.browSlant < 15 {
		t.Fatal("pleading needs wide upturned eyes and raised inner brows")
	}
}

func styleDistance(a, b eyeStyle) float64 {
	return math.Abs(a.open-b.open) + math.Abs(a.height-b.height) +
		math.Abs(a.top-b.top) + math.Abs(a.bottom-b.bottom) +
		math.Abs(a.slant-b.slant) + math.Abs(a.gazeX-b.gazeX) +
		math.Abs(a.gazeY-b.gazeY) + math.Abs(a.browUp-b.browUp) +
		math.Abs(a.browSlant-b.browSlant)
}
