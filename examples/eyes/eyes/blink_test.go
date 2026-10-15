package eyes

import (
	"testing"
	"time"
)

func TestBlink(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	a.st.blinkT = c.at
	tk(t, a, c, time.Second/60)
	if !a.st.blinkOn {
		t.Fatal("blink did not start")
	}
	shut := a.st.dispL
	for range 6 {
		tk(t, a, c, time.Second/60)
	}
	if a.st.dispL >= shut && shut < 0.9 {
		t.Fatal("blink did not keep closing")
	}
	for range 14 {
		tk(t, a, c, time.Second/60)
	}
	if a.st.blinkOn {
		t.Fatal("blink did not finish")
	}
	if a.st.dispL < 0.9 {
		t.Fatalf("eye never reopened: %v", a.st.dispL)
	}
}

func TestMoodGlides(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	a.st.mood = 1
	first := 0.0
	for range 10 {
		tk(t, a, c, time.Second/60)
		if first == 0 {
			first = a.st.moodL
		}
	}
	if a.st.moodL >= 1.0 || a.st.moodL <= 0.32 {
		t.Fatalf("mood did not glide: %v", a.st.moodL)
	}
	if first <= a.st.moodL {
		t.Fatal("mood did not keep easing down")
	}
}

func TestMoodOpen(t *testing.T) {
	for i := range moods {
		l, r := targetOpen(i)
		if l <= 0 || r <= 0 || l > 1.3 || r > 1.3 {
			t.Fatalf("mood %d open %v %v", i, l, r)
		}
	}
}
