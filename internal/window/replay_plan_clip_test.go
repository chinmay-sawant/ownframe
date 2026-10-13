package window

import (
	"image"
	"slices"
	"testing"
)

// TestPlanRepaintsClipsToTheDock checks the dock clip in the planner: hidden
// damage plans nothing, a crossing rect trims, and a grown clip repaints the
// strip the dock hid.
func TestPlanRepaintsClipsToTheDock(t *testing.T) {
	t.Parallel()

	state := partialState{tiles: newTileGrid(100, 100), gen: 1, clipX: 100}
	opened := state
	opened.clipX = 50
	blit := []repaintPlan{{mode: repaintBlit, rect: image.Rect(0, 0, 100, 100)}}
	one := []image.Rectangle{image.Rect(10, 10, 20, 20)}
	strip := repaintPlan{mode: repaintRect, rect: image.Rect(50, 0, 100, 100)}

	cases := []struct {
		name  string
		st    partialState
		gen   uint64
		rects []image.Rectangle
		ok    bool
		clip  int
		want  []repaintPlan
	}{
		{"hidden", state, 2, []image.Rectangle{image.Rect(60, 60, 80, 80)}, true, 50, blit},
		{"clipped", state, 2, []image.Rectangle{image.Rect(20, 20, 40, 40)}, true, 30,
			[]repaintPlan{{mode: repaintRect, rect: image.Rect(20, 20, 30, 40)}}},
		{"strip after close", opened, 1, nil, false, 100, []repaintPlan{strip}},
		{"strip joins damage", opened, 2, one, true, 100,
			[]repaintPlan{{mode: repaintRect, rect: one[0]}, strip}},
	}

	for _, c := range cases {
		got := planRepaints(100, 100, c.st, c.gen, c.rects, c.ok, c.clip)
		if !slices.Equal(got, c.want) {
			t.Fatalf("%s: plans = %v, want %v", c.name, got, c.want)
		}
	}
}
