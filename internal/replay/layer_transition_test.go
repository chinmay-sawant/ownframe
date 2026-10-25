package replay

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

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
