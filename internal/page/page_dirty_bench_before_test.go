package page

import (
	"image"
	"testing"
)

// oldMaxDirtyOps is the op-count cliff the diff used before the cost rule.
const oldMaxDirtyOps = 8

// oldDiffDisplay reproduces the removed rule: a diff of more than
// oldMaxDirtyOps changed operations fell back to the frame even when the
// changes sat in one small region.
func oldDiffDisplay(prev, next *Display) (image.Rectangle, bool, int) {
	dirty := image.Rectangle{}
	changed := 0
	frame := image.Rect(0, 0, next.Width, next.Height)

	for i := range prev.Ops {
		if sameOp(&prev.Ops[i], &next.Ops[i]) {
			continue
		}

		changed++
		if changed > oldMaxDirtyOps {
			return frame, true, len(next.Ops)
		}

		dirty = dirty.
			Union(opBounds(&prev.Ops[i], prev)).
			Union(opBounds(&next.Ops[i], next))
	}

	if changed == 0 {
		return image.Rectangle{}, false, 0
	}

	if dirty.Dx()*dirty.Dy() > frame.Dx()*frame.Dy()/3 {
		return frame, true, changed
	}

	return dirty, false, changed
}

// markOldDiff applies the old diff the way dirtyFromDisplay did.
func markOldDiff(p *Page, r image.Rectangle, full bool) {
	if full {
		p.markFull()

		return
	}

	p.markRect(r)
}

// BenchmarkDirtyTakeSmallChangesBefore runs the same generation through the
// old op-count cliff, so before and after share one fixture. partial is 0:
// this path fell back to the frame.
func BenchmarkDirtyTakeSmallChangesBefore(b *testing.B) {
	p, base, next := benchDiffPage(b)
	vals := benchDiffValues()
	mutateBenchText(next, vals)

	r, full, _ := oldDiffDisplay(base, next)
	markOldDiff(p, r, full)
	rects, _ := p.TakeDirtyRects()
	content := p.contentRect()
	partial := 0.0

	if len(rects) > 0 && (len(rects) != 1 || rects[0] != content) {
		partial = 1
	}

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		mutateBenchText(next, vals[(i%benchDiffGens)*benchDiffCells:])
		r, full, _ := oldDiffDisplay(base, next)
		markOldDiff(p, r, full)
		p.TakeDirtyRects()
	}

	b.ReportMetric(partial, "partial")
	b.ReportMetric(float64(len(rects)), "rects")
}
