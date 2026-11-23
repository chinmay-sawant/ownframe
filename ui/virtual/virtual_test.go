package virtual

import (
	"math"
	"testing"
)

func TestWindowEmptyAndInvalid(t *testing.T) {
	cases := []struct {
		name      string
		scrollY   float64
		viewportH float64
		rowH      float64
		total     int
		overscan  int
	}{
		{"zero rows", 0, 400, 20, 0, 2},
		{"negative rows", 10, 400, 20, -5, 2},
		{"zero row height", 10, 400, 0, 100, 2},
		{"negative row height", 10, 400, -20, 100, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Window(tc.scrollY, tc.viewportH, tc.rowH, tc.total, tc.overscan)
			if got != (Range{}) {
				t.Fatalf("Window = %+v, want zero Range", got)
			}
		})
	}
}

func TestWindowScrollPositionsAndSpacers(t *testing.T) {
	// 1000 rows * 20px = 20000px total; 200px viewport = 10 visible rows; overscan = 2.
	top := Window(0, 200, 20, 1000, 2)
	if top.Start != 0 || top.End != 12 || top.TopSpacer != 0 || top.BottomSpacer != float64(988*20) || top.TotalHeight != 20000 {
		t.Fatalf("top = %+v", top)
	}

	// Negative scroll clamps to 0.
	neg := Window(-150, 200, 20, 1000, 2)
	if neg != top {
		t.Fatalf("neg = %+v, want %+v", neg, top)
	}

	// Middle scroll: scrollY = 1000 -> row 50..60 visible; with overscan 2 -> [48, 62).
	mid := Window(1000, 200, 20, 1000, 2)
	if mid.Start != 48 || mid.End != 62 || mid.TopSpacer != 960 || mid.BottomSpacer != float64(938*20) {
		t.Fatalf("mid = %+v", mid)
	}

	// Bottom overflow clamps to maxScroll (19800 -> rows [988, 1000)).
	bot := Window(99999, 200, 20, 1000, 2)
	if bot.Start != 988 || bot.End != 1000 || bot.BottomSpacer != 0 || bot.TopSpacer != float64(988*20) {
		t.Fatalf("bot = %+v", bot)
	}

	// Invariant: TopSpacer + rendered rows + BottomSpacer == TotalHeight.
	for _, r := range []Range{top, mid, bot} {
		sum := r.TopSpacer + float64(r.End-r.Start)*20 + r.BottomSpacer
		if sum != r.TotalHeight {
			t.Fatalf("spacer invariant failed: %+v (sum=%v)", r, sum)
		}
	}
}

func TestWindowSmallListAndResize(t *testing.T) {
	// 5 rows * 24px = 120px inside 400px viewport -> renders all 5 rows, 0 spacers.
	small := Window(50, 400, 24, 5, 3)
	if small.Start != 0 || small.End != 5 || small.TopSpacer != 0 || small.BottomSpacer != 0 || small.TotalHeight != 120 {
		t.Fatalf("small = %+v", small)
	}

	// Viewport grows from 200px to 400px at scrollY=1000.
	resized := Window(1000, 400, 20, 1000, 1)
	if resized.Start != 49 || resized.End != 71 {
		t.Fatalf("resized = %+v, want [49, 71)", resized)
	}
}

func TestWindowNonFiniteInputs(t *testing.T) {
	nan, inf := math.NaN(), math.Inf(1)
	cases := []struct {
		name                     string
		scrollY, viewportH, rowH float64
	}{
		{"nan scroll", nan, 100, 20},
		{"inf scroll", inf, 100, 20},
		{"-inf scroll", -inf, 100, 20},
		{"nan viewport", 0, nan, 20},
		{"inf viewport", 0, inf, 20},
		{"nan row", 0, 100, nan},
		{"inf row", 0, 100, inf},
		{"tiny row", 1e300, 100, 1e-300},
	}
	for _, c := range cases {
		r := Window(c.scrollY, c.viewportH, c.rowH, 50, 2)
		if r.Start < 0 || r.End < r.Start || r.End > 50 {
			t.Errorf("%s: bad range %+v", c.name, r)
		}
		for _, v := range []float64{r.TopSpacer, r.BottomSpacer, r.TotalHeight} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Errorf("%s: non-finite spacer in %+v", c.name, r)
			}
		}
	}
}
