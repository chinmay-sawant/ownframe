package eyes

import (
	"math"
	"testing"
	"time"
)

func TestShakeSpiralHoldsAndFadesOut(t *testing.T) {
	a, c := newT(t)
	a.Shake()
	if a.st.mood == 3 {
		t.Fatal("shake changed state outside the frame tick")
	}
	for pulse := range 20 {
		if pulse > 0 {
			a.Shake()
		}
		for range 18 {
			tk(t, a, c, time.Second/60)
		}
		if a.st.mood != 3 || a.st.spiralMix < 0.95 {
			t.Fatalf("spiral did not hold through shake pulse %d", pulse)
		}
	}
	angle := a.st.spiral
	for range 12 {
		tk(t, a, c, time.Second/60)
	}
	advanced := math.Mod(a.st.spiral-angle+360, 360)
	if advanced < 100 || advanced > 140 {
		t.Fatalf("spiral advanced %.1f degrees in 200 ms", advanced)
	}
	c.adv(9 * time.Second)
	tk(t, a, c, time.Second/60)
	if a.st.mood != 0 || a.st.spiralMix <= 0 || a.st.spiralMix >= 1 {
		t.Fatal("spiral did not start a smooth return to open eyes")
	}
	for range 45 {
		tk(t, a, c, time.Second/60)
	}
	if a.st.spiralMix > 0.001 {
		t.Fatal("spiral did not finish fading out")
	}
}

func TestTiltMovesEyesWithDeviceDirection(t *testing.T) {
	a, c := newT(t)
	a.SetMotion(0.5, 0.5)
	tk(t, a, c, time.Second/60)
	if a.st.lookX <= 10 || a.st.lookY >= -7.5 {
		t.Fatalf("tilt response was too slow or weak: (%.1f, %.1f)", a.st.lookX, a.st.lookY)
	}
}
