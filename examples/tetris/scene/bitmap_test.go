package scene

import (
	"testing"
	"time"
)

// fallbackSrc is a rotated score. The replay painter keeps the display list,
// so this page does not need the bitmap path.
const fallbackSrc = `<!DOCTYPE html><html><head><meta charset="utf-8">
<!-- stylesheet --></head><body>
<div style="position:relative;width:100%;height:100%">
<div id="t-score" style="transform:rotate(10deg)">000000000</div>
</div></body></html>`

// ellipticalSrc is a fill the replay gate rejects. Display stays nil, the
// page keeps a bitmap, and that is the case the window badges.
const ellipticalSrc = `<!DOCTYPE html><html><head><meta charset="utf-8"></head><body>
<div style="width:60px;height:40px;background:#123456;border-radius:50% / 20%"></div>
</body></html>`

func TestBitmapFallbackRepaints(t *testing.T) {
	m := &fakeModel{}
	s, clock := newTestSceneHTML(t, m, nil, Options{Stepper: &fakeStepper{}}, pageHTML(fallbackSrc))

	if s.page.Display() == nil {
		t.Fatal("rotated score fell back to the bitmap")
	}

	if s.page.Image() != nil {
		t.Fatal("rotated score required the bitmap fallback")
	}

	gen := s.page.Generation()
	m.f.Score = 5
	tickAt(t, s, clock, time.Second/60)

	if s.page.Generation() != gen {
		t.Fatal("a replayable score change redrew the page")
	}

	score := textAt(s.page.Display(), s.page.Boxes(), "t-score")
	if score == nil || score.Text != padScore(5) {
		t.Fatalf("score text = %v, want %s", score, padScore(5))
	}
}

func TestBitmapFallbackSkipsUnchangedFrames(t *testing.T) {
	m := &fakeModel{}
	s, clock := newTestSceneHTML(t, m, nil, Options{Stepper: &fakeStepper{}}, pageHTML(fallbackSrc))

	if s.page.Display() == nil {
		t.Fatal("rotated score fell back to the bitmap")
	}

	tickAt(t, s, clock, time.Second/60)
	gen := s.page.Generation()

	tickAt(t, s, clock, 100*time.Millisecond)

	if s.page.Generation() != gen {
		t.Fatal("an unchanged replayable frame was redrawn")
	}

	if s.page.Display() == nil {
		t.Fatal("the tick dropped the display list")
	}
}

func TestBitmapFallbackElliptical(t *testing.T) {
	m := &fakeModel{}
	s, _ := newTestSceneHTML(t, m, nil, Options{Stepper: &fakeStepper{}}, ellipticalSrc)

	if s.page.Display() != nil {
		t.Fatal("elliptical border was replayed")
	}

	if s.page.Image() == nil {
		t.Fatal("no fallback bitmap")
	}

	if len(s.page.PNG()) == 0 {
		t.Fatal("fallback page has no PNG")
	}
}
