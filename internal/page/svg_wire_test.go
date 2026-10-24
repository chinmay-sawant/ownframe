package page

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const svgFillFrames = `
@keyframes svgpaint {
	0% { fill: #000000; }
	100% { fill: #ffffff; }
}`

const svgFillStyle = `animation-name: svgpaint; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both;`

const svgDashFrames = `
@keyframes svgdash {
	0% { stroke: #1188ff; stroke-width: 1px; stroke-dasharray: 2 2; stroke-dashoffset: 0px; }
	100% { stroke: #ffffff; stroke-width: 8px; stroke-dasharray: 6 2; stroke-dashoffset: 12px; }
}`

const svgDashStyle = `animation-name: svgdash; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both;`

const svgBlockBody = `<svg id="icon" width="24" height="24" viewBox="0 0 24 24" fill="#123456">
	<rect width="24" height="24"/></svg>`

const svgInlineBody = `<p style="margin:0">before <svg id="icon" width="24" height="24"
	viewBox="0 0 24 24" fill="#123456"><rect width="24" height="24"/></svg> after</p>`

const svgDashBody = `<svg id="icon" width="24" height="24" viewBox="0 0 24 24">
	<circle cx="12" cy="12" r="9" fill="none" stroke="#123456" stroke-width="1"/></svg>`

// newSVGPaintPage draws a page whose #icon element carries the given keyframe
// block and animation declarations. The first-frame damage is drained.
func newSVGPaintPage(t *testing.T, frames, style, body string, retain bool) *Page {
	t.Helper()

	source := `<style>` + frames + `
#icon { ` + style + ` }
</style>` + body
	p, err := New(Config{HTML: source, Width: 240, Height: 120, RetainState: retain})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// svgRaster returns the PNG bytes of one element's SVG image operation.
func svgRaster(t *testing.T, p *Page, id string) []byte {
	t.Helper()

	d := p.Display()
	if d == nil {
		t.Fatalf("the page painted no display list")
	}
	box := p.boxByID(id)
	if box.ID != id {
		t.Fatalf("the page has no box %q", id)
	}
	for _, index := range cssmotion.OperationIndices(d, box) {
		if d.Ops[index].Kind != layout.DisplayOpImage {
			continue
		}
		data, _, _ := d.Ops[index].ImageBytes()

		return data
	}
	t.Fatalf("element %q painted no SVG raster", id)

	return nil
}

func TestSVGPaintFillAnimationRerasterizesBlockAndInline(t *testing.T) {
	t.Parallel()

	cases := map[string]string{"block": svgBlockBody, "inline": svgInlineBody}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := newSVGPaintPage(t, svgFillFrames, svgFillStyle, body, true)
			base := bytes.Clone(svgRaster(t, p, "icon"))

			if err := p.Tick(context.Background()); err != nil {
				t.Fatal(err)
			}
			first := bytes.Clone(svgRaster(t, p, "icon"))
			if bytes.Equal(first, base) {
				t.Fatal("the first tick left the base raster alone")
			}

			time.Sleep(30 * time.Millisecond)
			if err := p.Tick(context.Background()); err != nil {
				t.Fatal(err)
			}
			second := bytes.Clone(svgRaster(t, p, "icon"))
			if bytes.Equal(second, first) {
				t.Fatal("the second tick kept the first raster")
			}
			if got := p.Motion().SVGPaintRestyles; got != 1 {
				t.Fatalf("restyles = %d, want one for the frame's element", got)
			}
			if got := p.Motion().SVGPaintError; got != "" {
				t.Fatalf("restyle error = %q, want none", got)
			}
		})
	}
}

