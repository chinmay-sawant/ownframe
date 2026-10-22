package page

// Invalidate marks the region the element id changed since the previous
// call. An empty id, or an id with no box, dirties the whole frame. A
// known id takes its box, the boxes inside it, and its previous rect. A
// handler calls it for the ids it changed before the next Redraw. A page
// that never calls it stays correct: Redraw diffs the display lists.
func (p *Page) Invalidate(id string) {
	if id == "" || !p.hasBox(id) {
		p.dirty.MarkFull()

		return
	}

	p.markPending(id)
}

// markPending marks the id's box now and its box after the next Redraw, so
// an element that grew repaints the space it used and the space it now
// uses.
func (p *Page) markPending(id string) {
	if id == "" {
		return
	}

	p.markID(id)

	p.dirty.Pend(id)
}

// markPair marks the boxes of two element ids, one before and one after a
// state change.
func (p *Page) markPair(oldID, newID string) {
	p.markPending(oldID)
	p.markPending(newID)
}

// markID marks the box an id names, or the whole frame when no box has it.
func (p *Page) markID(id string) {
	if id == "" {
		return
	}

	if !p.hasBox(id) {
		p.dirty.MarkFull()

		return
	}

	p.markElement(id)
}

// markElement marks the id's box, the boxes inside it, and the previous
// rect for the same id, so an element that moved repaints the space it left.
func (p *Page) markElement(id string) {
	r := p.growBox(boxRect(p.boxByID(id)))
	p.dirty.MarkRect(r)
	p.dirty.MarkRect(p.dirty.Prev(id))
	p.dirty.Note(id, r)
}

// markToggle marks a checkbox or radio, its label box when the layout has
// one, and a halo that covers an inline label it does not.
func (p *Page) markToggle(id string) {
	p.markID(id)

	ctl := boxRect(p.boxByID(id))
	p.dirty.MarkRect(ctl.Inset(-labelGap))

	for _, b := range p.boxes {
		if b.Tag == "label" && touches(ctl, boxRect(b), labelGap) {
			p.dirty.MarkRect(boxRect(b))
		}
	}
}
