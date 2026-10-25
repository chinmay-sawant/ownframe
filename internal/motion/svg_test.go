package motion

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// svgDisplay is one inline SVG box over a single image operation with a running
// fill animation from black to white.
func svgDisplay(playState string) *layout.Display {
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{{
			ID: testOpID, Kind: layout.DisplayOpImage,
			X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
		}},
		OpOwners: []int{0},
		Boxes: []layout.Box{{
			ID: testChipID, Tag: "svg",
			X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
			OperationIDs: []uint64{testOpID},
			Animations: []layout.Animation{{
				Name: "paint", DurationMS: testDurationMS, IterationCount: 1,
				IterationInfinite: true, PlayState: playState, FillMode: "both",
			}},
		}},
		Keyframes: []layout.KeyframesRule{{Name: "paint", Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: "fill", Value: "#000000"},
			}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: "fill", Value: "#ffffff"},
			}},
		}}},
	}
}

// svgTwoDisplay puts the same paint animation on two SVG boxes.
func svgTwoDisplay() *layout.Display {
	d := svgDisplay("running")
	second := d.Boxes[0]
	second.ID = "chip2"
	second.OperationIDs = []uint64{testOpID + 1}
	d.Ops = append(d.Ops, layout.DisplayOp{
		ID: testOpID + 1, Kind: layout.DisplayOpImage,
		X: testChipX + testShift, Y: testChipY, W: testChipSize, H: testChipSize,
	})
	d.OpOwners = append(d.OpOwners, 1)
	d.Boxes = append(d.Boxes, second)

	return d
}

func boundSVG(t *testing.T, d *layout.Display) *Runtime {
	t.Helper()

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	return r
}

// TestSVGPaintCollectsOneRequestPerElement pins the request shape: one element
// override with the fill serialized, while the image operation keeps its paint.
func TestSVGPaintCollectsOneRequestPerElement(t *testing.T) {
	d := svgDisplay("running")
	r := boundSVG(t, d)

	r.Step(d, testClock().Add(50*time.Millisecond))
	stats := r.Stats()
	if stats.SVGPaintRequests != 1 {
		t.Fatalf("requests = %d, want one", stats.SVGPaintRequests)
	}
	if stats.SVGPaintRestyles != 0 {
		t.Fatalf("restyles = %d before the page applied anything", stats.SVGPaintRestyles)
	}

	requests := r.TakeSVGPaint()
	if len(requests) != 1 {
		t.Fatalf("taken requests = %d, want one", len(requests))
	}
	request := requests[0]
	if request.ElementID != testChipID {
		t.Fatalf("element id = %q, want %q", request.ElementID, testChipID)
	}
	if len(request.Overrides) != 1 || !request.Overrides[0].HasFill {
		t.Fatalf("overrides = %+v, want one fill override", request.Overrides)
	}
	if got, want := request.Overrides[0].Fill, "rgb(128, 128, 128)"; got != want {
		t.Fatalf("fill = %q, want %q", got, want)
	}

	r.NoteSVGPaint(1, nil)
	if got := r.Stats().SVGPaintRestyles; got != 1 {
		t.Fatalf("restyles after the note = %d, want one", got)
	}

	// The operation is not a paint target: the SVG sample never writes it.
	if d.Ops[0].R != 0 || d.Ops[0].W != testChipSize || d.Ops[0].Alpha != 0 {
		t.Fatalf("the image operation changed: %+v", d.Ops[0])
	}

	// A repeated sample of the same paint produces no request.
	r.Step(d, testClock().Add(50*time.Millisecond))
	if got := r.Stats().SVGPaintRequests; got != 0 {
		t.Fatalf("requests for an unchanged sample = %d, want none", got)
	}
}

// TestSVGPaintSerializesNoneAndDashes pins the engine-facing strings: none is
// the keyword, not empty, and dash lengths carry the point unit the sampled
// value already uses.
func TestSVGPaintSerializesNoneAndDashes(t *testing.T) {
	override, ok := svgPaintOverride(animation.SVGPaintValue{
		HasFill:             true,
		Fill:                animation.SVGPaintColor{None: true, A: 1},
		HasStroke:           true,
		Stroke:              animation.SVGPaintColor{R: 1, G: .5, B: 0, A: .5},
		HasStrokeWidth:      true,
		StrokeWidth:         3,
		HasStrokeDashArray:  true,
		StrokeDashArray:     []float64{2, 4.5},
		HasStrokeDashOffset: true,
		StrokeDashOffset:    -4,
	}, testChipID)
	if !ok {
		t.Fatal("a declared paint produced no override")
	}
	if override.Fill != "none" {
		t.Fatalf("fill = %q, want none", override.Fill)
	}
	if override.Stroke != "rgba(255, 128, 0, 0.5)" {
		t.Fatalf("stroke = %q, want the rgba form", override.Stroke)
	}
	if override.StrokeDashArray != "2pt 4.5pt" {
		t.Fatalf("dash array = %q, want point lengths", override.StrokeDashArray)
	}
	if override.StrokeWidth != 3 || override.StrokeDashOffset != -4 {
		t.Fatalf("lengths = %v/%v, want the sampled values", override.StrokeWidth, override.StrokeDashOffset)
	}

	empty, ok := svgPaintOverride(animation.SVGPaintValue{
		HasStrokeDashArray: true,
		StrokeDashArray:    nil,
	}, testChipID)
	if !ok || empty.StrokeDashArray != "none" {
		t.Fatalf("empty dash list = %q, want none", empty.StrokeDashArray)
	}

	if _, ok := svgPaintOverride(animation.SVGPaintValue{Deferred: "fill: url(#gradient)"},
		testChipID); ok {
		t.Fatal("a fully deferred sample produced an override")
	}
}

