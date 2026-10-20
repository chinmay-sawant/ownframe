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
	if replace < 99.9 || replace > 100.1 || added < 199.9 || added > 200.1 || accumulated < 199.9 || accumulated > 200.1 {
		t.Fatalf("replace/add/accumulate samples = %.2f/%.2f/%.2f px, want 100/200/200", replace, added, accumulated)
	}
}
