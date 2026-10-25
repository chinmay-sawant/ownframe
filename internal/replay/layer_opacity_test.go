package replay

import (
	"image"
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

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
