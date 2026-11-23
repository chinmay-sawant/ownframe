package replay

import (
	"bytes"
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"image/png"
	"math"
	"strconv"
	"testing"
)

func TestPopulateRoundedFillPathZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	radii := [4]float64{5, 5, 5, 5}
	r := radius4(radii)

	// Warm path internal slices
	path.Reset()
	populateRoundedFillPath(path, 10, 10, 80, 80, r)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateRoundedFillPath(path, 10, 10, 80, 80, r)
	})

	if allocs != 0 {
		t.Fatalf("populateRoundedFillPath allocated %v objects per run, want 0", allocs)
	}
}

func TestFillRoundedPoolReusesPath(t *testing.T) {
	dst := ebiten.NewImage(100, 100)
	op := &layout.DisplayOp{
		Kind:  layout.DisplayOpFillRect,
		X:     10,
		Y:     10,
		W:     80,
		H:     80,
		R:     1,
		G:     0.5,
		B:     0.2,
		Alpha: 1,
	}
	radii := [4]float64{5, 5, 5, 5}

	// Calling fillRounded must succeed and leave the pool ready for reuse.
	fillRounded(dst, op, radii, 0, 0)
	fillRounded(dst, op, radii, 0, 0)

	borrowed := acquireVectorPath()
	defer releaseVectorPath(borrowed)
	// Borrowed path was Reset before being put back
	if borrowed == nil {
		t.Fatal("expected non-nil path from pool")
	}
}

// cornerPathHTML is a bevel-cornered box with a uniform border. The engine
// emits the background fill and the border as two DisplayOpPath ops on the
// same border box.
func cornerPathHTML(radius int) string {
	return `<body style="margin:0;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:` +
		strconv.Itoa(radius) + `px;corner-shape:bevel"></div></body>`
}

// cornerPathRotatedHTML is a scoop-cornered box under a 15 degree rotation, so
// replay has to map every command point through the op matrix.
const cornerPathRotatedHTML = `<body style="margin:0;background:#ffffff">` +
	`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:scoop;transform:rotate(15deg)"></div></body>`

// pathFillOp returns the first fill path op of a display.
func pathFillOp(t *testing.T, display *layout.Display) *layout.DisplayOp {
	t.Helper()

	op := findPathOp(display, false)
	if op == nil {
		t.Fatal("no fill path op")
	}

	return op
}

// pathStrokeOp returns the first border path op of a display.
func pathStrokeOp(t *testing.T, display *layout.Display) *layout.DisplayOp {
	t.Helper()

	op := findPathOp(display, true)
	if op == nil {
		t.Fatal("no stroke path op")
	}

	return op
}

// TestPathBoundsCoverTheBorderBox pins the shaped box's painted bounds: a fill
// reaches the whole border box, and a border reaches half its width past it,
// each with the one-pixel replay slack.
func TestPathBoundsCoverTheBorderBox(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), cornerPathHTML(24), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	ppt := display.PointsPerPixel

	fill := pathFillOp(t, display)

	box, ok := opBounds(fill, ppt)
	if !ok {
		t.Fatal("a path fill should be bounded")
	}

	if want := image.Rect(19, 19, 149, 109); box != want {
		t.Fatalf("fill bounds %v, want %v", box, want)
	}

	stroke := pathStrokeOp(t, display)

	box, ok = opBounds(stroke, ppt)
	if !ok {
		t.Fatal("a path border should be bounded")
	}

	if want := image.Rect(17, 17, 151, 111); box != want {
		t.Fatalf("border bounds %v, want %v", box, want)
	}
}

