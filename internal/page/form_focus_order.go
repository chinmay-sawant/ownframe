package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// focusOrder is the tab order: the document order of the control boxes,
// with disabled controls and tabindex="-1" left out.
func (p *Page) focusOrder() []string {
	ids := make([]string, 0, len(p.boxes))
	for _, b := range p.boxes {
		ids = append(ids, b.ID)
	}

	return form.Order(ids, p.control)
}
