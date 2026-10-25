package page

import (
	"context"
	"math"
	"testing"
)

// pageTimelineHTML declares the page scroller under a name and animates one
// target from it. The tall row gives the page a scroll range.
const pageTimelineHTML = `<style>body { margin: 0 }` + motionSlideFrames + `
#page { scroll-timeline: --page y; }
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	animation-name: slide; animation-duration: 1s;
	animation-timing-function: linear; animation-fill-mode: both;
	animation-timeline: --page; }
#tall { height: 500px; }
</style><div id="page"><div id="target"></div><div id="tall"></div></div>`

// newPageTimelinePage draws the named scroll-timeline fixture.
func newPageTimelinePage(t *testing.T) *Page {
	t.Helper()

	p, err := New(Config{HTML: pageTimelineHTML, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// pageScrollSpan is the page scroller range the fixture produces.
func pageScrollSpan(p *Page) float64 {
	content := p.contentRect()

	return float64(content.Max.Y - p.height)
}

// TestPageNamedScrollTimelineFollowsOffset checks the page wiring: a named
// timeline declared on an ancestor resolves from the box metadata, samples
// the window's scroll offset, and paints the right phase.
func TestPageNamedScrollTimelineFollowsOffset(t *testing.T) {
	p := newPageTimelinePage(t)
	index := motionTargetIndex(t, p)
	base := motionTargetX(p, index)

	var scroller Box
	for _, box := range p.Boxes() {
		if box.ID == "page" {
			scroller = box
		}
	}
	if len(scroller.Timelines) == 0 || scroller.Timelines[0].Name != "--page" {
		t.Fatalf("the engine exported no --page timeline on #page: %+v", scroller.Timelines)
	}

	span := pageScrollSpan(p)
	if span <= 0 {
		t.Fatalf("the fixture page does not scroll: span %v", span)
	}
	scale := p.Display().PointsPerPixel
	if scale <= 0 {
		scale = 1
	}
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := (motionTargetX(p, index) - base) / scale; math.Abs(got) > 0.01 {
		t.Fatalf("top sample moved %v pixels, want 0", got)
	}

	p.SetScrollOffset(0, int(span/2))
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := (motionTargetX(p, index) - base) / scale; math.Abs(got-60) > 0.5 {
		t.Fatalf("middle sample moved %v pixels, want 60", got)
	}

	p.SetScrollOffset(0, int(span))
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := (motionTargetX(p, index) - base) / scale; math.Abs(got-120) > 0.5 {
		t.Fatalf("bottom sample moved %v pixels, want 120", got)
	}
}

// TestPageScrollTimelineDoesNotTick checks pacing: a still page with a scroll
// animation samples once at the offset and never asks for another frame.
func TestPageScrollTimelineDoesNotTick(t *testing.T) {
	p := newPageTimelinePage(t)
	p.SetScrollOffset(0, 100)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.NextFrameAt()
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("a still scroll animation asked for a frame at %v", got)
	}
}
