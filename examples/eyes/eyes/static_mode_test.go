package eyes

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestStaticModeHoldsPoseWhileAnimationsRun(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	a.toggleMode(c.at)
	a.selectMood(spiralMood, c.at)
	a.st.blinkT = c.at.Add(850 * time.Millisecond)
	for range 20 {
		tk(t, a, c, 50*time.Millisecond)
	}
	if a.st.automatic || a.st.mood != spiralMood || a.st.spiralMix < .95 {
		t.Fatal("static mode did not hold the selected spiral pose")
	}
	angle := a.st.spiral
	if !a.st.blinkOn {
		t.Fatal("static mode stopped the blink animation")
	}
	tk(t, a, c, 50*time.Millisecond)
	if a.st.spiral == angle {
		t.Fatal("static mode stopped the spiral animation")
	}
	a.st.blinkOn = false
	if err := a.onKeyDown(t.Context(), "space"); err != nil {
		t.Fatal(err)
	}
	if !a.st.blinkOn {
		t.Fatal("manual mode ignored a blink command")
	}
	if err := a.onTap(t.Context(), ownframe.Box{ID: "eye-left"}); err != nil {
		t.Fatal(err)
	}
	if a.st.mood != spiralMood+1 {
		t.Fatal("static tap did not change the selected expression")
	}
}
