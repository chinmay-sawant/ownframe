package replay

import (
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"math"
	"os"
	"testing"
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

// paintedOpacity builds a fill op whose element opacity the animation runtime
// has written. Alpha is the op's own source alpha.
func paintedOpacity(paint, alpha float64) layout.DisplayOp {
	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: alpha}
	op.SetPaintOpacity(paint)

	return op
}

// TestLayerBakeOpacityRule pins the compensation rule: an operation whose
// element opacity is the sampled group alpha bakes without it, so the composite
// applies it once; an operation with an opacity of its own keeps it.
func TestLayerBakeOpacityRule(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		paint      float64
		alpha      float64
		groupAlpha float64
		want       float64
	}{
		{name: "alpha one is unchanged", paint: 1, alpha: 1, groupAlpha: 1, want: 1},
		{name: "own opacity at alpha one", paint: .4, alpha: 1, groupAlpha: 1, want: .4},
		{name: "group alpha removed", paint: .5, alpha: 1, groupAlpha: .5, want: 1},
		{name: "group alpha with own source alpha", paint: .5, alpha: .5, groupAlpha: .5, want: 1},
		{name: "own opacity kept", paint: .25, alpha: 1, groupAlpha: .5, want: .25},
		{name: "alpha zero leaves paint", paint: 1, alpha: 1, groupAlpha: 0, want: 1},
		{name: "zero paint reads unset", paint: 0, alpha: 1, groupAlpha: .5, want: 1},
		{name: "nan group leaves paint", paint: .4, alpha: 1, groupAlpha: math.NaN(), want: .4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			op := paintedOpacity(tc.paint, tc.alpha)
			if got := layerBakeOpacity(&op, tc.groupAlpha); math.Abs(got-tc.want) > 1e-9 {
				t.Fatalf("bake opacity = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestLayerBakeKeepsPerOpOpacity checks two children under one group alpha: the
// child the animation wrote bakes at 1, and the child with its own opacity keeps
// it. The composite then multiplies both by the group alpha once.
func TestLayerBakeKeepsPerOpOpacity(t *testing.T) {
	t.Parallel()

	const groupAlpha = .5
	animated := paintedOpacity(groupAlpha, 1)
	own := paintedOpacity(.25, 1)

	if got := layerBakeOpacity(&animated, groupAlpha); got != 1 {
		t.Fatalf("animated child bake opacity = %v, want 1", got)
	}
	if got := layerBakeOpacity(&own, groupAlpha); got != .25 {
		t.Fatalf("own-opacity child bake opacity = %v, want .25", got)
	}
	if got := layerBakeOpacity(&own, groupAlpha) * groupAlpha; got != .125 {
		t.Fatalf("own-opacity child composite = %v, want .125", got)
	}
}

// TestLayerBakeDoesNotMutateDisplay proves the bake borrows the operation's
// paint opacity and matrix and restores both. DisplayOp shares its rare payload,
// so a struct copy that writes either field would change the live display.
func TestLayerBakeDoesNotMutateDisplay(t *testing.T) {
	t.Parallel()

	d := layerDisplay()
	anim := layout.BakeOrigin(layerRotateDeg(30), 35, 25)
	base, ok := InvertLayerMatrix(anim)
	if !ok {
		t.Fatal("the animation matrix did not invert")
	}
	for _, index := range []int{1, 2} {
		d.Ops[index].SetPaintOpacity(.5)
		d.Ops[index].SetXform(anim)
	}

	beforeOpacity := make([]float64, len(d.Ops))
	beforeXform := make([]layout.Matrix2D, len(d.Ops))
	beforeSet := make([]bool, len(d.Ops))
	for i := range d.Ops {
		beforeOpacity[i] = d.Ops[i].Opacity()
		beforeXform[i] = d.Ops[i].Transform()
		beforeSet[i] = d.Ops[i].XformSet
	}

	var renderer Renderer
	defer renderer.Close()
	renderer.BeginLayers(d, 1, []LayerSpec{{
		ElementID: "card", BoxIndex: 1, Matrix: anim, Base: base, Alpha: .5,
	}})

	if stats := renderer.Layers(); stats.Active != 1 {
		t.Fatalf("layer not active: %+v", stats)
	}
	for i := range d.Ops {
		if got := d.Ops[i].Opacity(); got != beforeOpacity[i] {
			t.Fatalf("op %d opacity = %v, want %v", i, got, beforeOpacity[i])
		}
		if got := d.Ops[i].Transform(); got != beforeXform[i] {
			t.Fatalf("op %d matrix = %+v, want %+v", i, got, beforeXform[i])
		}
		if got := d.Ops[i].XformSet; got != beforeSet[i] {
			t.Fatalf("op %d XformSet = %v, want %v", i, got, beforeSet[i])
		}
	}
}

// TestLayerOpacityChangeReusesTexture pins the promotion win: changing only the
// sampled group alpha reports damage but does not re-bake the texture, because
// the group alpha is not baked into it.
func TestLayerOpacityChangeReusesTexture(t *testing.T) {
	t.Parallel()

	d := layerDisplay()
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(.5)
	}

	spec := LayerSpec{
		ElementID: "card", BoxIndex: 1,
		Matrix: layerTranslate(1, 0), Alpha: .5,
	}
	var renderer Renderer
	defer renderer.Close()
	renderer.BeginLayers(d, 1, []LayerSpec{spec})

	layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 1}]
	if layer == nil {
		t.Fatal("the layer was not retained")
	}
	tex, sig := layer.Tex, layer.Sig

	// The runtime clears the paint opacity when it samples zero, so the frame
	// carries no group alpha in the operations. The texture must still be the
	// same one: it already holds the group-free pixels and the composite is
	// skipped while alpha is zero.
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(1)
	}
	spec.Alpha = 0
	if damage := renderer.BeginLayers(d, 1, []LayerSpec{spec}); len(damage) == 0 {
		t.Fatal("an alpha change to zero reported no damage")
	}
	if stats := renderer.Layers(); stats.Rebaked != 0 {
		t.Fatalf("alpha zero rebaked: %+v", stats)
	}
	if layer.Tex != tex || layer.Sig != sig || layer.Alpha != 0 {
		t.Fatal("alpha zero did not reuse the group-free texture")
	}

	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(.25)
	}
	spec.Matrix = layerTranslate(2, 0)
	spec.Alpha = .25
	damage := renderer.BeginLayers(d, 1, []LayerSpec{spec})

	if len(damage) == 0 {
		t.Fatal("an alpha and matrix change reported no damage")
	}
	if stats := renderer.Layers(); stats.Rebaked != 0 {
		t.Fatalf("alpha change rebaked: %+v", stats)
	}
	if layer.Tex != tex || layer.Sig != sig {
		t.Fatal("an opacity-only change re-baked the texture")
	}
	if layer.Alpha != .25 {
		t.Fatalf("layer alpha = %v, want .25", layer.Alpha)
	}
}