// TestPathBoundsFollowTheTransform checks that a transformed path bounds as
// its mapped command points, not as the unmapped box: the box grows, holds
// every mapped corner, and still rejects a rect away from the ink.
func TestPathBoundsFollowTheTransform(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), cornerPathRotatedHTML, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	ppt := display.PointsPerPixel
	op := pathFillOp(t, display)

	box, ok := opBounds(op, ppt)
	if !ok {
		t.Fatal("a transformed path should be bounded")
	}

	// A scoop is concave, so its rotated bounds are smaller than the rotated
	// box's, but a rotation must still push the paint past the upright box on
	// both axes.
	upright := boxBounds(op.X, op.Y, op.W, op.H, 0, ppt)
	if box.Min.X >= upright.Min.X || box.Min.Y >= upright.Min.Y ||
		box.Max.X <= upright.Max.X || box.Max.Y <= upright.Max.Y {
		t.Fatalf("rotated bounds %v did not grow past the upright box %v", box, upright)
	}

	m := op.Transform()
	centerX, centerY := m.Apply(op.X+op.W/2, op.Y+op.H/2)
	center := image.Pt(int(math.Round(centerX/ppt)), int(math.Round(centerY/ppt)))

	if !center.In(box) {
		t.Fatalf("mapped centre %v falls outside bounds %v", center, box)
	}

	if !opTouches(op, ppt, image.Rect(center.X-5, center.Y-5, center.X+5, center.Y+5)) {
		t.Fatal("a rect over the rotated box should touch")
	}

	if opTouches(op, ppt, image.Rect(180, 120, 190, 130)) {
		t.Fatal("a rect away from the rotated box should miss")
	}
}

// TestPathBoundsWithoutCommandsAreEmpty keeps a paint-free payload out of the
// dirty-region walk instead of drawing it everywhere.
func TestPathBoundsWithoutCommandsAreEmpty(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{Kind: layout.DisplayOpPath}

	box, ok := opBounds(op, 0.75)
	if !ok {
		t.Fatal("a path bounds as empty, not unbounded")
	}

	if !box.Empty() {
		t.Fatalf("bounds %v, want empty", box)
	}

	if _, _, _, _, _, present := pathInkBox(op); present {
		t.Fatal("a path without commands should report no ink box")
	}
}

// TestSigPathFoldsTheCommandList checks that two paths with the same box and
// colour but different corner geometry sign differently, while a redraw of the
// same page signs the same. Without this, a cached chunk would keep the old
// shape after a corner mutation.
func TestSigPathFoldsTheCommandList(t *testing.T) {
	t.Parallel()

	sharp, err := render.DisplayList(context.Background(), cornerPathHTML(24), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	repeat, err := render.DisplayList(context.Background(), cornerPathHTML(24), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	blunt, err := render.DisplayList(context.Background(), cornerPathHTML(16), 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	for _, stroke := range []bool{false, true} {
		pick := pathFillOp
		if stroke {
			pick = pathStrokeOp
		}

		a := sigOp(fnvOffsetBasis, pick(t, sharp))
		if again := sigOp(fnvOffsetBasis, pick(t, repeat)); a != again {
			t.Fatalf("stroke=%v: the same path signed differently", stroke)
		}

		if b := sigOp(fnvOffsetBasis, pick(t, blunt)); a == b {
			t.Fatalf("stroke=%v: different corner geometry signed the same", stroke)
		}
	}
}

// TestPathStrokeWidthScalesWithTheMatrix pins the stroke width rule: a
// rotation keeps the width and a uniform scale grows it, so a transformed
// border keeps its weight instead of thinning.
func TestPathStrokeWidthScalesWithTheMatrix(t *testing.T) {
	t.Parallel()

	rotated, err := render.DisplayList(context.Background(), cornerPathRotatedHTML, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	op := pathStrokeOp(t, rotated)
	if want := float32(op.Width * pxPerPt); pathStrokeWidth(op) != want {
		t.Fatalf("rotated width %v, want %v", pathStrokeWidth(op), want)
	}

	scaledPage := `<body style="margin:0;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
		`corner-shape:bevel;transform:scale(2)"></div></body>`

	scaled, err := render.DisplayList(context.Background(), scaledPage, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	op = pathStrokeOp(t, scaled)
	if want := float32(op.Width * pxPerPt * 2); math.Abs(float64(pathStrokeWidth(op)-want)) > 0.01 {
		t.Fatalf("scaled width %v, want %v", pathStrokeWidth(op), want)
	}
}

func TestPopulateCornerPathZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	cmds := []layout.PathCmd{
		{Verb: layout.PathMoveTo, Pts: [3]layout.DisplayPoint{{X: 10, Y: 10}}},
		{Verb: layout.PathLineTo, Pts: [3]layout.DisplayPoint{{X: 50, Y: 10}}},
		{Verb: layout.PathQuadTo, Pts: [3]layout.DisplayPoint{{X: 60, Y: 10}, {X: 60, Y: 20}}},
		{Verb: layout.PathCubicTo, Pts: [3]layout.DisplayPoint{{X: 60, Y: 30}, {X: 50, Y: 40}, {X: 40, Y: 40}}},
		{Verb: layout.PathClose},
	}

	// Warm path internal slices
	path.Reset()
	populateCornerPath(path, cmds)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateCornerPath(path, cmds)
	})

	if allocs != 0 {
		t.Fatalf("populateCornerPath allocated %v objects per run, want 0", allocs)
	}
}

func TestStrokeGeometryZeroRadiiBounds(t *testing.T) {
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    0, Y: 0, W: 30, H: 15, Width: 2,
	}

	path, width, ok := strokeGeometry(op, 3, 7)
	if !ok {
		t.Fatal("zero-radius stroke rejected")
	}

	want := image.Rect(3, 7, 43, 27)
	if got := path.Bounds(); got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}

	if wantWidth := float32(2 * pxPerPt); width != wantWidth {
		t.Fatalf("width = %v, want %v", width, wantWidth)
	}
}

