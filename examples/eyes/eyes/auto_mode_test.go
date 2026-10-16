package eyes

import (
	"testing"
	"time"
)

func TestAutomaticPoseTransitionsEase(t *testing.T) {
	a, c := newT(t)
	a.selectMood(6, c.at)
	tk(t, a, c, time.Second/60)
	if a.st.styleL.height <= 1 || a.st.styleL.height >= 1.16 {
		t.Fatalf("surprised pose did not ease: height %.3f", a.st.styleL.height)
	}
}

func TestAutomaticCyclesEveryPose(t *testing.T) {
	a, c := newT(t)
	for next := 1; next < len(poses); next++ {
		c.adv(moodDuration(a.st.mood) + time.Millisecond)
		tk(t, a, c, time.Second/60)
		if a.st.mood != next {
			t.Fatalf("automatic cycle selected %s, want %s", moodName(a.st.mood), poses[next].name)
		}
	}
}