// TestLayerAlphaZeroCompositesNothing checks the composite skip. The texture is
// deliberately nil: a transparent layer must return before the blit and before
// touching it.
func TestLayerAlphaZeroCompositesNothing(t *testing.T) {
	dst := ebiten.NewImage(8, 8)
	defer dst.Deallocate()

	d := &layout.Display{PixelPerPoint: 1}
	layer := &Layer{Box: image.Rect(0, 0, 8, 8), Matrix: layout.IdentityMatrix()}
	renderer := &Renderer{}

	replayCounts.on.Store(true)
	defer replayCounts.on.Store(false)
	replayCounts.blits.Store(0)

	renderer.drawLayerPose(dst, d, layer, 0, 0)
	if got := replayCounts.blits.Load(); got != 0 {
		t.Fatalf("transparent layer blits = %d, want 0", got)
	}

	layer.Alpha = 1
	renderer.drawLayerPose(dst, d, layer, 0, 0)
	if got := replayCounts.blits.Load(); got != 1 {
		t.Fatalf("opaque layer blits = %d, want 1", got)
	}
}

// TestLayerStaticDisplayStaysDirect keeps a static page off the layer path: a
// spec with no animation is demoted, no layer becomes active, and no damage is
// reported, so the caller replays the display unchanged.
func TestLayerStaticDisplayStaysDirect(t *testing.T) {
	t.Parallel()

	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	damage := renderer.BeginLayers(d, 1, []LayerSpec{{
		ElementID: "card", BoxIndex: 1, Alpha: 1,
	}})
	if len(damage) != 0 {
		t.Fatalf("static spec damage = %v, want none", damage)
	}
	if renderer.LayersActive(d) {
		t.Fatal("a static page promoted a layer")
	}
	stats := renderer.Layers()
	if stats.Active != 0 || stats.Demoted != 1 {
		t.Fatalf("static stats = %+v", stats)
	}
	if got := stats.Reasons["no transform or opacity motion"]; got != 1 {
		t.Fatalf("demotion reasons = %v", stats.Reasons)
	}
}

