package console

import "github.com/chinmay-sawant/ownframe/examples/list"

// rowH matches .row: 12px padding on top and bottom plus two 16px text
// lines (12+12+16+16).
const (
	rowH         = 56
	listOverscan = 10
)

func (a *App) scrollWindow(offsetY, viewH int) bool {
	if viewH <= 0 {
		viewH = DefaultHeight
	}

	a.offsetY = offsetY
	a.viewH = viewH
	moved := a.sliceTickets(a.filtered())
	if !moved {
		return false
	}

	a.page.SetData(&a.view)

	return true
}

func (a *App) sliceTickets(all []Ticket) bool {
	viewH := a.viewH
	if viewH <= 0 {
		viewH = DefaultHeight
	}

	w := list.Window(a.offsetY, viewH, a.boardY(), rowH, len(all), listOverscan)
	moved := w.Start != a.winStart || w.Count != a.winCount
	a.winStart = w.Start
	a.winCount = w.Count
	a.view.ListTop = w.Top
	a.view.ListBottom = w.Bottom
	a.view.Tickets = all[w.Start : w.Start+w.Count]

	return moved
}

func (a *App) boardY() int {
	for _, b := range a.page.Boxes() {
		if b.ID == "board" {
			return int(b.Y)
		}
	}

	return 0
}
