package eyes

import (
	"strings"
	"testing"
	"time"
)

func TestNoEyebrowOrLidMarkup(t *testing.T) {
	html := buildHTML()
	if strings.Contains(stylesCSS, ".brow") || strings.Contains(stylesCSS, ".lid") {
		t.Fatal("eyebrow or lid styles remain")
	}
	for _, unwanted := range []string{"brow-left", "brow-right", "lid-left", "lid-right", "class=\"brow", "class=\"lid"} {
		if strings.Contains(html, unwanted) || strings.Contains(stylesCSS, unwanted) {
			t.Fatalf("found unwanted eyebrow or lid markup %q", unwanted)
		}
	}
}

func TestSquintReducesRoundedEyeHeight(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	openHeight := a.pm.left.H
	a.st.mood = 1
	for range 90 {
		tk(t, a, c, time.Second/60)
	}
	if a.pm.left.H >= openHeight*.5 {
		t.Fatalf("squint eye height %.1f is not less than half of open %.1f", a.pm.left.H, openHeight)
	}
}

func TestBlinkScalesEyesAndRestores(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	openHeight := a.pm.left.H
	a.st.blinkT = c.at
	tk(t, a, c, time.Second/60)
	if a.pm.left.H >= openHeight {
		t.Fatal("blink did not reduce eye height")
	}
	for range 20 {
		tk(t, a, c, time.Second/60)
	}
	if a.pm.left.H < openHeight*.9 {
		t.Fatalf("eye height did not recover after blinking: %.1f of %.1f", a.pm.left.H, openHeight)
	}
}