// stampTransitionMatrix writes one sampled transition matrix onto the card's
// owned operations, the way the transition renderer writes every subtree
// operation.
func stampTransitionMatrix(d *layout.Display, m layout.Matrix2D) {
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetXform(m)
	}
}

// transitionSpec builds the specification the page derives from one transition
// sample: the sampled delta, its inverse as Base, and the group alpha.
func transitionSpec(m layout.Matrix2D, alpha float64) LayerSpec {
	base, _ := InvertLayerMatrix(m)

	return LayerSpec{
		ElementID: "card", BoxIndex: 1,
		Matrix: m, Base: base, Alpha: alpha,
	}
}

// TestLayerTransitionMatrixReusesTexture pins the transform transition win: the
// sampled matrix reaches the composite, and the subtree texture is captured
// once for the whole run instead of re-baking as the matrix changes.
func TestLayerTransitionMatrixReusesTexture(t *testing.T) {
	t.Parallel()

	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	first := layerTranslate(4, 0)
	stampTransitionMatrix(d, first)
	renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(first, 1)})

	layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 1}]
	if layer == nil || layer.Tex == nil {
		t.Fatal("the transform transition did not promote")
	}
	tex, sig := layer.Tex, layer.Sig
	oldPose := layer.pose

	second := layerTranslate(10, 0)
	stampTransitionMatrix(d, second)
	damage := renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(second, 1)})

	if stats := renderer.Layers(); stats.Active != 1 || stats.Rebaked != 0 {
		t.Fatalf("second sample stats = %+v, want one layer and no rebake", stats)
	}
	if layer.Tex != tex || layer.Sig != sig {
		t.Fatal("a transform transition re-baked its texture")
	}
	if layer.Matrix != second {
		t.Fatalf("composite matrix = %+v, want the sampled %+v", layer.Matrix, second)
	}
	if len(damage) == 0 {
		t.Fatal("a moving transition layer reported no damage")
	}
	if !oldPose.In(damage[0]) {
		t.Fatalf("damage %v does not cover the old pose %v", damage[0], oldPose)
	}
	if !layer.pose.In(damage[0]) {
		t.Fatalf("damage %v does not cover the new pose %v", damage[0], layer.pose)
	}
}

// TestLayerTransitionOpacityAppliesGroupAlphaOnce pins the opacity transition
// rule: the specification carries the sampled alpha, every operation whose own
// element opacity is that group value bakes without it, and a later sample
// reuses the texture because the alpha is not baked in.
func TestLayerTransitionOpacityAppliesGroupAlphaOnce(t *testing.T) {
	t.Parallel()

	d := layerDisplay()
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(.5)
	}

	var renderer Renderer
	defer renderer.Close()
	renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(layout.IdentityMatrix(), .5)})

	layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 1}]
	if layer == nil {
		t.Fatal("the opacity transition did not promote")
	}
	if layer.Alpha != .5 {
		t.Fatalf("layer alpha = %v, want the sampled .5", layer.Alpha)
	}
	for _, index := range layer.Ops {
		if got := layerBakeOpacity(&d.Ops[index], layer.Alpha); got != 1 {
			t.Fatalf("op %d baked at %v; the group alpha must apply at composite, once", index, got)
		}
	}

	tex, sig := layer.Tex, layer.Sig
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(.75)
	}
	renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(layout.IdentityMatrix(), .75)})
	if stats := renderer.Layers(); stats.Rebaked != 0 {
		t.Fatalf("an alpha sample rebaked: %+v", stats)
	}
	if layer.Tex != tex || layer.Sig != sig {
		t.Fatal("an opacity transition re-baked its texture")
	}
	if layer.Alpha != .75 {
		t.Fatalf("layer alpha = %v, want the new sample .75", layer.Alpha)
	}
}

