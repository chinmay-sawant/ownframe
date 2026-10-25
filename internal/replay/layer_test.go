package replay

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// layerDisplay builds a page with a white background and a promoted card of
// two overlapping child fills.
func layerDisplay() *layout.Display {
	// Ops carry layout points; the canvas is CSS pixels. One CSS pixel is
	// 0.75 points, the same ratio a laid-out display carries. PixelPerPoint is
	// left at 1 so the test matrices translate in whole pixels.
	const pt = 72.0 / 96.0

	return &layout.Display{
		Width: 120, Height: 90, PointsPerPixel: pt, PixelPerPoint: 1,
		Boxes: []layout.Box{
			{ID: "page", X: 0, Y: 0, W: 120, H: 90},
			{ID: "card", X: 20, Y: 20, W: 30, H: 10, ParentIndex: 0},
		},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 120 * pt, H: 90 * pt, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 20 * pt, Y: 20 * pt, W: 30 * pt, H: 10 * pt, R: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 30 * pt, Y: 22 * pt, W: 20 * pt, H: 8 * pt, B: 1, Alpha: 1},
		},
		Order:    []int{0, 1, 2},
		OpOwners: []int{-1, 1, 1},
	}
}

func layerSpecsFor(d *layout.Display, matrix layout.Matrix2D, alpha float64) []LayerSpec {
	return []LayerSpec{{
		ElementID: d.Boxes[1].ID,
		BoxIndex:  1,
		Matrix:    matrix,
		Alpha:     alpha,
	}}
}

func TestLayerCacheHitInvalidateAndClose(t *testing.T) {
	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	specs := layerSpecsFor(d, layerTranslate(4, 0), 1)
	first := renderer.BeginLayers(d, 7, specs)
	if len(first) == 0 {
		t.Fatal("first promotion reported no damage")
	}
	stats := renderer.Layers()
	if stats.Active != 1 || stats.Promoted != 1 || stats.Rebaked != 1 {
		t.Fatalf("first stats = %+v", stats)
	}
	layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 7}]
	if layer == nil || layer.Tex == nil || !layer.Valid {
		t.Fatal("the layer was not retained")
	}
	tex, sig := layer.Tex, layer.Sig

	second := renderer.BeginLayers(d, 7, specs)
	if len(second) != 0 {
		t.Fatalf("unchanged layer damage = %v", second)
	}
	stats = renderer.Layers()
	if stats.Rebaked != 0 || stats.Active != 1 {
		t.Fatalf("hit stats = %+v", stats)
	}
	if layer.Tex != tex {
		t.Fatal("a cache hit allocated a new texture")
	}

	moved := layerSpecsFor(d, layerTranslate(12, 0), 1)
	damage := renderer.BeginLayers(d, 7, moved)
	if len(damage) == 0 {
		t.Fatal("a matrix change reported no damage")
	}
	if layer.Tex != tex || layer.Sig != sig {
		t.Fatal("a matrix change re-baked the layer")
	}

	d.Ops[1].R = 0.5
	renderer.BeginLayers(d, 7, moved)
	if layer.Tex != tex {
		t.Fatal("a same-box content change allocated a new texture")
	}
	if layer.Sig == sig {
		t.Fatal("a content change kept the signature")
	}
	if renderer.Layers().Rebaked != 1 {
		t.Fatalf("rebake stats = %+v", renderer.Layers())
	}

	renderer.BeginLayers(d, 8, moved)
	if renderer.layers.entries[layerKey{id: "card", box: 1, gen: 8}] == nil {
		t.Fatal("a new generation did not create a new slot")
	}
	if renderer.layers.entries[layerKey{id: "card", box: 1, gen: 7}] != nil {
		t.Fatal("a new generation kept the old slot")
	}

	renderer.Close()
	if len(renderer.layers.entries) != 0 || renderer.layers.pixels != 0 {
		t.Fatal("Close retained layers")
	}
}

