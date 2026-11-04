package components

import "github.com/chinmay-sawant/ownframe/ui"

// DotBar renders total dots with the first filledCount lit.
func DotBar(filled, total int) *ui.Element {
	if total < 0 {
		total = 0
	}
	if filled < 0 {
		filled = 0
	}
	if filled > total {
		filled = total
	}
	dots := make([]*ui.Element, 0, total)
	for i := range total {
		class := "mp-dot mp-dot--off"
		if i < filled {
			class = "mp-dot mp-dot--on"
		}
		dots = append(dots, ui.Span().Class(class))
	}
	return ui.Div().Class("mp-dots").Children(dots...)
}
