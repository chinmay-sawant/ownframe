package flappy

import "github.com/chinmay-sawant/ownframe"

// parts caches the display operations the paint step moves. A Redraw
// replaces the display list, so bind rebuilds the cache on a new
// generation.
type parts struct {
	bird    [birdCount]*ownframe.DisplayOp
	pipes   [pipeSlots][pipePartCount]*ownframe.DisplayOp
	clouds  [cloudMax]*ownframe.DisplayOp
	stripes [stripeMax]*ownframe.DisplayOp

	board     *ownframe.DisplayOp
	boardRect rect

	score  *ownframe.DisplayOp
	title  *ownframe.DisplayOp
	hint   *ownframe.DisplayOp
	over   *ownframe.DisplayOp
	oscore *ownframe.DisplayOp
	obest  *ownframe.DisplayOp
	again  *ownframe.DisplayOp
}

// bind caches the operations the paint step changes.
func (a *App) bind(d *ownframe.Display) {
	boxes := a.page.Boxes()

	for i, id := range birdIDs {
		a.parts.bird[i] = fillAt(d, boxes, id)
	}

	for slot := range pipeSlots {
		for i, part := range pipeParts {
			a.parts.pipes[slot][i] = fillAt(d, boxes, pipeID(slot, part))
		}
	}

	for i, id := range cloudIDs {
		a.parts.clouds[i] = fillAt(d, boxes, id)
	}

	for i, id := range stripeIDs {
		a.parts.stripes[i] = fillAt(d, boxes, id)
	}

	a.parts.score = textAt(d, boxes, "t-score")
	a.parts.title = textAt(d, boxes, "t-title")
	a.parts.hint = textAt(d, boxes, "t-hint")
	a.parts.over = textAt(d, boxes, "t-over")
	a.parts.oscore = textAt(d, boxes, "t-oscore")
	a.parts.obest = textAt(d, boxes, "t-obest")
	a.parts.again = textAt(d, boxes, "t-again")

	a.bindBoard(d, boxes)
	a.bound = a.page.Generation()
}

// bindBoard caches the panel fill and its hit-test box in CSS pixels, so
// showing it again restores its place.
func (a *App) bindBoard(d *ownframe.Display, boxes []ownframe.Box) {
	a.parts.board = fillAt(d, boxes, "board")

	if box, ok := boxByID(boxes, "board"); ok {
		a.parts.boardRect = rect{box.X, box.Y, box.W, box.H}
	}
}
