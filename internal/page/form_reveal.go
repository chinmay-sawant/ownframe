package page

func (p *Page) revealFocus(id string) {
	box := p.boxByID(id)
	if p.isFixed(box) {
		// A box the window pins to the viewport is always visible, so
		// revealing it must not scroll the page away from the content.
		return
	}
	x, y := p.ScrollOffset()
	if box.Y < float64(y) {
		y = int(box.Y)
	} else if box.Y+box.H > float64(y+p.height) {
		y = int(box.Y+box.H) - p.height + 1
	}
	if box.X < float64(x) {
		x = int(box.X)
	} else if box.X+box.W > float64(x+p.width) {
		x = int(box.X+box.W) - p.width + 1
	}
	p.ScrollTo(max(0, x), max(0, y))
}
