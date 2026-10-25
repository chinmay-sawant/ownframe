package animation

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const svgTolerance = 1e-9

// svgEpoch is the clock every SVG paint test starts at.
func svgEpoch() time.Time { return time.Unix(200, 0) }

// svgTestPlayer builds and starts a one-effect player at the test epoch.
func svgTestPlayer(t *testing.T, targets []Target, frames []Keyframe) *Player {
	t.Helper()
	player, err := NewComposite(targets, []Effect{{
		Frames: frames,
		Timing: Timing{Duration: time.Second, Iterations: 1},
	}})
	if err != nil {
		t.Fatalf("NewComposite: %v", err)
	}
	player.Start(svgEpoch())

	return player
}

// stepFraction samples the player at a fraction of its one-second duration.
func stepFraction(t *testing.T, player *Player, at time.Duration) {
	t.Helper()
	player.Step(svgEpoch().Add(at))
}

// svgOp is one opaque red fill operation the SVG samples never write.
func svgOp() *layout.DisplayOp {
	return &layout.DisplayOp{Kind: layout.DisplayOpFillRect, W: 20, H: 10, R: 1, Alpha: 1}
}

func near(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > svgTolerance {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func TestSVGPaintFillInterpolatesInSRGB(t *testing.T) {
	op := svgOp()
	frames := []Keyframe{
		{Offset: 0, SVG: SVGPaintValue{Fill: SVGPaintColor{A: 1}, HasFill: true}},
		{Offset: 1, SVG: SVGPaintValue{
			Fill: SVGPaintColor{R: 1, G: 1, B: 1, A: 1}, HasFill: true,
		}},
	}
	player := svgTestPlayer(t, []Target{{Op: op, From: cssmotion.Read(op)}}, frames)
	stepFraction(t, player, 500*time.Millisecond)

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	if !values[0].HasFill {
		t.Fatal("fill is not reported")
	}
	near(t, "fill red", values[0].Fill.R, .5)
	near(t, "fill green", values[0].Fill.G, .5)
	near(t, "fill blue", values[0].Fill.B, .5)
	near(t, "fill alpha", values[0].Fill.A, 1)
	if op.R != 1 || op.G != 0 || op.B != 0 {
		t.Fatalf("operation fill = %v,%v,%v, want the captured red left alone",
			op.R, op.G, op.B)
	}
}

func TestSVGPaintWidthAndDashOffsetInterpolate(t *testing.T) {
	op := svgOp()
	short := SVGPaintValue{StrokeWidth: 1, HasStrokeWidth: true, HasStrokeDashOffset: true}
	full := SVGPaintValue{StrokeWidth: 5, StrokeDashOffset: 10,
		HasStrokeWidth: true, HasStrokeDashOffset: true}
	frames := []Keyframe{
		{Offset: 0, SVG: short},
		{Offset: 1, SVG: full},
	}
	player := svgTestPlayer(t, []Target{{Op: op, From: cssmotion.Read(op)}}, frames)
	stepFraction(t, player, 500*time.Millisecond)

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	near(t, "stroke width", values[0].StrokeWidth, 3)
	near(t, "dash offset", values[0].StrokeDashOffset, 5)
	if !values[0].HasStrokeWidth || !values[0].HasStrokeDashOffset {
		t.Fatal("stroke width or dash offset is missing")
	}
}

func TestSVGPaintDashArrayMatchesComponentWise(t *testing.T) {
	op := svgOp()
	frames := []Keyframe{
		{Offset: 0, SVG: SVGPaintValue{StrokeDashArray: []float64{2, 4}, HasStrokeDashArray: true}},
		{Offset: 1, SVG: SVGPaintValue{StrokeDashArray: []float64{4, 8}, HasStrokeDashArray: true}},
	}
	player := svgTestPlayer(t, []Target{{Op: op, From: cssmotion.Read(op)}}, frames)
	stepFraction(t, player, 500*time.Millisecond)

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	if !slices.Equal(values[0].StrokeDashArray, []float64{3, 6}) {
		t.Fatalf("dash array = %v, want [3 6]", values[0].StrokeDashArray)
	}
}

func TestSVGPaintDashArrayMismatchedCountIsDiscrete(t *testing.T) {
	op := svgOp()
	frames := []Keyframe{
		{Offset: 0, SVG: SVGPaintValue{StrokeDashArray: []float64{2}, HasStrokeDashArray: true}},
		{Offset: 1, SVG: SVGPaintValue{StrokeDashArray: []float64{2, 4}, HasStrokeDashArray: true}},
	}
	player := svgTestPlayer(t, []Target{{Op: op, From: cssmotion.Read(op)}}, frames)

	stepFraction(t, player, 250*time.Millisecond)
	values := player.SVGPaintValues()
	if len(values) != 1 || !slices.Equal(values[0].StrokeDashArray, []float64{2}) {
		t.Fatalf("dash array before the midpoint = %v, want [2]", values)
	}

	stepFraction(t, player, 750*time.Millisecond)
	values = player.SVGPaintValues()
	if len(values) != 1 || !slices.Equal(values[0].StrokeDashArray, []float64{2, 4}) {
		t.Fatalf("dash array after the midpoint = %v, want [2 4]", values)
	}
}

func TestSVGPaintNoneIsDiscrete(t *testing.T) {
	op := svgOp()
	frames := []Keyframe{
		{Offset: 0, SVG: SVGPaintValue{Fill: SVGPaintColor{None: true, A: 1}, HasFill: true}},
		{Offset: 1, SVG: SVGPaintValue{Fill: SVGPaintColor{R: 1, A: 1}, HasFill: true}},
	}
	player := svgTestPlayer(t, []Target{{Op: op, From: cssmotion.Read(op)}}, frames)

	stepFraction(t, player, 250*time.Millisecond)
	values := player.SVGPaintValues()
	if len(values) != 1 || !values[0].Fill.None {
		t.Fatalf("fill before the midpoint = %+v, want none", values)
	}

	stepFraction(t, player, 750*time.Millisecond)
	values = player.SVGPaintValues()
	if len(values) != 1 || values[0].Fill.None {
		t.Fatalf("fill after the midpoint = %+v, want red", values)
	}
	near(t, "fill red", values[0].Fill.R, 1)
}

func TestSVGPaintValuesOnePerTarget(t *testing.T) {
	first, second := svgOp(), svgOp()
	frames := []Keyframe{
		{Offset: 0, SVG: SVGPaintValue{StrokeWidth: 1, HasStrokeWidth: true}},
		{Offset: 1, SVG: SVGPaintValue{StrokeWidth: 3, HasStrokeWidth: true}},
	}
	player := svgTestPlayer(t, []Target{
		{Op: first, From: cssmotion.Read(first)},
		{Op: second, From: cssmotion.Read(second)},
	}, frames)
	stepFraction(t, player, 500*time.Millisecond)

	values := player.SVGPaintValues()
	if len(values) != 2 {
		t.Fatalf("SVGPaintValues = %d values, want one per target", len(values))
	}
	if values[0].Target != 0 || values[1].Target != 1 {
		t.Fatalf("targets = %d,%d, want 0,1", values[0].Target, values[1].Target)
	}
}

func TestSVGPaintValuesNilForPlainAnimation(t *testing.T) {
	op := svgOp()
	player := svgTestPlayer(t, []Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 0, Opacity: 1, OpacitySet: true},
		{Offset: 1, Opacity: .5, OpacitySet: true},
	})
	stepFraction(t, player, 500*time.Millisecond)

	if values := player.SVGPaintValues(); values != nil {
		t.Fatalf("SVGPaintValues = %v, want nil for a plain animation", values)
	}
}

func TestSVGPaintDeferredReasonIsExposed(t *testing.T) {
	op := svgOp()
	frames := []Keyframe{
		{Offset: 0, SVG: SVGPaintValue{Deferred: "fill: url(#gradient) is deferred"}},
		{Offset: 1, SVG: SVGPaintValue{Deferred: "fill: url(#gradient) is deferred"}},
	}
	player := svgTestPlayer(t, []Target{{Op: op, From: cssmotion.Read(op)}}, frames)
	stepFraction(t, player, 500*time.Millisecond)

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	if values[0].Deferred == "" {
		t.Fatal("deferred reason is missing")
	}
	if values[0].HasFill {
		t.Fatal("a deferred fill must not report a sampled colour")
	}
}
