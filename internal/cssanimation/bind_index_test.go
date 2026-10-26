package cssanimation

import (
	"slices"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestBindBoxWithIndexMatchesDirect pins the shared-index contract at the
// binder: resolving through one per-display index returns the same players,
// target sets, masks, and gaps as the direct scan, across owner, subtree,
// anonymous, svg, and ownerless displays.
func TestBindBoxWithIndexMatchesDirect(t *testing.T) {
	shift := []layout.KeyframesRule{{Name: "shift", Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translate(0px)"}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "transform", Value: "translate(20px)"}}},
	}}}
	start := time.Unix(100, 0)

	animations := []layout.Animation{
		{Name: "shift", DurationMS: 1000, IterationCount: 1, FillMode: "both", PlayState: "running"},
		{Name: "tint", DurationMS: 1000, IterationCount: 1, FillMode: "both", PlayState: "running"},
	}
	newDisplay := func(owners []int) *layout.Display {
		d := indexCardDisplay(animations...)
		d.Keyframes = append(append([]layout.KeyframesRule{}, shift...),
			indexColorRule("tint", "background-color", "#ff0000", "#00ff00")...)
		d.OpOwners = owners
		d.Boxes[1].Animations = animations[:1]
		return d
	}

	displays := map[string]*layout.Display{
		"owned":     newDisplay([]int{0, 0, 1}),
		"ownerless": newDisplay(nil),
	}
	for name, d := range displays {
		idx := cssmotion.NewIndex(d)
		for boxIndex := range d.Boxes {
			if len(d.Boxes[boxIndex].Animations) == 0 {
				continue
			}
			want, wantGaps, err := BindBox(d, d.Boxes[boxIndex], boxIndex, start)
			if err != nil {
				t.Fatalf("%s box %d direct bind: %v", name, boxIndex, err)
			}
			got, gotGaps, err := BindBoxWithIndex(d, d.Boxes[boxIndex], boxIndex, start, idx)
			if err != nil {
				t.Fatalf("%s box %d indexed bind: %v", name, boxIndex, err)
			}
			compareBounds(t, name, boxIndex, want, got)
			if !slices.Equal(gotGaps, wantGaps) {
				t.Fatalf("%s box %d gaps = %v, want %v", name, boxIndex, gotGaps, wantGaps)
			}
		}
		anonymous := d.Boxes[0]
		anonymous.ID = ""
		want, wantGaps, err := BindBox(d, anonymous, -1, start)
		if err != nil {
			t.Fatalf("%s anonymous direct bind: %v", name, err)
		}
		got, gotGaps, err := BindBoxWithIndex(d, anonymous, -1, start, idx)
		if err != nil {
			t.Fatalf("%s anonymous indexed bind: %v", name, err)
		}
		compareBounds(t, name, -1, want, got)
		if !slices.Equal(gotGaps, wantGaps) {
			t.Fatalf("%s anonymous gaps = %v, want %v", name, gotGaps, wantGaps)
		}
	}
}

// TestBindBoxWithIndexMatchesDirectSVG covers the raster tier: an SVG paint
// player must resolve its image operation identically through the index.
func TestBindBoxWithIndexMatchesDirectSVG(t *testing.T) {
	d := svgBindDisplay(svgRule("paint", "fill", "#000000", "#ffffff"), "paint")
	idx := cssmotion.NewIndex(d)
	want, wantGaps, err := BindBox(d, d.Boxes[0], 0, svgStart())
	if err != nil {
		t.Fatalf("direct bind: %v", err)
	}
	got, gotGaps, err := BindBoxWithIndex(d, d.Boxes[0], 0, svgStart(), idx)
	if err != nil {
		t.Fatalf("indexed bind: %v", err)
	}
	compareBounds(t, "svg", 0, want, got)
	if !slices.Equal(gotGaps, wantGaps) {
		t.Fatalf("svg gaps = %v, want %v", gotGaps, wantGaps)
	}
}

// indexCardDisplay is one parent card (fill plus label) with a child card, for
// the shared-index equivalence test. Op ids are stable so both resolution
// paths see the same ownership.
func indexCardDisplay(animations ...layout.Animation) *layout.Display {
	return &layout.Display{
		Width: 200, Height: 120, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{ID: 11, Kind: layout.DisplayOpFillRect,
				X: 10, Y: 10, W: 100, H: 60, R: .1, G: .2, B: .3, Alpha: 1},
			{ID: 12, Kind: layout.DisplayOpText,
				X: 20, Y: 30, W: 40, H: 12, R: 1, G: 1, B: 1, Alpha: 1},
			{ID: 13, Kind: layout.DisplayOpFillRect,
				X: 30, Y: 40, W: 40, H: 20, R: .4, G: .5, B: .6, Alpha: 1},
		},
		Boxes: []layout.Box{
			{ID: "parent", X: 10, Y: 10, W: 100, H: 60,
				OperationIDs: []uint64{11, 12, 13},
				Animations:   animations, ParentIndex: -1},
			{ID: "child", X: 30, Y: 40, W: 40, H: 20,
				OperationIDs: []uint64{13}, ParentIndex: 0},
		},
	}
}

// indexColorRule builds one linear two-stop animation rule.
func indexColorRule(name, property, from, to string) []layout.KeyframesRule {
	return []layout.KeyframesRule{{Name: name, Keyframes: []layout.Keyframe{
		{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: from}}},
		{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: property, Value: to}}},
	}}}
}

func compareBounds(t *testing.T, name string, boxIndex int, want, got []Bound) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s box %d players = %d, want %d", name, boxIndex, len(got), len(want))
	}
	for i := range want {
		if !slices.Equal(got[i].Indices, want[i].Indices) {
			t.Fatalf("%s box %d player %d targets = %v, want %v",
				name, boxIndex, i, got[i].Indices, want[i].Indices)
		}
		if got[i].Mask != want[i].Mask {
			t.Fatalf("%s box %d player %d mask = %d, want %d",
				name, boxIndex, i, got[i].Mask, want[i].Mask)
		}
	}
}
