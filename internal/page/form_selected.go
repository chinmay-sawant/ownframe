package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// FormSelected reports whether id is the focused control and its whole
// value is selected. It is false when no control is focused, when the
// page has no form yet, or when another control is focused.
func (p *Page) FormSelected(id string) bool {
	fs := p.formState()
	if fs == nil || fs.focusID != id {
		return false
	}

	c, ok := p.control(id)

	return form.WholeSelected(form.FocusView{
		FocusID: fs.focusID,
		Caret:   fs.caret,
		Anchor:  fs.anchor,
		All:     fs.all,
		Control: c,
		Present: ok,
	}, id)
}
