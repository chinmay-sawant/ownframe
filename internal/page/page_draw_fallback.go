package page

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
)

// paintFallback stores a bitmap for a page the vector path declined. The
// picture is a replay of that same list. blinkless does not paint a page
// bitmap of its own.
func (p *Page) paintFallback(display *layout.Display, drawStart time.Time, track bool) error {
	ps := time.Now()
	img, err := bitmap.Paint(display)
	if err != nil {
		return err
	}
	if track {
		p.stats.paintTime = time.Since(ps)
	}

	boxes := display.Boxes
	p.stats.layouts++
	p.stats.repaints++
	p.stats.lastDraw = time.Since(drawStart)
	p.img = img
	p.display = nil
	p.fallbackDisplay = display
	p.setBoxes(boxes)
	p.generation++
	p.markFull()
	if track {
		p.recordFull(len(boxes))
	}
	p.applyPending()

	return nil
}
