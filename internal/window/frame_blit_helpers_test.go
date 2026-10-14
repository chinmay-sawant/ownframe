package window

import "image"

// stripes is n disjoint 5x5 rects in a row, for the cap cases.
func stripes(n int) []image.Rectangle {
	rects := make([]image.Rectangle, n)
	for i := range rects {
		rects[i] = image.Rect(i*10, 0, i*10+5, 5)
	}

	return rects
}

// planRects turns rects into repaintRect plans in order.
func planRects(rects []image.Rectangle) []repaintPlan {
	plans := make([]repaintPlan, len(rects))
	for i, r := range rects {
		plans[i] = repaintPlan{mode: repaintRect, rect: r}
	}

	return plans
}
