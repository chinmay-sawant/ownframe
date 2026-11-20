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
