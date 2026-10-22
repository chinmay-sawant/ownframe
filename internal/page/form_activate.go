package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// activate toggles the control id and reports whether its value, checked
// state, or selection changed. A disabled control does not change.
func (p *Page) activate(id string) bool {
	c, ok := p.control(id)
	if !ok || c.Disabled {
		return false
	}

	c, changed, clear := form.ToggleClick(c)
	if clear {
		p.uncheckRadios(c.Name, id)
	}

	p.formState().byID[id] = c

	if p.formState().focusID != id {
		p.formState().focusID = id
		if canEdit(c) {
			p.caretEnd(c)
		} else {
			p.clearRange()
		}
	} else {
		p.formState().all = false
	}

	return changed
}

func (p *Page) uncheckRadios(name, keep string) {
	if p.formState() == nil {
		return
	}

	form.UncheckRadios(p.formState().byID, name, keep)
}
