package motion

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"slices"
	"strings"
	"testing"
	"time"
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
	override, ok := svgPaintOverride(SVGPaintValue{
		HasFill:             true,
		Fill:                SVGPaintColor{None: true, A: 1},
		HasStroke:           true,
		Stroke:              SVGPaintColor{R: 1, G: .5, B: 0, A: .5},
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

	empty, ok := svgPaintOverride(SVGPaintValue{
		HasStrokeDashArray: true,
		StrokeDashArray:    nil,
	}, testChipID)
	if !ok || empty.StrokeDashArray != "none" {
		t.Fatalf("empty dash list = %q, want none", empty.StrokeDashArray)
	}

	if _, ok := svgPaintOverride(SVGPaintValue{Deferred: "fill: url(#gradient)"},
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

func assertSVGNear(t *testing.T, name string, got, want float64) {
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
	player := svgTestPlayer(t, []Target{{Op: op, From: Read(op)}}, frames)
	stepFraction(t, player, 500*time.Millisecond)

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	if !values[0].HasFill {
		t.Fatal("fill is not reported")
	}
	assertSVGNear(t, "fill red", values[0].Fill.R, .5)
	assertSVGNear(t, "fill green", values[0].Fill.G, .5)
	assertSVGNear(t, "fill blue", values[0].Fill.B, .5)
	assertSVGNear(t, "fill alpha", values[0].Fill.A, 1)
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
	player := svgTestPlayer(t, []Target{{Op: op, From: Read(op)}}, frames)
	stepFraction(t, player, 500*time.Millisecond)

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	assertSVGNear(t, "stroke width", values[0].StrokeWidth, 3)
	assertSVGNear(t, "dash offset", values[0].StrokeDashOffset, 5)
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
	player := svgTestPlayer(t, []Target{{Op: op, From: Read(op)}}, frames)
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
	player := svgTestPlayer(t, []Target{{Op: op, From: Read(op)}}, frames)

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
	player := svgTestPlayer(t, []Target{{Op: op, From: Read(op)}}, frames)

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
	assertSVGNear(t, "fill red", values[0].Fill.R, 1)
}

func TestSVGPaintValuesOnePerTarget(t *testing.T) {
	first, second := svgOp(), svgOp()
	frames := []Keyframe{
		{Offset: 0, SVG: SVGPaintValue{StrokeWidth: 1, HasStrokeWidth: true}},
		{Offset: 1, SVG: SVGPaintValue{StrokeWidth: 3, HasStrokeWidth: true}},
	}
	player := svgTestPlayer(t, []Target{
		{Op: first, From: Read(first)},
		{Op: second, From: Read(second)},
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
	player := svgTestPlayer(t, []Target{{Op: op, From: Read(op)}}, []Keyframe{
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
	player := svgTestPlayer(t, []Target{{Op: op, From: Read(op)}}, frames)
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
	if got := bounds[0].Mask; got != MaskRGB {
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

const svgParseTolerance = 1e-9

// svgStart is the epoch every SVG paint bind starts at.
func svgStart() time.Time { return time.Unix(300, 0) }

func svgNear(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > svgParseTolerance {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// svgRule builds a two-stop rule that declares one property at each edge.
func svgRule(name, property, from, to string) []layout.KeyframesRule {
	return []layout.KeyframesRule{{Name: name, Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: from}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: to}}},
	}}}
}

// bindSVG binds one box with one operation so the parser and the player share
// the same fixture.
func bindSVG(t *testing.T, rules []layout.KeyframesRule, name string) (*layout.Display, *Player) {
	t.Helper()
	d := &layout.Display{
		Width: 24, Height: 24, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, W: 24, H: 24, R: 1, Alpha: 1}},
		Boxes: []layout.Box{{ID: "icon", W: 24, H: 24,
			Animations: []layout.Animation{{
				Name: name, DurationMS: 1000, IterationCount: 1,
				FillMode: "both", PlayState: "running",
			}},
		}},
		Keyframes: rules,
	}
	initial := Read(&d.Ops[0])
	player, err := BindAnimation(d, d.Boxes[0], []int{0}, []Paint{initial}, svgStart())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}

	return d, player
}

func TestParseSVGPaintKeyframes(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "icon", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "fill", Value: "#000000"},
			{Property: "fill-opacity", Value: "0.2"},
			{Property: "stroke", Value: "rgb(18, 52, 86)"},
			{Property: "stroke-opacity", Value: "40%"},
			{Property: "stroke-width", Value: "1px"},
			{Property: "stroke-dasharray", Value: "2, 4"},
			{Property: "stroke-dashoffset", Value: "0px"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "fill", Value: "#ffffff"},
			{Property: "fill-opacity", Value: "0.8"},
			{Property: "stroke", Value: "red"},
			{Property: "stroke-opacity", Value: "0.9"},
			{Property: "stroke-width", Value: "5px"},
			{Property: "stroke-dasharray", Value: "4 8"},
			{Property: "stroke-dashoffset", Value: "10px"},
		}},
	}}}

	frames := keyframes(rules, "icon", 1, 24, 24)
	if len(frames) != 2 {
		t.Fatalf("frames = %d, want two", len(frames))
	}
	for i := range frames {
		if frames[i].Err != nil {
			t.Fatalf("frame %d error: %v", i, frames[i].Err)
		}
	}

	first := frames[0].SVG
	if !first.HasFill || first.Fill.R != 0 || first.Fill.G != 0 || first.Fill.B != 0 {
		t.Fatalf("first fill = %+v, want #000000", first)
	}
	if !first.HasFillOpacity {
		t.Fatal("fill-opacity is missing")
	}
	svgNear(t, "fill-opacity", first.FillOpacity, .2)
	if !first.HasStroke {
		t.Fatal("stroke is missing")
	}
	svgNear(t, "stroke red", first.Stroke.R, 18.0/255)
	svgNear(t, "stroke green", first.Stroke.G, 52.0/255)
	svgNear(t, "stroke blue", first.Stroke.B, 86.0/255)
	svgNear(t, "stroke-opacity", first.StrokeOpacity, .4)
	if !first.HasStrokeWidth {
		t.Fatal("stroke-width is missing")
	}
	svgNear(t, "stroke-width", first.StrokeWidth, 1)
	if !first.HasStrokeDashArray || len(first.StrokeDashArray) != 2 {
		t.Fatalf("dash array = %v, want two lengths", first.StrokeDashArray)
	}
	svgNear(t, "dash first", first.StrokeDashArray[0], 2)
	svgNear(t, "dash second", first.StrokeDashArray[1], 4)
	if !first.HasStrokeDashOffset {
		t.Fatal("stroke-dashoffset is missing")
	}
	svgNear(t, "dash offset", first.StrokeDashOffset, 0)

	last := frames[1].SVG
	if !last.HasFill || last.Fill.R != 1 || last.Fill.G != 1 || last.Fill.B != 1 {
		t.Fatalf("last fill = %+v, want #ffffff", last)
	}
	svgNear(t, "last stroke-width", last.StrokeWidth, 5)
	svgNear(t, "last dash offset", last.StrokeDashOffset, 10)
}

