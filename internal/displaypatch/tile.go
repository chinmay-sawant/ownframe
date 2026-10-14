package displaypatch

import (
	"image"
	"sort"
)

// TileDamage merges damaged rectangles at operation boundaries.
type TileDamage struct {
	Rects []image.Rectangle
}

// Merge records r and coalesces overlapping or touching rectangles.
func (t *TileDamage) Merge(r image.Rectangle) {
	if r.Empty() {
		return
	}
	if t.Rects == nil {
		t.Rects = []image.Rectangle{}
	}
	t.Rects = append(t.Rects, r)
	t.coalesce()
}

func (t *TileDamage) coalesce() {
	if len(t.Rects) < 2 {
		return
	}
	sort.Slice(t.Rects, func(i, j int) bool {
		if t.Rects[i].Min.X != t.Rects[j].Min.X {
			return t.Rects[i].Min.X < t.Rects[j].Min.X
		}
		return t.Rects[i].Min.Y < t.Rects[j].Min.Y
	})
	out := make([]image.Rectangle, 0, len(t.Rects))
	for _, r := range t.Rects {
		if len(out) > 0 && (out[len(out)-1].Overlaps(r) || touching(out[len(out)-1], r)) {
			out[len(out)-1] = out[len(out)-1].Union(r)
		} else {
			out = append(out, r)
		}
	}
	t.Rects = out
}

func touching(a, b image.Rectangle) bool {
	if a.Max.X == b.Min.X || b.Max.X == a.Min.X {
		return a.Min.Y < b.Max.Y && b.Min.Y < a.Max.Y
	}
	if a.Max.Y == b.Min.Y || b.Max.Y == a.Min.Y {
		return a.Min.X < b.Max.X && b.Min.X < a.Max.X
	}
	return false
}