// TestLayerTransitionReleasedAtEnd pins the release rule: a transition that
// stops being listed releases its layer and damages the pose it leaves, so the
// direct replay paints the final pose with no ghost and no waiting for LRU
// eviction.
func TestLayerTransitionReleasedAtEnd(t *testing.T) {
	t.Parallel()

	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	m := layerTranslate(6, 0)
	stampTransitionMatrix(d, m)
	renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(m, 1)})
	layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 1}]
	if layer == nil {
		t.Fatal("the transition did not promote")
	}
	lastPose := layer.pose

	damage := renderer.BeginLayers(d, 1, nil)
	if renderer.LayersActive(d) {
		t.Fatal("a released transition left the layer active")
	}
	if len(damage) == 0 || !lastPose.In(damage[0]) {
		t.Fatalf("release damage = %v, want the last pose %v", damage, lastPose)
	}
	stats := renderer.Layers()
	if stats.Active != 0 || stats.Pixels != 0 {
		t.Fatalf("release stats = %+v, want the texture released", stats)
	}
	if got := stats.Reasons[layerReasonEnded]; got != 1 {
		t.Fatalf("release reasons = %v, want one %q", stats.Reasons, layerReasonEnded)
	}
	if len(renderer.layers.entries) != 0 {
		t.Fatal("the layer cache kept the released entry")
	}
}

// TestLayerTransitionProjectionAndBandDemotions keeps the existing demotion
// rules: a perspective element, a non-zero z-index operation, and a fixed
// operation each stay direct with a reason. A level-zero stacking-context
// marker, which every transition target carries, does not demote.
func TestLayerTransitionProjectionAndBandDemotions(t *testing.T) {
	t.Parallel()

	t.Run("perspective", func(t *testing.T) {
		t.Parallel()

		d := layerDisplay()
		d.Boxes[1].Transform.HasPerspective = true
		var renderer Renderer
		defer renderer.Close()

		renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(layerTranslate(2, 0), 1)})
		if renderer.LayersActive(d) {
			t.Fatal("a perspective element promoted")
		}
		if got := renderer.Layers().Reasons["3d transform"]; got != 1 {
			t.Fatalf("reasons = %v, want the 3d reason", renderer.Layers().Reasons)
		}
	})

	t.Run("stacking context at level zero", func(t *testing.T) {
		t.Parallel()

		d := layerDisplay()
		for i := 1; i < len(d.Ops); i++ {
			d.Ops[i].ZIndexSet = true
		}

		m := layerTranslate(2, 0)
		stampTransitionMatrix(d, m)
		var renderer Renderer
		defer renderer.Close()

		renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(m, 1)})
		if !renderer.LayersActive(d) {
			t.Fatalf("a transition target stacking context stayed direct: %v",
				renderer.Layers().Reasons)
		}
	})

	t.Run("non-zero z-index", func(t *testing.T) {
		t.Parallel()

		d := layerDisplay()
		for i := 1; i < len(d.Ops); i++ {
			d.Ops[i].ZIndex, d.Ops[i].ZIndexSet = 2, true
		}
		var renderer Renderer
		defer renderer.Close()

		renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(layerTranslate(2, 0), 1)})
		if renderer.LayersActive(d) {
			t.Fatal("a non-zero z-index operation promoted")
		}
		if got := renderer.Layers().Reasons["fixed or z-indexed op"]; got != 1 {
			t.Fatalf("reasons = %v, want the z-index reason", renderer.Layers().Reasons)
		}
	})

	t.Run("fixed", func(t *testing.T) {
		t.Parallel()

		d := layerDisplay()
		d.Ops[1].Fixed = true
		var renderer Renderer
		defer renderer.Close()

		renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(layerTranslate(2, 0), 1)})
		if renderer.LayersActive(d) {
			t.Fatal("a fixed operation promoted")
		}
		if got := renderer.Layers().Reasons["fixed or z-indexed op"]; got != 1 {
			t.Fatalf("reasons = %v, want the fixed reason", renderer.Layers().Reasons)
		}
	})
}

