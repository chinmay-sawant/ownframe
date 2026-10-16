package page

// FormSelected reports whether id is the focused control and its whole
// value is selected. It is false when no control is focused, when the
// page has no form yet, or when another control is focused.
func (p *Page) FormSelected(id string) bool {
	if p.formState() == nil || p.formState().focusID != id {
		return false
	}

	c, ok := p.control(id)
	if !ok || !canEdit(c) {
		return false
	}

	if p.formState().all {
		return true
	}

	n := runeLen(c.Value)
	start, end := p.formState().bounds(n)

	return start != end && start == 0 && end == n
}
