package window

import (
	"image"
	"slices"
	"testing"
)

func TestPlanRepaintsPicksTheWork(t *testing.T) {
	t.Parallel()

	state := partialState{tiles: newTileGrid(100, 100), gen: 1}
	badSize := state
	badSize.tiles = newTileGrid(80, 100)
	full := image.Rect(0, 0, 100, 100)
	rebuild := []repaintPlan{{mode: repaintBuffer, rect: full}}
	blit := []repaintPlan{{mode: repaintBlit, rect: full}}
	one := []image.Rectangle{image.Rect(10, 10, 20, 20)}
	two := []image.Rectangle{one[0], image.Rect(60, 60, 80, 80)}
	overflow := make([]image.Rectangle, maxRepaintRects+1)
	for i := range overflow {
		overflow[i] = image.Rect(i*10, 0, i*10+5, 5)
	}

	cases := []struct {
		name  string
		st    partialState
		gen   uint64
		rects []image.Rectangle
		ok    bool
		want  []repaintPlan
	}{
		{"first frame", partialState{}, 1, one, true, rebuild},
		{"one rect", state, 2, one, true,
			[]repaintPlan{{mode: repaintRect, rect: one[0]}}},
		{"two rects", state, 2, two, true,
			[]repaintPlan{{mode: repaintRect, rect: two[0]}, {mode: repaintRect, rect: two[1]}}},
		{"full rect", state, 2, []image.Rectangle{full}, true, rebuild},
		{"overflow", state, 2, overflow, true, rebuild},
		{"empty", state, 1, nil, true, blit},
		{"idle", state, 1, nil, false, blit},
		{"stale idle", state, 2, nil, false, rebuild},
		{"size change", badSize, 2, one, true, rebuild},
		{"clamped", state, 2, []image.Rectangle{image.Rect(-20, -20, 30, 30)}, true,
			[]repaintPlan{{mode: repaintRect, rect: image.Rect(0, 0, 30, 30)}}},
		{"away", state, 2, []image.Rectangle{image.Rect(-30, -30, -20, -20)}, true, rebuild},
	}

	for _, c := range cases {
		got := planRepaints(100, 100, c.st, c.gen, c.rects, c.ok)
		if !slices.Equal(got, c.want) {
			t.Fatalf("%s: plans = %v, want %v", c.name, got, c.want)
		}
	}
}