// TestLayerTransitionDamageUnion pins the damage shape of a running transition:
// the repaint covers the old and the new pose, so the tile rebuild clears the
// hole the layer left without a separate erase.
func TestLayerTransitionDamageUnion(t *testing.T) {
	t.Parallel()

	d := layerDisplay()
	var renderer Renderer
	defer renderer.Close()

	old := layerTranslate(0, 3)
	stampTransitionMatrix(d, old)
	renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(old, 1)})
	layer := renderer.layers.entries[layerKey{id: "card", box: 1, gen: 1}]
	if layer == nil {
		t.Fatal("the transition did not promote")
	}
	oldPose := layer.pose

	next := layerTranslate(0, 20)
	stampTransitionMatrix(d, next)
	damage := renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(next, 1)})
	if len(damage) != 1 {
		t.Fatalf("damage = %v, want one union rect", damage)
	}
	want := oldPose.Union(layer.pose)
	if damage[0] != want {
		t.Fatalf("damage = %v, want the old ∪ new %v", damage[0], want)
	}
	if damage[0] == oldPose || damage[0] == layer.pose {
		t.Fatalf("damage = %v, want both poses, not one", damage[0])
	}
}

// checkLayerTransitionPixels drives the transition shape through the composite:
// a moving layer with a group alpha must move by its matrix and fade the
// overlapping children once. It runs inside the gated pixel test.
func checkLayerTransitionPixels() error {
	d := layerDisplay()
	for i := 1; i < len(d.Ops); i++ {
		d.Ops[i].SetPaintOpacity(.5)
	}
	m := layerTranslate(20, 0)
	stampTransitionMatrix(d, m)

	var renderer Renderer
	defer renderer.Close()

	dst := ebiten.NewImage(120, 90)
	defer dst.Deallocate()

	renderer.BeginLayers(d, 1, []LayerSpec{transitionSpec(m, .5)})
	if stats := renderer.Layers(); stats.Active != 1 {
		return fmt.Errorf("transition layer not active: %+v", stats)
	}
	renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)

	pixels := make([]byte, 120*90*4)
	dst.ReadPixels(pixels)

	// The card's base box is (20,20)-(50,30). A 20px move puts it at
	// (40,20)-(70,30); the overlap of the two children is at (55,24).
	moved := pixelAt(pixels, 120, 42, 25)
	if !layerColorNear(moved, color.RGBA{R: 255, G: 128, B: 128, A: 255}) {
		return fmt.Errorf("moved card pixel = %v, want one 50%% red over white", moved)
	}
	overlap := pixelAt(pixels, 120, 55, 24)
	blue := pixelAt(pixels, 120, 66, 24)
	if !layerColorNear(overlap, blue) {
		return fmt.Errorf("overlap %v != single blue %v; the group alpha applied twice", overlap, blue)
	}
	old := pixelAt(pixels, 120, 22, 25)
	if !layerColorNear(old, color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		return fmt.Errorf("old card pixel = %v, want the background", old)
	}

	return nil
}

