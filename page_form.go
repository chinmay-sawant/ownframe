package ownframe

// FormValue returns the current value of id, or "" when id is absent.
func (p *Page) FormValue(id string) string { return innerPage(p).FormValue(id) }

// SetFormValue stores value for id and does not draw.
func (p *Page) SetFormValue(id, value string) { innerPage(p).SetFormValue(id, value) }

// FormChecked reports whether id is checked.
func (p *Page) FormChecked(id string) bool { return innerPage(p).FormChecked(id) }

// SetFormChecked sets a checkbox or radio and does not draw.
func (p *Page) SetFormChecked(id string, checked bool) {
	innerPage(p).SetFormChecked(id, checked)
}

// FormSelected reports whether id has a whole-value selection.
func (p *Page) FormSelected(id string) bool { return innerPage(p).FormSelected(id) }
