package ownframe

import "context"

// Form reads and writes control values and focus.
type Form struct{ page *Page }

// Form returns control values and focus for this page.
func (p *Page) Form() Form { return Form{page: p} }

// FormValue returns the current value of id, or "" when id is absent.
func (f Form) FormValue(id string) string { return innerPage(f.page).FormValue(id) }

// SetFormValue stores value for id and does not draw.
func (f Form) SetFormValue(id, value string) { innerPage(f.page).SetFormValue(id, value) }

// FormChecked reports whether id is checked.
func (f Form) FormChecked(id string) bool { return innerPage(f.page).FormChecked(id) }

// SetFormChecked sets a checkbox or radio and does not draw.
func (f Form) SetFormChecked(id string, checked bool) {
	innerPage(f.page).SetFormChecked(id, checked)
}

// FormSelected reports whether id has a whole-value selection.
func (f Form) FormSelected(id string) bool { return innerPage(f.page).FormSelected(id) }

// Focus focuses the control id and draws. An empty id blurs.
func (f Form) Focus(ctx context.Context, id string) error {
	return innerPage(f.page).Focus(ctx, id)
}

// FocusID returns the focused control id, or "" when none is focused.
func (f Form) FocusID() string { return innerPage(f.page).FocusID() }

// FocusNext focuses the next control in document order and wraps.
func (f Form) FocusNext(ctx context.Context) error {
	return innerPage(f.page).FocusNext(ctx)
}

// FocusPrev focuses the previous control in document order and wraps.
func (f Form) FocusPrev(ctx context.Context) error {
	return innerPage(f.page).FocusPrev(ctx)
}