// TestBeginViewportLayersRejectsCoveredLayer pins the viewport z-order guard:
// when a later unfixed static op overlaps the layer's base or animated bounds,
// BeginViewportLayers rejects promotion so DrawLayers never composites above
// the later sibling/popover, while non-overlapping or fixed/pinned later ops
// still allow promotion.
func TestBeginViewportLayersRejectsCoveredLayer(t *testing.T) {
	t.Parallel()
	const pt = 72.0 / 96.0

	// 1. Overlapping later static sibling (popover) covering card's base box.
	d := layerDisplay()
	d.Ops = append(d.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 25 * pt, Y: 20 * pt, W: 30 * pt, H: 20 * pt, G: 1, Alpha: 1,
	})
	d.Order = append(d.Order, 3)
	d.OpOwners = append(d.OpOwners, 0)

	var r Renderer
	defer r.Close()
	specs := layerSpecsFor(d, layerTranslate(2, 0), 1)

	// Tile path composites at firstRank, so it still promotes.
	r.BeginLayers(d, 1, specs)
	if !r.LayersActive(d) {
		t.Fatal("tile path rejected layer unnecessarily")
	}
	r.DemoteLayers()

	// Viewport path composites above viewport.buf, so it must reject.
	r.BeginViewportLayers(d, 1, specs, 0)
	if r.LayersActive(d) {
		t.Fatal("viewport path promoted a layer covered by a later static op")
	}
	if got := r.Layers().Reasons[layerReasonCovered]; got != 1 {
		t.Fatalf("covered reason = %d, want 1 (%v)", got, r.Layers().Reasons)
	}

	// 2. Later static op overlapping ONLY the animated pose must also reject.
	dAnim := layerDisplay()
	dAnim.Ops = append(dAnim.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 70 * pt, Y: 20 * pt, W: 20 * pt, H: 15 * pt, G: 1, Alpha: 1,
	})
	dAnim.Order = append(dAnim.Order, 3)
	dAnim.OpOwners = append(dAnim.OpOwners, 0)
	r.BeginViewportLayers(dAnim, 2, layerSpecsFor(dAnim, layerTranslate(45, 0), 1), 0)
	if r.LayersActive(dAnim) {
		t.Fatal("viewport path promoted a layer whose animated pose overlaps a later static op")
	}

	// 3. Non-overlapping later static op still promotes.
	dClear := layerDisplay()
	dClear.Ops = append(dClear.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 80 * pt, Y: 60 * pt, W: 20 * pt, H: 15 * pt, G: 1, Alpha: 1,
	})
	dClear.Order = append(dClear.Order, 3)
	dClear.OpOwners = append(dClear.OpOwners, 0)
	r.BeginViewportLayers(dClear, 3, layerSpecsFor(dClear, layerTranslate(2, 0), 1), 0)
	if !r.LayersActive(dClear) {
		t.Fatal("viewport path rejected a layer with a non-overlapping later op")
	}

	// 4. Overlapping later Fixed op draws after DrawLayers, so it still promotes.
	dFixed := layerDisplay()
	dFixed.Ops = append(dFixed.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 25 * pt, Y: 20 * pt, W: 30 * pt, H: 20 * pt, G: 1, Alpha: 1, Fixed: true,
	})
	dFixed.Order = append(dFixed.Order, 3)
	dFixed.OpOwners = append(dFixed.OpOwners, 0)
	r.BeginViewportLayers(dFixed, 4, layerSpecsFor(dFixed, layerTranslate(2, 0), 1), 0)
	if !r.LayersActive(dFixed) {
		t.Fatal("viewport path rejected a layer covered only by a fixed op")
	}
}

// TestLayerStoreEvictsWhenExistingLayerGrows pins that resizing an existing
// layer still enforces layerMaxPixels by evicting LRU entries.
func TestLayerStoreEvictsWhenExistingLayerGrows(t *testing.T) {
	t.Parallel()

	var store layerStore
	defer store.close()
	renderer := &Renderer{}
	d := &layout.Display{}

	// Seed an older unprotected entry taking 20M pixels (4000x5000).
	kOld := layerKey{id: "old", box: 0, gen: 1}
	store.use = 1
	candOld := layerCandidate{
		key: kOld, box: image.Rect(0, 0, 4000, 5000), sig: 1, alpha: 1, matrix: layout.IdentityMatrix(),
	}
	if _, _, ok, reason := store.getLocked(candOld, nil, renderer, d); !ok {
		t.Fatalf("seed old failed: %s", reason)
	}

	// Seed the growing entry taking 1M pixels (1000x1000). Total = 21M <= 32M.
	kGrow := layerKey{id: "grow", box: 1, gen: 1}
	store.use = 2
	candSmall := layerCandidate{
		key: kGrow, box: image.Rect(0, 0, 1000, 1000), sig: 2, alpha: 1, matrix: layout.IdentityMatrix(),
	}
	if _, _, ok, reason := store.getLocked(candSmall, nil, renderer, d); !ok {
		t.Fatalf("seed grow failed: %s", reason)
	}

	// Grow kGrow to 20M pixels (4000x5000). 20M + 20M > 32M, so kOld must evict.
	store.use = 3
	candBig := layerCandidate{
		key: kGrow, box: image.Rect(0, 0, 4000, 5000), sig: 3, alpha: 1, matrix: layout.IdentityMatrix(),
	}
	if _, _, ok, reason := store.getLocked(candBig, map[layerKey]bool{kGrow: true}, renderer, d); !ok {
		t.Fatalf("growing layer failed: %s", reason)
	}
	if store.pixels > layerMaxPixels {
		t.Fatalf("pixels = %d > layerMaxPixels %d after resize", store.pixels, layerMaxPixels)
	}
	if _, exists := store.entries[kOld]; exists {
		t.Fatal("old LRU entry was not evicted when existing layer grew")
	}
}