func TestLayerLRUEvictionAndPixelCap(t *testing.T) {
	var store layerStore
	defer store.close()
	renderer := &Renderer{}
	d := &layout.Display{}

	for i := 0; i < layerMaxEntries+4; i++ {
		store.use++
		cand := layerCandidate{
			key:    layerKey{id: fmt.Sprintf("l%d", i), box: i, gen: 1},
			box:    image.Rect(0, 0, 2, 2),
			sig:    uint64(i + 1),
			alpha:  1,
			matrix: layout.IdentityMatrix(),
		}
		if _, _, ok, _ := store.getLocked(cand, nil, renderer, d); !ok {
			t.Fatalf("layer %d was not retained", i)
		}
	}
	if got := len(store.entries); got != layerMaxEntries {
		t.Fatalf("entries = %d, want %d", got, layerMaxEntries)
	}
	if _, ok := store.entries[layerKey{id: "l0", box: 0, gen: 1}]; ok {
		t.Fatal("the least recently used layer survived")
	}
	if store.pixels > layerMaxPixels {
		t.Fatalf("pixels = %d, want at most %d", store.pixels, layerMaxPixels)
	}

	store.close()
	if len(store.entries) != 0 || store.pixels != 0 {
		t.Fatal("close retained entries or pixels")
	}

	huge := layerCandidate{
		key: layerKey{id: "huge", box: 99, gen: 1},
		box: image.Rect(0, 0, 6000, 6000), sig: 1, alpha: 1,
		matrix: layout.IdentityMatrix(),
	}
	if _, _, ok, _ := store.getLocked(huge, nil, renderer, d); ok {
		t.Fatal("a layer larger than the pixel cap was retained")
	}
}

func TestLayerPromotionPredicate(t *testing.T) {
	d := layerDisplay()
	active := layerTranslate(2, 0)

	bad := active
	if _, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: bad}, 1, nil, nil); reason != "" {
		t.Fatalf("valid layer rejected: %s", reason)
	}

	noAnim := layout.IdentityMatrix()
	if _, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: noAnim, Alpha: 1}, 1, nil, nil); reason != "" {
		t.Fatalf("a rest-pose candidate was rejected before the cache check: %s", reason)
	}

	d.Boxes[1].Transform.HasTransform3D = true
	if _, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: active}, 1, nil, nil); reason != "3d transform" {
		t.Fatalf("3d reason = %q", reason)
	}
	d.Boxes[1].Transform.HasTransform3D = false

	d.Ops[1].Fixed = true
	if _, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: active}, 1, nil, nil); reason == "" {
		t.Fatal("a fixed op was promoted")
	}
	d.Ops[1].Fixed = false

	one := layerDisplay()
	one.Ops = one.Ops[:2]
	one.Order = []int{0, 1}
	one.OpOwners = []int{-1, 1}
	if _, reason := resolveLayer(one, LayerSpec{BoxIndex: 1, Matrix: active}, 1, nil, nil); reason != "fewer than two owned ops" {
		t.Fatalf("one-op reason = %q", reason)
	}

	candidate, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: active}, 1, nil, nil)
	if reason != "" {
		t.Fatal(reason)
	}
	if len(candidate.owned) != 2 || candidate.firstRank != 1 || candidate.lastRank != 2 {
		t.Fatalf("capture = %+v", candidate)
	}

	claimed := make([]bool, len(d.Ops))
	claimed[1] = true
	if _, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: active}, 1, claimed, nil); reason != "overlaps another layer" {
		t.Fatalf("nested reason = %q", reason)
	}
}

func TestLayerCountCapKeepsTheRecentSet(t *testing.T) {
	d := layerCapDisplay(20)
	var renderer Renderer
	defer renderer.Close()

	specs := make([]LayerSpec, 20)
	for i := range specs {
		specs[i] = LayerSpec{
			ElementID: d.Boxes[i+1].ID, BoxIndex: i + 1,
			Matrix: layerTranslate(1, 0), Alpha: 1,
		}
	}

	renderer.BeginLayers(d, 1, specs)
	stats := renderer.Layers()
	if stats.Active != layerMaxEntries {
		t.Fatalf("active = %d, want %d", stats.Active, layerMaxEntries)
	}
	if got := stats.Reasons[layerReasonCountCap]; got != 4 {
		t.Fatalf("count-cap demotions = %d, want 4 (%v)", got, stats.Reasons)
	}

	// The second frame keeps the same recent set instead of thrashing the
	// four new candidates back in.
	renderer.BeginLayers(d, 1, specs)
	if got := renderer.Layers().Reasons[layerReasonCountCap]; got != 4 {
		t.Fatalf("second-frame cap demotions = %d, want 4", got)
	}
}

