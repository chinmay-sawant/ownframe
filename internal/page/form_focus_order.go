package page

// focusOrder is the tab order: the document order of the control boxes,
// with disabled controls and tabindex="-1" left out.
func (p *Page) focusOrder() []string {
	seen := map[string]bool{}
	out := []string{}
	for _, b := range p.boxes {
		id := b.ID
		if id == "" || seen[id] {
			continue
		}

		c, ok := p.control(id)
		if !ok || c.Disabled || c.TabIndex < 0 {
			continue
		}

		seen[id] = true
		out = append(out, id)
	}

	return out
}
