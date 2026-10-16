package eyes

import (
	"testing"
	"time"
)

func TestRelaxedPoseLooksSoftAndBreathes(t *testing.T) {
	mood := poseIndex("relaxed")
	if mood < 0 {
		t.Fatal("relaxed pose is missing")
	}
	pose := poseFor(mood)
	if pose.left.open > .75 || pose.left.gazeY < 3 {
		t.Fatal("relaxed pose needs softer eyes and a lower gaze")
	}
	a, c := newT(t)
	a.toggleMode(c.at)
	a.selectMood(mood, c.at)
	a.st.styleL, a.st.styleR = pose.left, pose.right
	a.st.blinkT = c.at.Add(time.Hour)
	tk(t, a, c, 20*time.Millisecond)
	start := a.pm.lidLT.H
	c.adv(700 * time.Millisecond)
	tk(t, a, c, 20*time.Millisecond)
	inhale := a.pm.lidLT.H
	c.adv(1100 * time.Millisecond)
	tk(t, a, c, 20*time.Millisecond)
	exhale := a.pm.lidLT.H
	if inhale >= start || exhale <= start {
		t.Fatal("relaxed eyelids did not rise and fall with a slow breath")
	}
}