func TestLayerStoreChurnPrunesAndCloses(t *testing.T) {
	t.Parallel()

	var store layerStore
	k1 := layerKey{id: "expired", box: 1, gen: 1}
	k2 := layerKey{id: "active", box: 2, gen: 1}
	store.use = 100
	store.churn = map[layerKey]uint64{k1: 100, k2: 150}

	store.pruneChurnLocked()
	if _, ok := store.churn[k1]; ok {
		t.Fatal("expired churn entry was not pruned")
	}
	if _, ok := store.churn[k2]; !ok {
		t.Fatal("active churn entry was pruned early")
	}

	store.close()
	if store.churn != nil {
		t.Fatal("close did not clear churn map")
	}
}

func TestLayerUniformOverflowClip(t *testing.T) {
	const pt = 72.0 / 96.0
	d := layerDisplay()
	d.Boxes[1].ID = ""
	d.Ops[1].SetOverflowClip(22*pt, 18*pt, 16*pt, 14*pt)
	d.Ops[2].SetOverflowClip(22*pt, 18*pt, 16*pt, 14*pt)

	var renderer Renderer
	defer renderer.Close()

	specs := []LayerSpec{{
		BoxIndex: 1,
		Matrix:   layerTranslate(10, 0),
		Alpha:    1,
	}}
	damage := renderer.BeginLayers(d, 1, specs)
	if len(damage) == 0 {
		t.Fatal("uniform-clipped layer reported no damage")
	}
	stats := renderer.Layers()
	if stats.Active != 1 || stats.Promoted != 1 {
		t.Fatalf("stats = %+v, want active=1 promoted=1", stats)
	}
	layer := renderer.layers.entries[layerKey{id: "", box: 1, gen: 1}]
	if layer == nil || !layer.hasClip {
		t.Fatalf("layer = %+v, want hasClip=true", layer)
	}
	wantClip := image.Rect(22, 18, 38, 32)
	if layer.clip != wantClip {
		t.Fatalf("clip = %v, want %v", layer.clip, wantClip)
	}
	if !layer.pose.In(wantClip) {
		t.Fatalf("pose %v escapes clip %v", layer.pose, wantClip)
	}

	boxes := renderer.ActiveLayerBoxIndices(d)
	if len(boxes) != 1 || boxes[0] != 1 {
		t.Fatalf("ActiveLayerBoxIndices = %v, want [1]", boxes)
	}
	if ids := renderer.ActiveLayerIDs(); len(ids) != 0 {
		t.Fatalf("ActiveLayerIDs = %v, want empty for anonymous layer", ids)
	}

	// Drawing onto a dst outside the clip returns early without panic.
	dst := ebiten.NewImage(16, 16)
	defer dst.Deallocate()
	renderer.drawLayerPose(dst, d, layer, 0, 0)

	// Drawing onto a dst overlapping the clip succeeds.
	full := ebiten.NewImage(120, 90)
	defer full.Deallocate()
	renderer.drawLayerPose(full, d, layer, 0, 0)
}

func TestLayerMixedOverflowClipRejected(t *testing.T) {
	const pt = 72.0 / 96.0
	active := layerTranslate(4, 0)

	// One op clipped, one op unclipped.
	d := layerDisplay()
	d.Ops[1].SetOverflowClip(10*pt, 10*pt, 40*pt, 40*pt)
	if _, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: active}, 1, nil, nil); reason != "mixed clip rects" {
		t.Fatalf("clipped+unclipped reason = %q, want %q", reason, "mixed clip rects")
	}

	// Two ops with differing clip rects.
	d.Ops[2].SetOverflowClip(12*pt, 10*pt, 40*pt, 40*pt)
	if _, reason := resolveLayer(d, LayerSpec{BoxIndex: 1, Matrix: active}, 1, nil, nil); reason != "mixed clip rects" {
		t.Fatalf("differing clips reason = %q, want %q", reason, "mixed clip rects")
	}
}