// TestSVGPaintDeferredServerCountsAndLeavesTheBasePaint pins the paint-server
// path: the bind gap names it, the sample carries the reason, and no restyle
// request is collected.
func TestSVGPaintDeferredServerCountsAndLeavesTheBasePaint(t *testing.T) {
	d := svgDisplay("running")
	d.Keyframes[0].Keyframes[0].Declarations[0] = layout.KeyframeDeclaration{
		Property: "fill", Value: "url(#gradient)",
	}
	d.Keyframes[0].Keyframes[1].Declarations[0] = layout.KeyframeDeclaration{
		Property: "fill", Value: "url(#gradient)",
	}
	r := boundSVG(t, d)

	bindStats := r.Stats()
	found := false
	for _, gap := range bindStats.Gaps {
		if strings.Contains(gap, "url(#gradient)") {
			found = true
		}
	}
	if !found {
		t.Fatalf("bind gaps = %v, want the paint server named", bindStats.Gaps)
	}

	r.Step(d, testClock().Add(50*time.Millisecond))
	stats := r.Stats()
	if stats.SVGPaintRequests != 0 {
		t.Fatalf("requests = %d, want none for a deferred paint", stats.SVGPaintRequests)
	}
	if stats.SVGPaintDeferred != 1 {
		t.Fatalf("deferred = %d, want one", stats.SVGPaintDeferred)
	}
	if !strings.Contains(stats.SVGPaintDeferredReason, "url(#gradient)") {
		t.Fatalf("deferred reason = %q, want the paint server", stats.SVGPaintDeferredReason)
	}
	if got := r.TakeSVGPaint(); len(got) != 0 {
		t.Fatalf("taken requests = %d, want none", len(got))
	}
	if d.Ops[0].R != 0 {
		t.Fatalf("the deferred paint wrote the operation: %+v", d.Ops[0])
	}
}

// TestSVGPaintBudgetRotatesTheSecondElement pins the one-per-frame budget: the
// first take applies one element, the second take reaches the other instead of
// starving it behind the busiest element.
func TestSVGPaintBudgetRotatesTheSecondElement(t *testing.T) {
	d := svgTwoDisplay()
	r := boundSVG(t, d)

	r.Step(d, testClock().Add(20*time.Millisecond))
	stats := r.Stats()
	if stats.SVGPaintRequests != 2 {
		t.Fatalf("requests = %d, want both elements", stats.SVGPaintRequests)
	}
	first := r.TakeSVGPaint()
	if len(first) != 1 || first[0].ElementID != testChipID {
		t.Fatalf("first take = %+v, want the first element", first)
	}
	if stats := r.Stats(); !stats.SVGPaintBudgetHit || stats.SVGPaintQueued != 1 {
		t.Fatalf("stats after the first take = %+v, want one queued and a budget hit", stats)
	}
	r.NoteSVGPaint(1, nil)

	r.Step(d, testClock().Add(40*time.Millisecond))
	second := r.TakeSVGPaint()
	if len(second) != 1 || second[0].ElementID != "chip2" {
		t.Fatalf("second take = %+v, want the second element", second)
	}
	if stats := r.Stats(); !stats.SVGPaintBudgetHit || stats.SVGPaintQueued != 1 {
		t.Fatalf("stats after the second take = %+v, want the other element queued", stats)
	}
}

// TestSVGPaintEntryNeverPromotesToALayer pins the layer rule: paint restyling
// is a raster rebuild, so the entry is not a compositor promotion candidate.
func TestSVGPaintEntryNeverPromotesToALayer(t *testing.T) {
	d := svgDisplay("running")
	r := boundSVG(t, d)

	r.Step(d, testClock().Add(50*time.Millisecond))
	if entries := r.LayerEntries(); len(entries) != 0 {
		t.Fatalf("layer entries = %+v, want none for SVG paint", entries)
	}
}

// TestSVGPaintPausedRequestSurvivesADeferredFrame pins the retained-state
// retry: a paused player is normally sampled once and parked, but its held SVG
// sample has to be collected again until the page can apply it.
func TestSVGPaintPausedRequestSurvivesADeferredFrame(t *testing.T) {
	d := svgDisplay("paused")
	r := boundSVG(t, d)

	r.Step(d, testClock().Add(50*time.Millisecond))
	if got := len(r.TakeSVGPaint()); got != 1 {
		t.Fatalf("first take = %d, want the held sample", got)
	}
	// The page could not apply it yet and did not call NoteSVGPaint.
	r.Step(d, testClock().Add(60*time.Millisecond))
	if got := len(r.TakeSVGPaint()); got != 1 {
		t.Fatalf("retry take = %d, want the same held sample", got)
	}
}

// TestASecondDimensionalPageCollectsNoSVGPaint pins the guard: a translate
// animation has no SVG sample, so the restyle path is never reached.
func TestASecondDimensionalPageCollectsNoSVGPaint(t *testing.T) {
	d := animatedDisplay("running", true)
	r := boundSVG(t, d)

	damage := r.Step(d, testClock().Add(50*time.Millisecond))
	if damage.Empty() {
		t.Fatal("the 2D fixture produced no damage")
	}
	stats := r.Stats()
	if stats.SVGPaintRequests != 0 || stats.SVGPaintRestyles != 0 {
		t.Fatalf("stats = %+v, want no SVG request", stats)
	}
	if got := r.TakeSVGPaint(); len(got) != 0 {
		t.Fatalf("taken requests = %d, want none", len(got))
	}
}
