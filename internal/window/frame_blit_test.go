package window

import (
	"image"
	"slices"
	"testing"
)

func TestBlitPlanDecidesTheFrame(t *testing.T) {
	t.Parallel()

	full := image.Rect(0, 0, 100, 100)
	one := image.Rect(10, 10, 30, 30)
	two := image.Rect(60, 60, 80, 80)
	near := image.Rect(20, 20, 40, 40)
	wide := image.Rect(35, 35, 55, 55)
	chrome := image.Rect(0, 0, 8, 100)
	blit := []repaintPlan{{mode: repaintBlit, rect: full}}
	buffer := []repaintPlan{{mode: repaintBuffer, rect: full}}

	cases := []struct {
		name   string
		plans  []repaintPlan
		chrome []image.Rectangle
		full   bool
		rects  []image.Rectangle
	}{
		{"no plans", nil, nil, false, nil},
		{"blit only", blit, nil, false, nil},
		{"blit keeps chrome", blit, []image.Rectangle{chrome}, false,
			[]image.Rectangle{chrome}},
		{"one rect", planRects([]image.Rectangle{one}), nil, false,
			[]image.Rectangle{one}},
		{"two rects", planRects([]image.Rectangle{one, two}), nil, false,
			[]image.Rectangle{one, two}},
		{"overlapping rects", planRects([]image.Rectangle{one, near}), nil, false,
			[]image.Rectangle{one.Union(near)}},
		{"chrome unions a rect", planRects([]image.Rectangle{one}), []image.Rectangle{near}, false,
			[]image.Rectangle{one.Union(near)}},
		{"overlap through a third", planRects([]image.Rectangle{one, near, wide}), nil, false,
			[]image.Rectangle{one.Union(near).Union(wide)}},
		{"empty rects drop", planRects([]image.Rectangle{image.ZR, one}),
			[]image.Rectangle{image.ZR}, false, []image.Rectangle{one}},
		{"buffer plan", buffer, []image.Rectangle{chrome}, true, nil},
		{"at the cap", planRects(stripes(8)), nil, false, stripes(8)},
		{"past the cap", planRects(stripes(9)), nil, false,
			[]image.Rectangle{image.Rect(0, 0, 85, 5)}},
	}

	for _, c := range cases {
		full, rects := blitPlan(c.plans, c.chrome)
		if full != c.full {
			t.Fatalf("%s: full = %v, want %v", c.name, full, c.full)
		}

		if !slices.Equal(rects, c.rects) {
			t.Fatalf("%s: rects = %v, want %v", c.name, rects, c.rects)
		}
	}
}