func layerCapDisplay(boxes int) *layout.Display {
	const pt = 72.0 / 96.0

	d := &layout.Display{
		Width: 400, Height: 40, PointsPerPixel: pt, PixelPerPoint: 1,
		Boxes: []layout.Box{{ID: "page", X: 0, Y: 0, W: 400, H: 40}},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 400 * pt, H: 40 * pt, R: 1, G: 1, B: 1, Alpha: 1},
		},
		Order:    []int{0},
		OpOwners: []int{-1},
	}
	for i := 0; i < boxes; i++ {
		x := float64(i * 5)
		d.Boxes = append(d.Boxes, layout.Box{
			ID: fmt.Sprintf("card%d", i), X: x, Y: 5, W: 4, H: 4, ParentIndex: 0,
		})
		for row := 0; row < 2; row++ {
			index := len(d.Ops)
			d.Ops = append(d.Ops, layout.DisplayOp{
				Kind: layout.DisplayOpFillRect,
				X:    (x + 1) * pt, Y: (6 + float64(row)) * pt,
				W: 2 * pt, H: pt, R: 1, Alpha: 1,
			})
			d.Order = append(d.Order, index)
			d.OpOwners = append(d.OpOwners, i+1)
		}
	}

	return d
}

func TestLayerSkipAndFrameDamage(t *testing.T) {
	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	renderer.BeginLayers(d, 1, layerSpecsFor(d, layerTranslate(3, 0), 1))
	frame := &renderer.layers.frame
	if !frame.skip[1] || !frame.skip[2] || frame.skip[0] {
		t.Fatalf("skip = %v", frame.skip)
	}
	if frame.start[1] == nil || frame.start[1].ElementID != "card" {
		t.Fatalf("start = %v", frame.start)
	}

	damage := renderer.BeginLayers(d, 1, nil)
	if len(damage) != 1 {
		t.Fatalf("demotion damage = %v", damage)
	}
	if renderer.LayersActive(d) {
		t.Fatal("a demoted frame stayed active")
	}
	if got := renderer.Layers().Reasons[layerReasonEnded]; got != 1 {
		t.Fatalf("demotion reasons = %v", renderer.Layers().Reasons)
	}
}

func TestLayerBasePoseCancelsAnimation(t *testing.T) {
	d := layerDisplay()
	clean, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: layerTranslate(1, 0)}, 1, nil, nil)
	if reason != "" {
		t.Fatal(reason)
	}

	centerX, centerY := 35*0.75, 25*0.75
	for _, deg := range []float64{30, 75} {
		anim := layout.BakeOrigin(layerRotateDeg(deg), centerX, centerY)
		for _, index := range []int{1, 2} {
			d.Ops[index].SetXform(anim)
		}
		base, ok := InvertLayerMatrix(anim)
		if !ok {
			t.Fatal("the animation matrix did not invert")
		}
		got, reason := resolveLayer(d, LayerSpec{
			BoxIndex: 1, Matrix: anim, Base: base,
		}, 1, nil, nil)
		if reason != "" {
			t.Fatalf("rotated layer rejected: %s", reason)
		}
		if got.box != clean.box {
			t.Fatalf("base box at %.0f deg = %v, want %v", deg, got.box, clean.box)
		}
		if got.sig != clean.sig {
			t.Fatalf("matrix change entered the signature at %.0f deg", deg)
		}
		if got.matrix != anim {
			t.Fatalf("composite matrix at %.0f deg = %v", deg, got.matrix)
		}
	}
}

func TestLayerGeoMMapsBoxAndMatrix(t *testing.T) {
	layer := &Layer{Box: image.Rect(10, 20, 30, 40), Matrix: layerTranslate(5, 7)}
	d := &layout.Display{PixelPerPoint: 2}
	geo := layerGeoM(layer, d, 100, 200)

	if got := geo.Element(0, 2); got != 10+5*2+100 {
		t.Fatalf("x translation = %v", got)
	}
	if got := geo.Element(1, 2); got != 20+7*2+200 {
		t.Fatalf("y translation = %v", got)
	}

	rot := layout.BakeOrigin(layerRotateDeg(90), 35, 25)
	layer.Matrix = rot
	geo = layerGeoM(layer, &layout.Display{PixelPerPoint: 1}, 0, 0)
	if math.Abs(geo.Element(0, 0)) > 1e-9 || math.Abs(geo.Element(0, 1)+1) > 1e-9 {
		t.Fatalf("rotation linear part = %v", geo)
	}
}