func TestStrokeGeometryCircularRadiusPath(t *testing.T) {
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    10, Y: 5, W: 30, H: 15, Width: 1, Radius: 6,
	}

	path, _, ok := strokeGeometry(op, 0, 0)
	if !ok {
		t.Fatal("circular-radius stroke rejected")
	}

	bounds := path.Bounds()
	if bounds.Empty() {
		t.Fatal("circular-radius path is empty")
	}

	want := image.Rect(13, 6, 54, 27)
	if bounds != want {
		t.Fatalf("bounds = %v, want %v", bounds, want)
	}
}

func TestPopulateRoundedPathZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	rx := [4]float32{6, 6, 6, 6}
	ry := [4]float32{6, 6, 6, 6}

	// Warm path internal slices
	path.Reset()
	populateRoundedPath(path, 10, 5, 40, 20, rx, ry)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateRoundedPath(path, 10, 5, 40, 20, rx, ry)
	})

	if allocs != 0 {
		t.Fatalf("populateRoundedPath allocated %v objects per run, want 0", allocs)
	}
}

func TestPopulateMaskPathZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	// Warm path internal slices
	path.Reset()
	populateMaskPath(path, 10, 5, 40, 20, 15)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateMaskPath(path, 10, 5, 40, 20, 15)
	})

	if allocs != 0 {
		t.Fatalf("populateMaskPath allocated %v objects per run, want 0", allocs)
	}
}

func TestPopulateStrokeGeometryZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    10, Y: 5, W: 30, H: 15, Width: 1, Radius: 6,
	}

	// Warm path internal slices
	path.Reset()
	populateStrokeGeometry(path, op, 0, 0)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateStrokeGeometry(path, op, 0, 0)
	})

	if allocs != 0 {
		t.Fatalf("populateStrokeGeometry allocated %v objects per run, want 0", allocs)
	}
}

func TestStrokeDirectPoolReusesPath(t *testing.T) {
	dst := ebiten.NewImage(100, 100)
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    10, Y: 5, W: 30, H: 15, Width: 1, Radius: 6,
		R: 1, G: 0, B: 0, Alpha: 1,
	}

	// Calling strokeDirect must succeed and leave the pool ready for reuse.
	strokeDirect(dst, op, 0, 0)
	strokeDirect(dst, op, 0, 0)

	borrowed := acquireVectorPath()
	defer releaseVectorPath(borrowed)
	if borrowed == nil {
		t.Fatal("expected non-nil path from pool")
	}
}

func TestDrawSegmentPoolReusesPath(t *testing.T) {
	dst := ebiten.NewImage(100, 100)
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpLine,
		R:    1, G: 0, B: 0, Alpha: 1,
	}

	// Calling drawSegment must succeed and leave the pool ready for reuse.
	drawSegment(dst, op, 10, 10, 50, 50, 2, 0, 0)
	drawSegment(dst, op, 10, 10, 50, 50, 2, 0, 0)

	borrowed := acquireVectorPath()
	defer releaseVectorPath(borrowed)
	if borrowed == nil {
		t.Fatal("expected non-nil path from pool")
	}
}

