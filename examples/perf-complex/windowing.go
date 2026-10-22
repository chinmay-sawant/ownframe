package main

import "github.com/chinmay-sawant/ownframe/internal/page"

// Windowing is the 1.0 list recipe: SetWindowing plus SetScrollWindow
// slices app-owned rows around the viewport and pads with spacers so layout
// stays cheap. This is not a compositor pin layer. Overscan rows sit around
// the viewport; the row count pads to a stripe pair so nth-child stripes
// match the full document.
const (
	windowOverscan = 10
	windowSlack    = 5
	stripePair     = 2
)

func installWindow(p *page.Page, v view) {
	all := v.Rows
	previousStart, previousEnd := -1, -1
	p.SetWindowing(true)
	p.SetScrollWindow(func(offsetY, viewH int) bool {
		gridY := 0
		for _, b := range p.Boxes() {
			if b.ID == "grid" {
				gridY = int(b.Y)
				break
			}
		}
		first := max(0, (offsetY-gridY)/fixtureRowH)
		last := min(len(all), max(0, (offsetY+viewH-gridY+fixtureRowH-1)/fixtureRowH))
		topSafe := previousStart == 0 || first >= previousStart+windowSlack
		bottomSafe := previousEnd == len(all) || last <= previousEnd-windowSlack
		if previousStart >= 0 && topSafe && bottomSafe {
			return false
		}
		start := (offsetY-gridY)/fixtureRowH - windowOverscan
		if start < 0 {
			start = 0
		}
		count := viewH/fixtureRowH + windowOverscan*2 + 1
		if count < fixtureWindowRows {
			count = fixtureWindowRows
		}
		if start > len(all)-count {
			start = len(all) - count
		}
		if start < 0 {
			start = 0
		}
		// Keep nth-child stripes aligned with the full document.
		start -= start % stripePair
		end := min(start+count, len(all))
		if start == previousStart && end == previousEnd {
			return false
		}
		previousStart, previousEnd = start, end
		v.Rows = all[start:end]
		v.Top = start * fixtureRowH
		v.Bottom = (len(all) - end) * fixtureRowH
		p.SetData(v)
		return true
	})
}