// TestLayerPixels runs the readback checks inside Ebiten's Draw loop. It is
// gated like the other replay pixel checks because readback requires a live
// graphics session.
func TestLayerPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_REPLAY_GPU_TEST") != "1" && os.Getenv("OWNFRAME_LAYER_GPU_TEST") != "1" {
		t.Skip("requires an isolated graphics session")
	}

	game := &layerPixelGame{checks: []func() error{
		checkLayerAlpha,
		checkLayerAlphaOneMatchesDirect,
		checkLayerAlphaZero,
		checkLayerMatrixAndAlpha,
		checkLayerMatrixAndHole,
		checkLayerRotation,
		checkLayerTransitionPixels,
	}}
	ebiten.SetWindowSize(120, 90)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}
	if game.err != nil {
		t.Fatal(game.err)
	}
}

type layerPixelGame struct {
	checks []func() error
	ran    bool
	done   bool
	err    error
}

func (g *layerPixelGame) Layout(int, int) (int, int) { return 120, 90 }

func (g *layerPixelGame) Update() error {
	if g.done {
		return ebiten.Termination
	}

	return nil
}

func (g *layerPixelGame) Draw(*ebiten.Image) {
	if g.ran {
		return
	}
	g.ran = true
	for _, check := range g.checks {
		if err := check(); err != nil {
			g.err = err
			break
		}
	}
	g.done = true
}

// checkLayerAlpha verifies group alpha flattens two overlapping children once:
// the overlap must match the top child alone, not a double blend. The owned ops
// carry the sampled opacity the animation runtime writes, so a bake that keeps
// it would apply the group alpha twice.
func checkLayerAlpha() error {
	d := layerDisplay()
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(0.5)
	}

	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(120, 90)
	defer dst.Deallocate()

	renderer.BeginLayers(d, 1, layerSpecsFor(d, layout.IdentityMatrix(), 0.5))
	if stats := renderer.Layers(); stats.Active != 1 {
		return fmt.Errorf("layer not active: %+v", stats)
	}
	renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)

	pixels := make([]byte, 120*90*4)
	dst.ReadPixels(pixels)

	red := pixelAt(pixels, 120, 22, 25)     // red only
	overlap := pixelAt(pixels, 120, 35, 24) // blue over red
	blue := pixelAt(pixels, 120, 46, 24)    // blue only
	bg := pixelAt(pixels, 120, 5, 5)

	if !layerColorNear(red, color.RGBA{R: 255, G: 128, B: 128, A: 255}) {
		layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 1}]
		debug := "nil"
		if layer != nil {
			tp := make([]byte, layer.Tex.Bounds().Dx()*layer.Tex.Bounds().Dy()*4)
			layer.Tex.ReadPixels(tp)
			debug = fmt.Sprintf("box=%v pose=%v tex=%v", layer.Box, layer.pose, pixelAt(tp, layer.Tex.Bounds().Dx(), 2, 2))
		}

		return fmt.Errorf("red area = %v; %s", red, debug)
	}
	if !layerColorNear(overlap, blue) {
		return fmt.Errorf("overlap %v != single blue %v; the layer alpha applied twice", overlap, blue)
	}
	if !layerColorNear(blue, color.RGBA{R: 128, G: 128, B: 255, A: 255}) {
		return fmt.Errorf("blue area = %v", blue)
	}
	if !layerColorNear(bg, color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		return fmt.Errorf("background = %v", bg)
	}

	return nil
}