func TestStrokeGeometryElliptical(t *testing.T) {
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    0, Y: 0, W: 20, H: 10, Width: 1,
		Radius: 4, RadiusY: 2,
	}

	path, _, ok := strokeGeometry(op, 0, 0)
	if !ok {
		t.Fatal("elliptical stroke rejected")
	}

	if path.Bounds().Empty() {
		t.Fatal("elliptical path is empty")
	}
}

func TestStrokeGeometryMasked(t *testing.T) {
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    0, Y: 0, W: 20, H: 10, Width: 1,
		StrokeMask: 1,
	}

	path, _, ok := strokeGeometry(op, 0, 0)
	if !ok {
		t.Fatal("masked stroke rejected")
	}

	if path.Bounds().Dx() <= 0 {
		t.Fatal("masked path has no length")
	}
}

func TestStrokeGeometryUnknownMask(t *testing.T) {
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    0, Y: 0, W: 20, H: 10, Width: 1,
		StrokeMask: 16,
	}

	if _, _, ok := strokeGeometry(op, 0, 0); ok {
		t.Fatal("unknown mask accepted")
	}
}

func TestStrokeGeometryWidthConversion(t *testing.T) {
	cases := []struct {
		name  string
		width float64
		want  float32
	}{
		{name: "points scale", width: 1, want: float32(pxPerPt)},
		{name: "three points", width: 3, want: float32(3 * pxPerPt)},
		{name: "zero clamps to one", width: 0, want: 1},
		{name: "negative clamps to one", width: -2, want: 1},
		{name: "thin clamps to one", width: 0.5, want: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op := &layout.DisplayOp{
				Kind: layout.DisplayOpStrokeRect,
				X:    0, Y: 0, W: 20, H: 10, Width: tc.width,
			}

			_, width, ok := strokeGeometry(op, 0, 0)
			if !ok {
				t.Fatal("stroke rejected")
			}

			if width != tc.want {
				t.Fatalf("width = %v, want %v", width, tc.want)
			}
		})
	}
}

func opaquePNG() ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func imageKind(d *render.Display) *layout.DisplayOp {
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpImage {
			return &d.Ops[i]
		}
	}
	return nil
}

func TestTranslucentImageAgrees(t *testing.T) {
	data, err := opaquePNG()
	if err != nil {
		t.Fatal(err)
	}
	src := `<body style="margin:0"><div style="opacity:.5"><img src="t.png" style="width:20px;height:20px"></div></body>`
	st := render.State{Images: func(s string) ([]byte, error) { return data, nil }}
	d, err := render.DisplayListState(context.Background(), src, 100, 100, st)
	if err != nil {
		t.Fatal(err)
	}
	if masking.ImageAlpha(imageKind(d)) != 0.5 {
		t.Fatal("alpha")
	}
}
func TestTranslucentFillPixel(t *testing.T) {
	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, R: 1, Alpha: 0.5}
	if got := rgba(&op); got != masking.PaintRGBA(&op) || got.A != 128 || got.R != 128 {
		t.Fatalf("rgba %+v", got)
	}
	src := `<body style="margin:0"><div style="width:40px;height:40px;background:red;opacity:.5"></div></body>`
	d, err := render.DisplayList(context.Background(), src, 100, 100)
	if err != nil {
		t.Fatal(err)
	}
	for i := range d.Ops {
		if d.Ops[i].R == 1 && d.Ops[i].G == 0 && d.Ops[i].Kind == layout.DisplayOpFillRect {
			if rgba(&d.Ops[i]).A != 128 {
				t.Fatalf("alpha %+v", rgba(&d.Ops[i]))
			}
			return
		}
	}
	t.Fatal("no red op")
}

func gridRunOp(t *testing.T) (*render.Display, *layout.DisplayOp) {
	t.Helper()
	src := `<body style="margin:0"><table style="border-collapse:collapse"><tr><td style="border:4px solid red;width:40px;height:20px;opacity:.5">x</td></tr></table></body>`
	d, err := render.DisplayList(context.Background(), src, 200, 140)
	if err != nil {
		t.Fatal(err)
	}
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpGridRun && len(d.Ops[i].Grid.Segs) > 0 {
			return d, &d.Ops[i]
		}
	}
	t.Fatal("no gridrun")
	return nil, nil
}

