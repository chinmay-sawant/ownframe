package scene

import (
	"testing"
	"time"
)

// fallbackSrc is a template the replay cannot accept: the transform makes
// the engine paint a bitmap instead of a display list.
const fallbackSrc = `<!DOCTYPE html><html><head><meta charset="utf-8">
<!-- stylesheet --></head><body>
<div style="position:relative;width:100%;height:100%">
<div id="t-score" style="transform:rotate(10deg)">000000000</div>
</div></body></html>`

func TestBitmapFallbackRepaints(t *testing.T) {
	m := &fakeModel{}
	s, clock := newTestSceneHTML(t, m, nil, Options{Stepper: &fakeStepper{}}, pageHTML(fallbackSrc))

	if s.page.Testing().Display() != nil {
		t.Skip("the engine kept a display list for the fallback page")
	}

	gen := s.page.Testing().Generation()
	m.f.Score = 5
	tickAt(t, s, clock, time.Second/60)

	if s.page.Testing().Generation() == gen {
		t.Fatal("the bitmap path did not repaint a changed frame")
	}

	if s.page.Testing().Image() == nil {
		t.Fatal("the bitmap path has no image")
	}

	gen = s.page.Testing().Generation()
	m.f.Score = 6
	tickAt(t, s, clock, 10*time.Millisecond)

	if s.page.Testing().Generation() != gen {
		t.Fatal("the bitmap throttle did not hold")
	}

	tickAt(t, s, clock, 60*time.Millisecond)

	if s.page.Testing().Generation() == gen {
		t.Fatal("the bitmap throttle never released")
	}
}

func TestBitmapFallbackSkipsUnchangedFrames(t *testing.T) {
	m := &fakeModel{}
	s, clock := newTestSceneHTML(t, m, nil, Options{Stepper: &fakeStepper{}}, pageHTML(fallbackSrc))

	if s.page.Testing().Display() != nil {
		t.Skip("the engine kept a display list for the fallback page")
	}

	tickAt(t, s, clock, time.Second/60)
	gen := s.page.Testing().Generation()

	tickAt(t, s, clock, 100*time.Millisecond)

	if s.page.Testing().Generation() != gen {
		t.Fatal("an unchanged bitmap frame was repainted")
	}
}