// checkLayerAlphaOneMatchesDirect pins that a group alpha of one leaves the
// content unchanged: the layer texture must match the direct replay pixel for
// pixel once composited at the base pose.
func checkLayerAlphaOneMatchesDirect() error {
	d := layerDisplay()

	var renderer Renderer
	defer renderer.Close()

	// resolveLayer refuses a spec with no animation, which is what keeps a
	// static page direct. Resolve a moving spec for its captured box and rank,
	// then composite it at the base pose with alpha one to isolate the bake.
	cand, reason := resolveLayer(d, LayerSpec{
		ElementID: "card", BoxIndex: 1, Matrix: layerTranslate(1, 0), Alpha: 1,
	}, 1, nil, nil)
	if reason != "" {
		return fmt.Errorf("alpha-one candidate: %s", reason)
	}
	cand.matrix = layout.IdentityMatrix()
	layer, _, ok, _ := renderer.layers.getLocked(cand, nil, &renderer, d)
	if !ok {
		return fmt.Errorf("alpha-one layer was not retained")
	}
	layer.pose = layerPose(layer, d)
	layer.firstRank = cand.firstRank
	renderer.layers.buildFrame(d, []*Layer{layer})

	layered := ebiten.NewImage(120, 90)
	defer layered.Deallocate()
	direct := ebiten.NewImage(120, 90)
	defer direct.Deallocate()

	renderer.DrawVisibleUnfixedZLayers(layered, d, 0, 0, 0)
	renderer.DrawVisibleUnfixedZ(direct, d, 0, 0, 0)

	a := make([]byte, 120*90*4)
	b := make([]byte, 120*90*4)
	layered.ReadPixels(a)
	direct.ReadPixels(b)

	for i := 0; i < len(a); i += 4 {
		if !bytesNear(a[i:i+4], b[i:i+4], 2) {
			x, y := (i/4)%120, (i/4)/120
			return fmt.Errorf("alpha 1 pixel (%d,%d) = %v, want direct %v", x, y, a[i:i+4], b[i:i+4])
		}
	}

	return nil
}

// checkLayerAlphaZero verifies a transparent layer composites nothing and the
// page beneath it shows through.
func checkLayerAlphaZero() error {
	d := layerDisplay()

	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(120, 90)
	defer dst.Deallocate()

	renderer.BeginLayers(d, 1, layerSpecsFor(d, layout.IdentityMatrix(), 0))
	renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)

	pixels := make([]byte, 120*90*4)
	dst.ReadPixels(pixels)

	for _, at := range [][2]int{{22, 25}, {35, 24}, {46, 24}} {
		if got := pixelAt(pixels, 120, at[0], at[1]); !layerColorNear(got, color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
			return fmt.Errorf("alpha 0 card pixel (%d,%d) = %v, want the background", at[0], at[1], got)
		}
	}

	return nil
}

// checkLayerMatrixAndAlpha verifies a transform and an opacity entry composite
// together: the card moves by its matrix and the whole layer fades once.
func checkLayerMatrixAndAlpha() error {
	d := layerDisplay()
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(0.5)
	}

	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(120, 90)
	defer dst.Deallocate()

	renderer.BeginLayers(d, 1, layerSpecsFor(d, layerTranslate(30, 0), 0.5))
	if stats := renderer.Layers(); stats.Active != 1 {
		return fmt.Errorf("layer not active: %+v", stats)
	}
	renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)

	pixels := make([]byte, 120*90*4)
	dst.ReadPixels(pixels)

	moved := pixelAt(pixels, 120, 52, 25)
	if !layerColorNear(moved, color.RGBA{R: 255, G: 127, B: 127, A: 255}) {
		return fmt.Errorf("moved card pixel = %v, want one 50%% red over white", moved)
	}
	old := pixelAt(pixels, 120, 22, 25)
	if !layerColorNear(old, color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		return fmt.Errorf("old card pixel = %v, want the background", old)
	}

	return nil
}

// checkLayerMatrixAndHole moves a retained layer and repaints only old ∪ new,
// then compares the damaged region with a fresh full composite. The old
// position must be rebuilt from the page background, not left as a ghost.
func checkLayerMatrixAndHole() error {
	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(120, 90)
	defer dst.Deallocate()

	old := layerTranslate(0, -1)
	next := layerTranslate(30, 0)
	renderer.BeginLayers(d, 1, layerSpecsFor(d, old, 1))
	renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)

	damage := renderer.BeginLayers(d, 1, layerSpecsFor(d, next, 1))
	if len(damage) == 0 {
		return fmt.Errorf("a moving layer reported no damage")
	}
	renderer.DrawRectUnfixedZLayers(dst, d, damage[0], 0)

	fresh := ebiten.NewImage(120, 90)
	defer fresh.Deallocate()
	renderer.BeginLayers(d, 1, layerSpecsFor(d, next, 1))
	renderer.DrawVisibleUnfixedZLayers(fresh, d, 0, 0, 0)

	oldPixels := make([]byte, 120*90*4)
	newPixels := make([]byte, 120*90*4)
	dst.ReadPixels(oldPixels)
	fresh.ReadPixels(newPixels)

	for y := damage[0].Min.Y; y < damage[0].Max.Y; y++ {
		for x := damage[0].Min.X; x < damage[0].Max.X; x++ {
			i := (y*120 + x) * 4
			if !bytesNear(oldPixels[i:i+4], newPixels[i:i+4], 2) {
				return fmt.Errorf("damage pixel (%d,%d) = %v, want %v", x, y,
					oldPixels[i:i+4], newPixels[i:i+4])
			}
		}
	}

	hole := pixelAt(oldPixels, 120, 22, 25)
	if !layerColorNear(hole, color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		return fmt.Errorf("old layer pixel = %v, want the background", hole)
	}

	moved := pixelAt(oldPixels, 120, 52, 25)
	if !layerColorNear(moved, color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		return fmt.Errorf("moved card pixel = %v", moved)
	}

	return nil
}