func TestSVGPaintLengthsUseLayoutPoints(t *testing.T) {
	frames := keyframes(svgRule("icon", "stroke-width", "8px", "8px"), "icon", .75, 24, 24)

	svgNear(t, "stroke width in points", frames[0].SVG.StrokeWidth, 6)
}

func TestSVGPaintNoneIsPreserved(t *testing.T) {
	frames := keyframes(svgRule("icon", "fill", "none", "#ff0000"), "icon", 1, 24, 24)

	if !frames[0].SVG.HasFill || !frames[0].SVG.Fill.None {
		t.Fatalf("fill none = %+v, want a none fill", frames[0].SVG)
	}
	if frames[1].SVG.Fill.None {
		t.Fatal("red parsed as none")
	}
}

func TestSVGPaintFillMidpointFromBind(t *testing.T) {
	d, player := bindSVG(t, svgRule("paint", "fill", "#000000", "#ffffff"), "paint")
	player.Step(svgStart().Add(500 * time.Millisecond))

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	svgNear(t, "fill red", values[0].Fill.R, .5)
	svgNear(t, "fill green", values[0].Fill.G, .5)
	svgNear(t, "fill blue", values[0].Fill.B, .5)
	if d.Ops[0].R != 1 || d.Ops[0].G != 0 || d.Ops[0].B != 0 {
		t.Fatalf("operation fill = %v,%v,%v, want the captured red left alone",
			d.Ops[0].R, d.Ops[0].G, d.Ops[0].B)
	}
}