// TestLayerBlendSiblingPromotion verifies that an unrelated mix-blend-mode
// element on the page does not demote a separate animated retained layer, while
// an element inside a blend group, a layer interleaved with a blend group, and
// a viewport layer covered by a later blend group are still rejected.
func TestLayerBlendSiblingPromotion(t *testing.T) {
	const pt = 72.0 / 96.0
	blendGroup := &layout.DisplayGroup{ID: 1, Mode: "multiply", Isolate: true}

	// Build a page with an animated card and an unrelated blend-mode sibling.
	makeDisplay := func(badgeX, badgeY float64) *layout.Display {
		d := layerDisplay()
		d.Boxes = append(d.Boxes, layout.Box{
			ID: "badge", X: badgeX, Y: badgeY, W: 20, H: 20, ParentIndex: 0,
		})
		badgeOp := layout.DisplayOp{
			Kind: layout.DisplayOpFillRect,
			X:    badgeX * pt, Y: badgeY * pt, W: 20 * pt, H: 20 * pt,
			G: 1, Alpha: 1,
		}
		badgeOp.SetBlendGroup(blendGroup)
		d.Ops = append(d.Ops, badgeOp)
		d.Order = append(d.Order, 3)
		d.OpOwners = append(d.OpOwners, 2)
		return d
	}

	// 1. Unrelated blend sibling promotes in BeginLayers and BeginViewportLayers,
	//    and drawing uses the group plan so the blend group stays buffered.
	d := makeDisplay(80, 50)
	var renderer Renderer
	defer renderer.Close()

	specs := layerSpecsFor(d, layerTranslate(6, 0), 1)
	renderer.BeginLayers(d, 1, specs)
	if !renderer.LayersActive(d) {
		t.Fatalf("unrelated blend sibling demoted BeginLayers: %+v", renderer.Layers())
	}

	dst := ebiten.NewImage(120, 90)
	defer dst.Deallocate()
	plansBefore := groupPlanBuildCount()
	renderer.DrawVisibleUnfixedZBackground(dst, d, 0, 0, 0)
	if groupPlanBuildCount() == plansBefore {
		t.Fatal("DrawVisibleUnfixedZBackground bypassed the group plan on a display with blend groups")
	}
	renderer.DrawLayers(dst, d, 0, 0)
	renderer.DrawVisibleUnfixedZLayers(dst, d, 0, 0, 0)
	renderer.DrawRectUnfixedZLayers(dst, d, image.Rect(0, 0, 120, 90), 0)

	renderer.DemoteLayers()
	renderer.BeginViewportLayers(d, 1, specs, 0)
	if !renderer.LayersActive(d) {
		t.Fatalf("unrelated blend sibling demoted BeginViewportLayers: %+v", renderer.Layers())
	}

	// 2. An element whose own op is inside a blend group is still rejected.
	dOwn := makeDisplay(80, 50)
	badgeOp2 := layout.DisplayOp{
		Kind: layout.DisplayOpFillRect,
		X:    82 * pt, Y: 52 * pt, W: 10 * pt, H: 10 * pt,
		B: 1, Alpha: 1,
	}
	badgeOp2.SetBlendGroup(blendGroup)
	dOwn.Ops = append(dOwn.Ops, badgeOp2)
	dOwn.Order = append(dOwn.Order, 4)
	dOwn.OpOwners = append(dOwn.OpOwners, 2)
	if _, reason := resolveLayer(dOwn, LayerSpec{
		ElementID: "badge", BoxIndex: 2, Matrix: layerTranslate(4, 0), Alpha: 1,
	}, 1, nil, nil); reason != "op inside a blend group" {
		t.Fatalf("blended element reason = %q, want %q", reason, "op inside a blend group")
	}

	// 3. A blend group op sitting in paint order between firstRank and lastRank
	//    must reject the interleaved layer.
	dInterleaved := makeDisplay(80, 50)
	dInterleaved.Order = []int{0, 1, 3, 2}
	if _, reason := resolveLayer(dInterleaved, specs[0], 1, nil, nil); reason != "non-contiguous ops" {
		t.Fatalf("interleaved blend group reason = %q, want %q", reason, "non-contiguous ops")
	}

	// 4. A later overlapping blend group still rejects BeginViewportLayers.
	dOverlap := makeDisplay(25, 20)
	renderer.DemoteLayers()
	renderer.BeginViewportLayers(dOverlap, 2, specs, 0)
	if renderer.LayersActive(dOverlap) {
		t.Fatal("viewport path promoted a layer covered by a later blend group")
	}
	if got := renderer.Layers().Reasons[layerReasonCovered]; got != 1 {
		t.Fatalf("covered reason = %d, want 1 (%v)", got, renderer.Layers().Reasons)
	}
}
