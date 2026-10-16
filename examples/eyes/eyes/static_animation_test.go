package eyes

import (
	"math"
	"testing"
	"time"
)

func TestStaticAngryBrowsKeepMoving(t *testing.T) {
	a, c := newT(t)
	a.toggleMode(c.at)
	a.selectMood(2, c.at)
	a.st.blinkT = c.at.Add(time.Hour)
	tk(t, a, c, time.Second/60)
	first := a.pm.lidLT.Transform().B
	c.adv(500 * time.Millisecond)
	tk(t, a, c, time.Second/60)
	if math.Abs(a.pm.lidLT.Transform().B-first) < 0.01 {
		t.Fatal("angry brow stopped moving in static mode")
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
