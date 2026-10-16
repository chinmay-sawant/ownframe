package eyes

import (
	"strings"
	"testing"
	"time"
)

func TestNoEyebrowMarkup(t *testing.T) {
	html := buildHTML()
	if strings.Contains(stylesCSS, ".brow") {
		t.Fatal("eyebrow styles remain")
	}
	for _, unwanted := range []string{"brow-left", "brow-right", "class=\"brow"} {
		if strings.Contains(html, unwanted) || strings.Contains(stylesCSS, unwanted) {
			t.Fatalf("found unwanted eyebrow markup %q", unwanted)
		}
	}
}

func TestSquintPreservesOuterEyeShape(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	openHeight := a.pm.left.H
	a.st.mood = 1
	for range 90 {
		tk(t, a, c, time.Second/60)
	}
	if a.pm.left.H < openHeight*.95 {
		t.Fatalf("squint shrank the outer eye from %.1f to %.1f", openHeight, a.pm.left.H)
	}
	if a.pm.lidLT == nil || a.pm.lidLT.H <= 0 {
		t.Fatal("squint did not shape the eye opening with an eyelid")
	}
}

func TestBlinkMovesEyelidsAndPreservesEyeShape(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	openHeight := a.pm.left.H
	a.st.blinkT = c.at
	tk(t, a, c, time.Second/60)
	if a.pm.left.H != openHeight || a.pm.lidLT.H <= 0 {
		t.Fatal("blink changed the eye shell instead of closing the eyelids")
	}
	for range 20 {
		tk(t, a, c, time.Second/60)
	}
	if a.pm.left.H != openHeight || a.pm.lidLT.H > 10 || a.pm.lidLB.H > 10 {
		t.Fatalf("eye shell or eyelids did not return to the open shape")
	}
}
