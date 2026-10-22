package page

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// commitDisplay installs a replayable list and its dirty region.
func (p *Page) commitDisplay(display *Display, executed string, drawStart time.Time, state render.State) {
	p.dirtyFromDisplay(p.draw.Display, display)
	p.noteCommit(drawStart)
	p.draw.CommitList(display)
	p.setBoxes(display.Boxes)
	p.hits = hitList(display.Boxes, fixedBoxes(display))
	p.generation++
	p.remember(state)
	p.commitAccessibility(executed)
	p.applyPending()
	p.runAfterRedraw()
}

func (p *Page) noteCommit(drawStart time.Time) {
	p.stats.layouts++
	p.stats.repaints++
	p.stats.lastDraw = time.Since(drawStart)
}
