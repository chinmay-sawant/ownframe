package page

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
)

// paintFallback stores a bitmap for a page the vector path declined. The
// picture is a replay of that same list. blinkless does not paint a page
// bitmap of its own. The hit map is built here too, so a fixed box still
// wins over the content it covers on the bitmap path.
func (p *Page) paintFallback(display *layout.Display, drawStart time.Time, track bool) error {
	ps := draw.Clock(track)
	img, err := draw.Picture(display)
	if err != nil {
		return err
	}
	draw.Note(track, &p.stats.paintTime, ps)

	boxes := display.Boxes
	p.noteCommit(drawStart)
	p.draw.CommitBitmap(display, img)
	p.setBoxes(boxes)
	p.hits = hitList(boxes, fixedBoxes(display))
	p.generation++
	p.markFull()
	p.lastPatch = DisplayPatch{
		Changed: len(display.Ops),
		Added:   len(display.Ops),
		Regions: 1,
		Full:    true,
	}
	if track {
		p.recordFull(len(boxes))
	}
	p.applyPending()
	p.runAfterRedraw()

	return nil
}
