package eyes

import (
	"testing"
	"time"
)

func TestScaredEyesScanLeftAndRight(t *testing.T) {
	pose := poseFor(poseIndex("scared"))
	left, _ := animatePose(pose, .4)
	right, _ := animatePose(pose, 1.2)
	if left.gazeX < 15 || right.gazeX > -15 {
		t.Fatalf("scared gaze did not dart both ways: %.1f then %.1f", left.gazeX, right.gazeX)
	}
}

func TestWinksAnimateOnViewedSides(t *testing.T) {
	for _, tc := range []struct {
		name      string
		closedEye string
	}{{"wink-left", "right"}, {"wink-right", "left"}} {
		a, c := newT(t)
		a.toggleMode(c.at)
		a.st.blinkT = c.at.Add(time.Hour)
		a.selectMood(poseIndex(tc.name), c.at)
		for range 20 {
			tk(t, a, c, 20*time.Millisecond)
		}
		closedLeft := a.pm.lidLT.H > a.pm.lidRT.H
		if (tc.closedEye == "left") != closedLeft {
			t.Fatalf("%s closed the wrong eye", tc.name)
		}
		for range 30 {
			tk(t, a, c, 20*time.Millisecond)
		}
		if a.pm.lidLT.H > 10 || a.pm.lidRT.H > 10 {
			t.Fatalf("%s did not reopen both eyes", tc.name)
		}
	}
}

func poseIndex(name string) int {
	for i, pose := range poses {
		if pose.name == name {
			return i
		}
	}
	return -1
}