// checkLayerRotation verifies the composite applies a rotation matrix to the
// retained texture rather than baking it into the layer.
func checkLayerRotation() error {
	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(120, 90)
	defer dst.Deallocate()

	// The card's box is (20,20)-(50,30), centred at (35,25). A 90 degree
	// rotation puts it at (30,10)-(40,40).
	rot := layout.BakeOrigin(layerRotateDeg(90), 35, 25)
	renderer.BeginLayers(d, 1, layerSpecsFor(d, rot, 1))
	renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)

	pixels := make([]byte, 120*90*4)
	dst.ReadPixels(pixels)

	top := pixelAt(pixels, 120, 35, 15)
	if !layerColorNear(top, color.RGBA{R: 255, A: 255}) {
		return fmt.Errorf("rotated card top = %v", top)
	}
	old := pixelAt(pixels, 120, 22, 25)
	if !layerColorNear(old, color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		return fmt.Errorf("base-pose pixel = %v, want background", old)
	}

	return nil
}

// BenchmarkLayerRotatedCard composites a retained card through a changing
// rotation. The texture must survive the matrix change, so the benchmark
// reports the operations each frame and the bakes the cache paid.
func BenchmarkLayerRotatedCard(b *testing.B) {
	d := layerBenchDisplay()
	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(d.Width, d.Height)
	defer dst.Deallocate()

	replayCounts.on.Store(true)
	defer replayCounts.on.Store(false)
	replayCounts.ops.Store(0)
	replayCounts.blits.Store(0)

	cx, cy := 300.0, 210.0
	spec := layerSpecsFor(d, layout.BakeOrigin(layerRotateDeg(1), cx, cy), 1)[0]
	renderer.BeginLayers(d, 1, []LayerSpec{spec})
	bakeOps := replayCounts.ops.Load()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		spec.Matrix = layout.BakeOrigin(layerRotateDeg(float64(i%360+1)), cx, cy)
		renderer.BeginLayers(d, 1, []LayerSpec{spec})
		renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)
	}

	b.StopTimer()
	frameOps := replayCounts.ops.Load() - bakeOps
	frameBlits := replayCounts.blits.Load()
	b.ReportMetric(float64(frameOps)/float64(b.N), "ops/frame")
	b.ReportMetric(float64(frameBlits)/float64(b.N), "blits/frame")
	b.ReportMetric(float64(bakeOps), "bake-ops")
	if stats := renderer.Layers(); stats.Rebaked > 0 {
		b.Logf("re-bakes = %d", stats.Rebaked)
	}
}

func layerBenchDisplay() *layout.Display {
	const (
		pt    = 72.0 / 96.0
		cards = 40
		cardX = 200.0
		cardY = 150.0
		cardW = 200.0
		cardH = 120.0
	)

	d := &layout.Display{
		Width: 800, Height: 600, PointsPerPixel: pt, PixelPerPoint: 1,
		Boxes: []layout.Box{
			{ID: "page", X: 0, Y: 0, W: 800, H: 600},
			{ID: "card", X: cardX, Y: cardY, W: cardW, H: cardH, ParentIndex: 0},
		},
	}
	d.Ops = append(d.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 800 * pt, H: 600 * pt,
		R: 0.9, G: 0.9, B: 0.9, Alpha: 1,
	})
	d.Order = append(d.Order, 0)
	d.OpOwners = append(d.OpOwners, -1)

	for i := 0; i < cards; i++ {
		d.Ops = append(d.Ops, layout.DisplayOp{
			Kind: layout.DisplayOpFillRect,
			X:    (cardX + 5) * pt,
			Y:    (cardY + float64(i)*2.8) * pt,
			W:    (cardW - 10) * pt,
			H:    2 * pt,
			R:    0.1, G: 0.2, B: 0.6, Alpha: 1,
		})
		d.Order = append(d.Order, i+1)
		d.OpOwners = append(d.OpOwners, 1)
	}

	return d
}

