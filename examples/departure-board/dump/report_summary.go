package dump

import (
	"runtime"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// summarize fills the histogram, the tick percentiles, the bounding-box and
// damage-list areas, the rect count, and the allocation rate, then records
// the final op and box counts. The mean area and rect count average the
// frames that reported a damage list, so blit frames do not drag them down.
func (rep *Report) summarize(app *board.App, before, after runtime.MemStats) {
	ticks := make([]int64, len(rep.Frames))
	boxSum, boxN, listSum, listN, rectSum := 0, 0, 0, 0, 0
	for i, f := range rep.Frames {
		ticks[i] = f.TickNS
		rep.Kinds[f.Kind]++
		if a := rectArea(f.Rect); a > 0 {
			boxSum += a
			boxN++
			if a > rep.RectMaxArea {
				rep.RectMaxArea = a
			}
		}
		if f.Area > 0 {
			listSum += f.Area
			listN++
			rectSum += f.Rects
			if f.Area > rep.AreaMax {
				rep.AreaMax = f.Area
			}
		}
	}
	if boxN > 0 {
		rep.RectMeanArea = boxSum / boxN
	}
	if listN > 0 {
		rep.AreaMean = listSum / listN
		rep.RectsMean = float64(rectSum) / float64(listN)
	}
	rep.TickP50NS = percentile(ticks, percentP50)
	rep.TickP95NS = percentile(ticks, percentP95)
	rep.TickMaxNS = percentile(ticks, percentMax)
	if n := len(rep.Frames); n > 0 {
		rep.AllocsFrame = float64(after.TotalAlloc-before.TotalAlloc) / float64(n)
	}

	p := app.Page()
	rep.Boxes = len(p.Boxes())
	if d := p.Testing().Display(); d != nil {
		rep.Ops = len(d.Ops)
	}
}
