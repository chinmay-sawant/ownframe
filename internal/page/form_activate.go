package page

// activate toggles the control id and reports whether its value, checked
// state, or selection changed. A disabled control does not change.
func (p *Page) activate(id string) bool {
	c, ok := p.control(id)
	if !ok || c.Disabled {
		return false
	}

	changed := false
	switch {
	case c.Type == "checkbox":
		c.Checked = !c.Checked
		changed = true
	case c.Type == "radio":
		changed = !c.Checked
		c.Checked = true
		p.uncheckRadios(c.Name, id)
	case c.Tag == "select":
		before := selectedIndex(c.Options)
		old := c.Value
		c = nextOption(c)
		changed = old != c.Value || before != selectedIndex(c.Options)
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
	if name == "" || p.formState() == nil {
		return
	}

	for id, other := range p.formState().byID {
		if id == keep || other.Type != "radio" || other.Name != name {
			continue
		}

		other.Checked = false
		p.formState().byID[id] = other
	}
}