func TestTranslucentGridAgrees(t *testing.T) {
	d, op := gridRunOp(t)
	op.Alpha, op.Fixed = 0.5, true
	line := masking.GridLine(op, 0)
	if line.Grid != nil || line.Alpha != op.Alpha {
		t.Fatalf("line %+v owner %+v", line, op)
	}
	if rgba(&line).A != 128 {
		t.Fatalf("alpha %+v", rgba(&line))
	}
	_ = d
}

// TestRasterCacheReusesShapeAcrossColours pins the white-mask bake: a shape
// whose colour changes reuses one cached raster and applies the colour at
// blit time, so a colour animation does not re-rasterize its path per frame.
func TestRasterCacheReusesShapeAcrossColours(t *testing.T) {
	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(40, 40)
	defer dst.Deallocate()

	op := &layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 20, H: 20,
		R: 1, G: 0, B: 0, Alpha: 1, Radius: 4,
	}
	renderer.fillRect(dst, op, 0, 0)
	if got := len(renderer.rasters.entries); got != 1 {
		t.Fatalf("rasters = %d, want one", got)
	}

	op.R, op.G, op.B = 0, 0, 1
	renderer.fillRect(dst, op, 0, 0)
	if got := len(renderer.rasters.entries); got != 1 {
		t.Fatalf("a colour change baked a second raster: %d entries", got)
	}
}

// TestRasterCacheTouchOnHit verifies that a cache hit moves the key to the back
// of the eviction order for true LRU retention.
func TestRasterCacheTouchOnHit(t *testing.T) {
	var c rasterCache
	k1 := rasterKey{variant: 1, w: 10, h: 10}
	k2 := rasterKey{variant: 1, w: 20, h: 20}

	c.raster(k1, image.Pt(10, 10), func(img *ebiten.Image) {})
	c.raster(k2, image.Pt(20, 20), func(img *ebiten.Image) {})

	if c.order.length() != 2 {
		t.Fatalf("order length = %d, want 2", c.order.length())
	}
	if oldest, _ := c.order.oldest(rasterKey{}); oldest != k1 {
		t.Fatalf("oldest = %v, want %v", oldest, k1)
	}

	// Hit k1: it becomes the most recent, so k2 is now the oldest.
	c.raster(k1, image.Pt(10, 10), func(img *ebiten.Image) {})
	if oldest, _ := c.order.oldest(rasterKey{}); oldest != k2 {
		t.Fatalf("oldest after hit = %v, want %v", oldest, k2)
	}
}

// rasterParityOps are the shapes the parity test draws both ways.
func rasterParityOps() []layout.DisplayOp {
	return []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 10.5, Y: 20.25, W: 42.5, H: 30, Radius: 3, R: 1, G: 0.5},
		{Kind: layout.DisplayOpStrokeRect, X: 10.5, Y: 20.25, W: 42.5, H: 30, Width: 2, Radius: 3, B: 1},
		{Kind: layout.DisplayOpFillRect, X: 70.5, Y: 20.25, W: 42.5, H: 30, Radius: 3, R: 1, G: 0.5},
		{Kind: layout.DisplayOpStrokeRect, X: 70, Y: 20.25, W: 42.5, H: 30, Width: 2, StrokeMask: 5, B: 1},
		{Kind: layout.DisplayOpFillRect, X: 130.75, Y: 60.5, W: 20, H: 20, Radius: 6, Alpha: 0.5},
	}
}

// drawDirect draws one op through the uncached path only.
func drawDirect(op *layout.DisplayOp, dst *ebiten.Image) {
	switch op.Kind {
	case layout.DisplayOpFillRect:
		radii, ok := masking.FillRadii(op)
		if !ok {
			return
		}
		fillRounded(dst, op, radii, 0, 0)
	case layout.DisplayOpStrokeRect:
		strokeDirect(dst, op, 0, 0)
	}
}

func TestQuantizeRasterFracSnapsToCornerGrid(t *testing.T) {
	t.Parallel()

	a := quantizeRasterFrac(0.5001)
	b := quantizeRasterFrac(0.4999)
	if a != b || a != 0.5 {
		t.Fatalf("quantizeRasterFrac drift: a=%v b=%v, want both 0.5", a, b)
	}
}
