package cssanimation

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// svgBindDisplay is one inline SVG box over an image operation. The fill op
// before the raster stands for the element's background, so the test can check
// that an SVG paint player leaves the element's own paint alone.
func svgBindDisplay(rules []layout.KeyframesRule, name string) *layout.Display {
	return &layout.Display{
		Width: 24, Height: 24, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, W: 24, H: 24, R: .1, G: .2, B: .3},
			{ID: 2, Kind: layout.DisplayOpImage, W: 24, H: 24},
		},
		OpOwners: []int{0, 0},
		Boxes: []layout.Box{{
			ID: "icon", Tag: "svg", W: 24, H: 24,
			OperationIDs: []uint64{1, 2},
			Animations: []layout.Animation{{
				Name: name, DurationMS: 1000, IterationCount: 1,
				FillMode: "both", PlayState: "running",
			}},
		}},
		Keyframes: rules,
	}
}

// TestBindBoxCreatesAPlayerForSVGPaintOnly pins the bind class an animation
// whose only declarations are inline-SVG paint needs: the player targets the
// SVG root's image operation with SVGPaintMask, which writes no display field.
func TestBindBoxCreatesAPlayerForSVGPaintOnly(t *testing.T) {
	d := svgBindDisplay(svgRule("paint", "fill", "#000000", "#ffffff"), "paint")

	bounds, gaps, err := BindBox(d, d.Boxes[0], 0, svgStart())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 {
		t.Fatalf("bound players = %d, want the SVG paint player", len(bounds))
	}
	if got := bounds[0].Mask; got != SVGPaintMask {
		t.Fatalf("mask = %d, want SVGPaintMask", got)
	}
	if len(bounds[0].Indices) != 1 || bounds[0].Indices[0] != 1 {
		t.Fatalf("targets = %v, want the image operation", bounds[0].Indices)
	}
	for _, gap := range gaps {
		if strings.Contains(gap, "phase 11") {
			t.Fatalf("SVG paint is still reported deferred: %q", gap)
		}
	}

	bounds[0].Player.Step(svgStart().Add(500 * time.Millisecond))
	values := bounds[0].Player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d, want one", len(values))
	}
	svgNear(t, "fill red", values[0].Fill.R, .5)

	if d.Ops[1].R != 0 || d.Ops[1].W != 24 {
		t.Fatalf("the SVG player wrote the image operation: %+v", d.Ops[1])
	}
	if d.Ops[0].R != .1 {
		t.Fatalf("the SVG player wrote the element's fill operation: %+v", d.Ops[0])
	}
}

// TestBindBoxKeepsSVGPaintOnAMixedPlayer pins the single-player rule: a rule
// that also declares a paint class carries the SVG sample on that player, so
// one element never produces two SVG samples.
func TestBindBoxKeepsSVGPaintOnAMixedPlayer(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "both", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "background-color", Value: "#000000"},
			{Property: "fill", Value: "#000000"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "background-color", Value: "#ffffff"},
			{Property: "fill", Value: "#ffffff"},
		}},
	}}}
	d := svgBindDisplay(rules, "both")

	bounds, _, err := BindBox(d, d.Boxes[0], 0, svgStart())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 {
		t.Fatalf("bound players = %d, want the fill player only", len(bounds))
	}
	if got := bounds[0].Mask; got != cssmotion.MaskRGB {
		t.Fatalf("mask = %d, want the colour field", got)
	}
	bounds[0].Player.Step(svgStart().Add(500 * time.Millisecond))
	if values := bounds[0].Player.SVGPaintValues(); len(values) == 0 {
		t.Fatal("the mixed player dropped its SVG sample")
	}
}

// TestBindBoxReportsDeferredSVGPaintServer pins the gap counter for a paint
// server: the declaration stays out of the sample and the element keeps its
// base paint, but the bind says why.
func TestBindBoxReportsDeferredSVGPaintServer(t *testing.T) {
	d := svgBindDisplay(svgRule("paint", "fill", "url(#gradient)", "url(#gradient)"), "paint")

	bounds, gaps, err := BindBox(d, d.Boxes[0], 0, svgStart())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 {
		t.Fatalf("bound players = %d, want one", len(bounds))
	}
	found := false
	for _, gap := range gaps {
		if strings.Contains(gap, "url(#gradient)") {
			found = true
		}
	}
	if !found {
		t.Fatalf("gaps = %v, want the paint server named", gaps)
	}

	bounds[0].Player.Step(svgStart().Add(500 * time.Millisecond))
	values := bounds[0].Player.SVGPaintValues()
	if len(values) != 1 || values[0].HasFill {
		t.Fatalf("values = %+v, want no sampled fill", values)
	}
	if values[0].Deferred == "" {
		t.Fatal("the deferred reason is missing from the sample")
	}
}

// TestBindBoxRefusesSVGPaintOnANonSVGBox pins the element gate: fill and stroke
// only restyle an inline svg raster, so a div with an image operation is
// reported instead of sent to a restyle the engine would reject.
func TestBindBoxRefusesSVGPaintOnANonSVGBox(t *testing.T) {
	d := svgBindDisplay(svgRule("paint", "fill", "#000000", "#ffffff"), "paint")
	d.Boxes[0].Tag = "div"

	bounds, gaps, err := BindBox(d, d.Boxes[0], 0, svgStart())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 0 {
		t.Fatalf("bound players = %d, want none", len(bounds))
	}
	if len(gaps) == 0 {
		t.Fatal("a non-SVG box reported no gap")
	}
}

// TestBindBoxRefusesSVGPaintWithoutAnID pins the restyle precondition: the
// engine finds the element by id, so an anonymous box reports a gap instead of
// binding a player that can never apply.
func TestBindBoxRefusesSVGPaintWithoutAnID(t *testing.T) {
	d := svgBindDisplay(svgRule("paint", "fill", "#000000", "#ffffff"), "paint")
	d.Boxes[0].ID = ""

	bounds, gaps, err := BindBox(d, d.Boxes[0], -1, svgStart())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 0 {
		t.Fatalf("bound players = %d, want none", len(bounds))
	}
	if len(gaps) == 0 {
		t.Fatal("an anonymous SVG box reported no gap")
	}
}