func TestSVGPaintStrokeWidthAndDashOffsetFromBind(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "dash", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "stroke-width", Value: "1px"},
			{Property: "stroke-dashoffset", Value: "0px"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "stroke-width", Value: "5px"},
			{Property: "stroke-dashoffset", Value: "10px"},
		}},
	}}}
	_, player := bindSVG(t, rules, "dash")
	player.Step(svgStart().Add(500 * time.Millisecond))

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	svgNear(t, "stroke width", values[0].StrokeWidth, 3)
	svgNear(t, "dash offset", values[0].StrokeDashOffset, 5)
}

func TestSVGPaintDashArrayMismatchFromBind(t *testing.T) {
	_, player := bindSVG(t, svgRule("dash", "stroke-dasharray", "2 4", "4"), "dash")

	player.Step(svgStart().Add(250 * time.Millisecond))
	values := player.SVGPaintValues()
	if len(values) != 1 || len(values[0].StrokeDashArray) != 2 {
		t.Fatalf("dash array before the midpoint = %+v, want the two-length list", values)
	}
	svgNear(t, "dash first", values[0].StrokeDashArray[0], 2)
	svgNear(t, "dash second", values[0].StrokeDashArray[1], 4)

	player.Step(svgStart().Add(750 * time.Millisecond))
	values = player.SVGPaintValues()
	if len(values) != 1 || len(values[0].StrokeDashArray) != 1 {
		t.Fatalf("dash array after the midpoint = %+v, want the one-length list", values)
	}
	svgNear(t, "dash after", values[0].StrokeDashArray[0], 4)
}

func TestSVGPaintNoneToColourFromBind(t *testing.T) {
	_, player := bindSVG(t, svgRule("paint", "fill", "none", "#ff0000"), "paint")

	player.Step(svgStart().Add(250 * time.Millisecond))
	values := player.SVGPaintValues()
	if len(values) != 1 || !values[0].Fill.None {
		t.Fatalf("fill before the midpoint = %+v, want none", values)
	}

	player.Step(svgStart().Add(750 * time.Millisecond))
	values = player.SVGPaintValues()
	if len(values) != 1 || values[0].Fill.None {
		t.Fatalf("fill after the midpoint = %+v, want red", values)
	}
	svgNear(t, "fill red", values[0].Fill.R, 1)
}

func TestSVGPaintServerIsDeferred(t *testing.T) {
	if _, err := svgColor("url(#gradient)", Color{}, false); !errors.Is(err, ErrPaintServerDeferred) {
		t.Fatalf("svgColor error = %v, want ErrPaintServerDeferred", err)
	}

	frames := keyframes(svgRule("paint", "fill", "url(#gradient)", "#ff0000"), "paint", 1, 24, 24)
	if frames[0].Err != nil {
		t.Fatalf("url fill failed the bind: %v", frames[0].Err)
	}
	if frames[0].SVG.HasFill {
		t.Fatal("url fill guessed a colour")
	}
	if !strings.Contains(frames[0].SVG.Deferred, "url(#gradient)") {
		t.Fatalf("deferred reason = %q, want the paint server named", frames[0].SVG.Deferred)
	}
}

func TestSVGPaintWithTransformSamplesBoth(t *testing.T) {
	rules := []layout.KeyframesRule{{Name: "move", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "transform", Value: "translate(0px)"},
			{Property: "fill", Value: "#000000"},
		}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "transform", Value: "translate(100px)"},
			{Property: "fill", Value: "#ffffff"},
		}},
	}}}
	_, player := bindSVG(t, rules, "move")
	player.Step(svgStart().Add(500 * time.Millisecond))

	values := player.SVGPaintValues()
	if len(values) != 1 {
		t.Fatalf("SVGPaintValues = %d values, want one", len(values))
	}
	svgNear(t, "fill red", values[0].Fill.R, .5)

	current, _, ok := player.TargetPaint(0)
	if !ok {
		t.Fatal("target paint is missing")
	}
	if !current.HasMatrix {
		t.Fatal("transform did not write a matrix")
	}
	svgNear(t, "matrix translate x", current.Matrix.E, 50)
}
