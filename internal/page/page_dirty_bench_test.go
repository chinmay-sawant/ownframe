package page

import "testing"

// BenchmarkDirtyTakeSmallChanges measures one generation's page-side dirty
// path: the display diff, the mark, and the take the window reads. The
// partial metric is 1 while the change stays a partial repaint and rects
// counts the rect list the take returned.
func BenchmarkDirtyTakeSmallChanges(b *testing.B) {
	p, base, next := benchDiffPage(b)
	vals := benchDiffValues()
	mutateBenchText(next, vals)
	p.dirtyFromDisplay(base, next)
	rects, _ := p.TakeDirtyRects()
	content := p.contentRect()
	partial := 0.0

	if len(rects) > 0 && (len(rects) != 1 || rects[0] != content) {
		partial = 1
	}

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		mutateBenchText(next, vals[(i%benchDiffGens)*benchDiffCells:])
		p.dirtyFromDisplay(base, next)
		p.TakeDirtyRects()
	}

	b.ReportMetric(partial, "partial")
	b.ReportMetric(float64(len(rects)), "rects")
}
