package page

import "testing"

// Two changes far apart must keep two list entries instead of unioning
// into their bounding box.
func TestInvalidateTwoIdsKeepsTwoRects(t *testing.T) {
	t.Parallel()

	p := bedPage(t, `<p id="a">a</p><div style="height:240px"></div><p id="b">b</p>`, nil)
	p.TakeDirtyRects()

	p.Invalidate("a")
	p.Invalidate("b")

	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 2 {
		t.Fatalf("rects %v ok %v", rects, ok)
	}
}
