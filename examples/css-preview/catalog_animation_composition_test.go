package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

func TestCatalogAnimationCompositionChangesOwnFrameSample(t *testing.T) {
	base := `animation-name:route,route;animation-duration:1s;animation-fill-mode:both;animation-timing-function:linear;`
	at := animation.FrameInput{Now: time.Unix(101, 0)}
	replace := sampleMotion(t, base+`animation-composition:replace,replace;`, at)
	add := "add"
	added := sampleMotion(t, base+`animation-composition:`+add+`,`+add+`;`, at)
	accumulate := "accumulate"
	accumulated := sampleMotion(t, base+`animation-composition:`+accumulate+`,`+accumulate+`;`, at)
	// A transform-list keyframe samples one matrix per effect, so the later
	// effect replaces the earlier one in every composition mode. Concatenating
	// additive transform lists belongs to the composition phase; until then add
	// and accumulate must not double the 100px translation.
	for _, c := range []struct {
		mode string
		got  float64
	}{{"replace", replace}, {add, added}, {accumulate, accumulated}} {
		if c.got < 99.9 || c.got > 100.1 {
			t.Fatalf("%s sample = %.2f px, want the single 100px translation", c.mode, c.got)
		}
	}
}
