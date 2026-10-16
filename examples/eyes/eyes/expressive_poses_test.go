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

func TestEmotionMotionCues(t *testing.T) {
	left, _ := animatePose(poseFor(poseIndex("happy")), .7)
	grin, _ := animatePose(poseFor(poseIndex("happy")), 1.5)
	if left.open <= grin.open || grin.bottom <= left.bottom {
		t.Fatal("happy eyes should brighten, then squeeze into a grin")
	}

	boredLeft, _ := boredGaze(.5)
	boredRight, _ := boredGaze(3)
	if boredLeft >= 0 || boredRight <= 0 {
		t.Fatal("bored eyes should move between slow side glances")
	}

	excited := poseFor(poseIndex("excited"))
	a, _ := animatePose(excited, .2)
	b, _ := animatePose(excited, .3)
	if math.Abs(a.gazeX-b.gazeX) < 1 {
		t.Fatal("excited eyes should shake between frames")
	}

	nervous := poseFor(poseIndex("nervous"))
	if nervous.left.gazeY < 5 || nervous.left.top < .45 || nervous.left.open >= .86 {
		t.Fatal("nervous eyes should look softer and sadder")
	}
}

func styleDistance(a, b eyeStyle) float64 {
	return math.Abs(a.open-b.open) + math.Abs(a.height-b.height) +
		math.Abs(a.top-b.top) + math.Abs(a.bottom-b.bottom) +
		math.Abs(a.slant-b.slant) +
		math.Abs(a.gazeX-b.gazeX) + math.Abs(a.gazeY-b.gazeY)
}
