package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"testing"
)

func TestParseBackdropFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		want      backdropFilter
		wantValid bool
	}{
		{
			name:      "blur and saturate",
			value:     "blur(14px) saturate(1.35)",
			want:      backdropFilter{BlurPx: 14, Saturation: 1.35},
			wantValid: true,
		},
		{
			name:      "blur saturation and refraction",
			value:     "blur(28px) saturate(1.4) refract(5px)",
			want:      backdropFilter{BlurPx: 28, Saturation: 1.4, RefractionPx: 5},
			wantValid: true,
		},
		{
			name:      "refraction only",
			value:     "refract(4px)",
			want:      backdropFilter{Saturation: 1, RefractionPx: 4},
			wantValid: true,
		},
		{
			name:      "percentage saturation",
			value:     "saturate(135%)",
			want:      backdropFilter{Saturation: 1.35},
			wantValid: true,
		},
		{name: "none", value: "none"},
		{name: "svg URL remains unsupported", value: "url(#liquid-glass)"},
		{name: "unsupported function rejects the list", value: "blur(12px) grayscale(0.5)"},
		{name: "negative blur is invalid", value: "blur(-4px)"},
		{name: "pixel blur only", value: "blur(1rem)"},
		{name: "duplicate blur is unsupported", value: "blur(4px) blur(8px)"},
		{name: "duplicate saturation is unsupported", value: "saturate(1.2) saturate(1.4)"},
		{name: "duplicate refraction is unsupported", value: "refract(2px) refract(4px)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseBackdropFilter(test.value)
			if ok != test.wantValid || got != test.want {
				t.Fatalf("parseBackdropFilter(%q) = %+v, %v; want %+v, %v",
					test.value, got, ok, test.want, test.wantValid)
			}
		})
	}
}

// TestPlanBucketsNestedGroups pins the group tree: a child group's node sits
// inside its parent's items where the child first paints, whatever order the
// boundary markers take.
func TestPlanBucketsNestedGroups(t *testing.T) {
	t.Parallel()

	outer := &layout.DisplayGroup{ID: 1, Mode: "multiply", Isolate: true}
	inner := &layout.DisplayGroup{ID: 2, Mode: "screen", Isolate: true, Parent: outer}

	rootOp := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 50, H: 20}
	first := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 10, H: 10}
	first.SetBlendGroup(outer)
	second := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 20, Y: 0, W: 10, H: 10}
	second.SetBlendGroup(inner)

	d := &layout.Display{PointsPerPixel: 1, Ops: []layout.DisplayOp{rootOp, first, second}, Order: []int{0, 1, 2}}
	plan := buildPlan(d)

	if len(plan.items) != 2 || plan.items[0].op == nil || plan.items[1].child == nil {
		t.Fatalf("root items = %v", plan.items)
	}

	outerNode := plan.items[1].child
	if outerNode.group != outer || len(outerNode.items) != 2 {
		t.Fatalf("outer node = %+v", outerNode)
	}

	innerNode := outerNode.items[1].child
	if innerNode == nil || innerNode.group != inner || len(innerNode.items) != 1 {
		t.Fatalf("inner node = %+v", innerNode)
	}

	box, ok := nodeBox(outerNode, 1)
	if !ok || box != image.Rect(-1, -1, 31, 11) {
		t.Fatalf("group box = %v, %v", box, ok)
	}
}

// TestGroupShaderCompiles pins the Kage program: the compositor's modes are
// useless if the shader does not parse.
func TestGroupShaderCompiles(t *testing.T) {
	t.Parallel()

	if _, err := ebiten.NewShader([]byte(groupBlendSource)); err != nil {
		t.Fatal(err)
	}
}

func TestParseFilterBlur(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		wantPx  float64
		wantHas bool
	}{
		{"blur(2px)", 2, true},
		{"grayscale(1) blur(4px)", 4, true},
		{"blur(3px) blur(4px)", 5, true},
		{"blur(3pt)", 4, true},
		{"blur(0px)", 0, false},
		{"blur(0)", 0, false},
		{"none", 0, false},
		{"grayscale(1)", 0, false},
		{"blur(-2px)", 0, false},
	}

	for _, tc := range tests {
		got, ok := parseFilterBlur(tc.value)
		if ok != tc.wantHas || got != tc.wantPx {
			t.Fatalf("parseFilterBlur(%q) = (%v, %v), want (%v, %v)", tc.value, got, ok, tc.wantPx, tc.wantHas)
		}
	}
}

func TestPlanFilterBlurPadsNodeBoxAndKeepsGroupUnit(t *testing.T) {
	t.Parallel()

	group := &layout.DisplayGroup{ID: 1, Filter: "blur(2px)", Isolate: true}
	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 20, Y: 20, W: 10, H: 10}
	op.SetBlendGroup(group)

	d := &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{op}, Order: []int{0},
	}
	root := buildPlan(d)
	if len(root.items) != 1 || root.items[0].child == nil {
		t.Fatalf("expected 1 child group node, got %+v", root.items)
	}
	node := root.items[0].child
	if !node.hasFilterBlur || node.filterBlurPx != 2 {
		t.Fatalf("node blur = (%v, %v), want (true, 2)", node.hasFilterBlur, node.filterBlurPx)
	}

	pad := filterBlurPad(2)
	box, ok := nodeBox(node, 1)
	wantBox := image.Rect(20-1-pad, 20-1-pad, 30+1+pad, 30+1+pad)
	if !ok || box != wantBox {
		t.Fatalf("blurred nodeBox = (%v, %v), want (%v, true)", box, ok, wantBox)
	}

	chunks := buildChunks(d, root)
	if len(chunks.items) != 1 || chunks.items[0].group != node {
		t.Fatalf("expected blurred group to stay a single chunk unit, got %+v", chunks.items)
	}

	// A rect inside the blur bleed outside the unblurred op must still touch.
	bleedRect := image.Rect(20-pad, 20, 20-1, 30)
	if !opTouches(&d.Ops[0], 1, bleedRect) {
		t.Fatalf("opTouches(%v) = false, want true for blur bleed", bleedRect)
	}
}

func TestIsolationGroupNeedsLayerOnlyWithBlendedDescendant(t *testing.T) {
	t.Parallel()

	iso := &layout.DisplayGroup{ID: 1, Isolate: true}
	child := &layout.DisplayGroup{ID: 2, Mode: "multiply", Isolate: true, Parent: iso}

	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20}
	op.SetBlendGroup(child)

	d := &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{op}, Order: []int{0},
	}
	root := buildPlan(d)
	if len(root.items) != 1 || root.items[0].child == nil {
		t.Fatalf("expected 1 isolation group node, got %+v", root.items)
	}
	isoNode := root.items[0].child
	if !isoNode.needsLayer() {
		t.Fatalf("isolation:isolate wrapping mix-blend-mode:multiply child must buffer a layer")
	}

	chunks := buildChunks(d, root)
	if len(chunks.items) != 1 || chunks.items[0].group != isoNode {
		t.Fatalf("expected isolation group with blended child to stay a single chunk unit, got %+v", chunks.items)
	}
}
