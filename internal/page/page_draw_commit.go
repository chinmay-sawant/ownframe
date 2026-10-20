package page

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// commitDisplay installs a replayable list and its dirty region.
func (p *Page) commitDisplay(display *Display, executed string, drawStart time.Time, state render.State) {
	p.dirtyFromDisplay(p.display, display)
	p.stats.layouts++
	p.stats.repaints++
	p.stats.lastDraw = time.Since(drawStart)
	p.img = nil
	p.display = display
	p.fallbackDisplay = nil
	p.setBoxes(display.Boxes)
	p.generation++
	p.remember(state)
	p.commitAccessibility(executed)
	p.applyPending()
}