func TestSVGPaintStrokeWidthAndDashOffsetAnimate(t *testing.T) {
	t.Parallel()

	p := newSVGPaintPage(t, svgDashFrames, svgDashStyle, svgDashBody, true)
	base := bytes.Clone(svgRaster(t, p, "icon"))

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := bytes.Clone(svgRaster(t, p, "icon"))
	if bytes.Equal(first, base) {
		t.Fatal("stroke paint did not change the raster")
	}

	time.Sleep(30 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if second := svgRaster(t, p, "icon"); bytes.Equal(second, first) {
		t.Fatal("stroke width and dash offset did not advance the raster")
	}
	if got := p.Motion().SVGPaintRequests; got != 1 {
		t.Fatalf("requests = %d, want one", got)
	}
}

func TestSVGPaintWithoutPaintNeverRestyles(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	gen := p.Generation()

	for i := 0; i < 2; i++ {
		if err := p.Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	stats := p.Motion()
	if stats.SVGPaintRequests != 0 || stats.SVGPaintRestyles != 0 {
		t.Fatalf("SVG stats = %+v, want none for a page with no SVG paint", stats)
	}
	if stats.SVGPaintError != "" {
		t.Fatalf("SVG error = %q, want none", stats.SVGPaintError)
	}
	if p.Generation() != gen {
		t.Fatal("a 2D page adopted an SVG patch")
	}
}

func TestSVGPaintDeferredPaintServerLeavesTheBaseRaster(t *testing.T) {
	t.Parallel()

	const frames = `
@keyframes svgpaint {
	0% { fill: url(#gradient); }
	100% { fill: url(#gradient); }
}`
	p := newSVGPaintPage(t, frames, svgFillStyle, svgBlockBody, true)
	base := bytes.Clone(svgRaster(t, p, "icon"))

	for i := 0; i < 2; i++ {
		if err := p.Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	if got := svgRaster(t, p, "icon"); !bytes.Equal(got, base) {
		t.Fatal("a deferred paint server restyled the base raster")
	}
	stats := p.Motion()
	if stats.SVGPaintRestyles != 0 {
		t.Fatalf("restyles = %d, want none for a paint server", stats.SVGPaintRestyles)
	}
	if stats.SVGPaintDeferred == 0 || !strings.Contains(stats.SVGPaintDeferredReason, "url(#gradient)") {
		t.Fatalf("deferred stats = %+v, want the paint server counted", stats)
	}
	found := false
	for _, gap := range stats.Gaps {
		if strings.Contains(gap, "url(#gradient)") {
			found = true
		}
	}
	if !found {
		t.Fatalf("gaps = %v, want the paint server named", stats.Gaps)
	}
}

func TestFirstSVGPaintTickEnablesRetainedStateAndSecondMoves(t *testing.T) {
	t.Parallel()

	p := newSVGPaintPage(t, svgFillFrames, svgFillStyle, svgBlockBody, false)
	if p.retainState {
		t.Fatal("the fixture started with retained layout state")
	}
	base := bytes.Clone(svgRaster(t, p, "icon"))
	gen := p.Generation()

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !p.retainState {
		t.Fatal("the first tick did not turn retained layout state on")
	}
	if p.Generation() == gen {
		t.Fatal("the first tick did not redraw with retained layout state")
	}
	if got := p.Motion().SVGPaintRestyles; got != 0 {
		t.Fatalf("the first tick restyled %d elements, want the owed redraw instead", got)
	}
	if got := svgRaster(t, p, "icon"); !bytes.Equal(got, base) {
		t.Fatal("the first tick moved the raster before retained state landed")
	}
	if p.restampRedraw {
		t.Fatal("the owed redraw stayed queued after the first tick")
	}

	time.Sleep(30 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().SVGPaintRestyles; got != 1 {
		t.Fatalf("the second tick restyled %d elements, want one", got)
	}
	if got := svgRaster(t, p, "icon"); bytes.Equal(got, base) {
		t.Fatal("the second tick kept the base raster")
	}
}

func TestSVGPaintBudgetDefersTheSecondElement(t *testing.T) {
	t.Parallel()

	source := `<style>` + svgFillFrames + `
#icon, #icon2 { ` + svgFillStyle + ` }
</style>` + svgBlockBody + `
<svg id="icon2" width="24" height="24" viewBox="0 0 24 24" fill="#123456">
	<rect width="24" height="24"/></svg>`
	p, err := New(Config{HTML: source, Width: 240, Height: 160, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	base := bytes.Clone(svgRaster(t, p, "icon"))
	otherBase := bytes.Clone(svgRaster(t, p, "icon2"))

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	stats := p.Motion()
	if stats.SVGPaintRestyles != 1 {
		t.Fatalf("restyles = %d, want the one-per-frame budget", stats.SVGPaintRestyles)
	}
	if !stats.SVGPaintBudgetHit || stats.SVGPaintQueued != 1 {
		t.Fatalf("stats = %+v, want the second element deferred", stats)
	}
	if got := svgRaster(t, p, "icon"); bytes.Equal(got, base) {
		t.Fatal("the first element was not restyled")
	}
	if got := svgRaster(t, p, "icon2"); !bytes.Equal(got, otherBase) {
		t.Fatal("the budgeted-out element changed anyway")
	}

	time.Sleep(30 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().SVGPaintRestyles; got != 1 {
		t.Fatalf("second tick restyles = %d, want one", got)
	}
	if got := svgRaster(t, p, "icon2"); bytes.Equal(got, otherBase) {
		t.Fatal("the deferred element never restyled")
	}
}
