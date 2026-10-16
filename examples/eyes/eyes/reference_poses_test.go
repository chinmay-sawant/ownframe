package eyes

import (
	"testing"
	"time"
)

func TestReferenceExpressionsKeepBrowsThroughBalancedBlinks(t *testing.T) {
	for _, name := range []string{"smug", "mischievous", "pleading"} {
		t.Run(name, func(t *testing.T) {
			a, c := newT(t)
			a.st.blinkT = c.at.Add(time.Hour)
			a.selectMood(poseIndex(name), c.at)
			for range 30 {
				tk(t, a, c, 20*time.Millisecond)
			}
			if a.pm.browL == nil || a.pm.browR == nil ||
				a.pm.browL.Alpha == 0 || a.pm.browR.Alpha == 0 {
				t.Fatal("expression eyebrows were not painted")
			}
			a.st.blinkOn, a.st.blinkP = true, .5
			tk(t, a, c, 20*time.Millisecond)
			if a.pm.lidLT.H < 20 || a.pm.lidLB.H < 20 ||
				a.pm.lidRT.H < 20 || a.pm.lidRB.H < 20 {
				t.Fatal("blink did not close from both lids")
			}
			if a.pm.browL.Alpha == 0 || a.pm.browR.Alpha == 0 {
				t.Fatal("blink hid the expression eyebrows")
			}
		})
	}
}