func layerTranslate(tx, ty float64) layout.Matrix2D {
	return layout.Matrix2D{A: 1, D: 1, E: tx, F: ty}
}

func layerRotateDeg(deg float64) layout.Matrix2D {
	rad := deg * math.Pi / 180
	c, s := math.Cos(rad), math.Sin(rad)

	return layout.Matrix2D{A: c, B: s, C: -s, D: c}
}

func pixelAt(pixels []byte, width, x, y int) color.RGBA {
	i := (y*width + x) * 4

	return color.RGBA{R: pixels[i], G: pixels[i+1], B: pixels[i+2], A: pixels[i+3]}
}

func layerColorNear(got, want color.RGBA) bool {
	const tol = 3

	diff := func(a, b uint8) bool {
		d := int(a) - int(b)
		if d < 0 {
			d = -d
		}

		return d <= tol
	}

	return diff(got.R, want.R) && diff(got.G, want.G) && diff(got.B, want.B) && diff(got.A, want.A)
}

func bytesNear(a, b []byte, tol int) bool {
	for i := range a {
		d := int(a[i]) - int(b[i])
		if d < 0 {
			d = -d
		}
		if d > tol {
			return false
		}
	}

	return true
}

// TestLayerContentChurnDemotes keeps a per-frame content animation from
// re-baking a texture every frame. After two misses the layer is dropped and
// a cooldown keeps it direct until its content settles.
func TestLayerContentChurnDemotes(t *testing.T) {
	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	specs := layerSpecsFor(d, layerTranslate(4, 0), 1)
	renderer.BeginLayers(d, 7, specs)
	if !renderer.LayersActive(d) {
		t.Fatal("the first frame did not promote")
	}

	d.Ops[1].R = 0.5
	renderer.BeginLayers(d, 7, specs)
	d.Ops[1].R = 0.4
	renderer.BeginLayers(d, 7, specs)

	if renderer.LayersActive(d) {
		t.Fatal("a layer that re-bakes every frame stayed active")
	}
	if got := renderer.Layers().Reasons[layerReasonContentChurn]; got != 1 {
		t.Fatalf("churn reasons = %v, want one %q", renderer.Layers().Reasons, layerReasonContentChurn)
	}

	renderer.BeginLayers(d, 7, specs)
	if renderer.LayersActive(d) {
		t.Fatal("a churning layer promoted during its cooldown")
	}

	for range layerChurnCooldown + 2 {
		renderer.BeginLayers(d, 7, specs)
	}
	if !renderer.LayersActive(d) {
		t.Fatal("a settled layer did not promote after the cooldown")
	}
}

// TestLayerRestPoseKeepsRetainedLayer keeps an alternate animation on its
// layer at the turn. A new static element is still rejected, but an element
// already held as a layer keeps its texture through the identity pose instead
// of demoting and re-promoting on the next frame.
func TestLayerRestPoseKeepsRetainedLayer(t *testing.T) {
	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	renderer.BeginLayers(d, 7, layerSpecsFor(d, layerTranslate(4, 0), 1))
	layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 7}]
	if layer == nil {
		t.Fatal("the moving layer was not retained")
	}
	tex := layer.Tex

	renderer.BeginLayers(d, 7, layerSpecsFor(d, layout.IdentityMatrix(), 1))
	if !renderer.LayersActive(d) {
		t.Fatal("the rest pose demoted a held layer")
	}
	if layer.Tex != tex {
		t.Fatal("the rest pose re-baked the texture")
	}
	if !layer.Matrix.IsIdentity() {
		t.Fatalf("rest matrix = %+v, want identity", layer.Matrix)
	}

	var fresh Renderer
	defer fresh.Close()
	fresh.BeginLayers(d, 7, layerSpecsFor(d, layout.IdentityMatrix(), 1))
	if fresh.LayersActive(d) {
		t.Fatal("a new static element was promoted")
	}
	if got := fresh.Layers().Reasons[layerReasonRest]; got != 1 {
		t.Fatalf("rest reasons = %v, want one %q", fresh.Layers().Reasons, layerReasonRest)
	}
}
