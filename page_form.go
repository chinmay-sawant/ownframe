package ownframe

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// Form reads and writes control values and focus.
type Form struct{ page *Page }

// Form returns control values and focus for this page.
func (p *Page) Form() Form { return Form{page: p} }

// FormValue returns the current value of id, or "" when id is absent.
func (f Form) FormValue(id string) string {
	inner := innerPage(f.page)
	if inner == nil {
		return ""
	}

	return inner.FormValue(id)
}

// SetFormValue stores value for id and does not draw.
func (f Form) SetFormValue(id, value string) {
	inner := innerPage(f.page)
	if inner == nil {
		return
	}
	inner.SetFormValue(id, value)
}

// FormChecked reports whether id is checked.
func (f Form) FormChecked(id string) bool {
	inner := innerPage(f.page)
	if inner == nil {
		return false
	}

	return inner.FormChecked(id)
}

// SetFormChecked sets a checkbox or radio and does not draw.
func (f Form) SetFormChecked(id string, checked bool) {
	inner := innerPage(f.page)
	if inner == nil {
		return
	}
	inner.SetFormChecked(id, checked)
}

// FormSelected reports whether id has a whole-value selection.
func (f Form) FormSelected(id string) bool {
	inner := innerPage(f.page)
	if inner == nil {
		return false
	}

	return inner.FormSelected(id)
}

// Focus focuses the control id and draws. An empty id blurs.
func (f Form) Focus(ctx context.Context, id string) error {
	inner := innerPage(f.page)
	if inner == nil {
		return ErrNilPage
	}

	return inner.Focus(ctx, id)
}

// FocusID returns the focused control id, or "" when none is focused.
func (f Form) FocusID() string {
	inner := innerPage(f.page)
	if inner == nil {
		return ""
	}

	return inner.FocusID()
}

// FocusNext focuses the next control in document order and wraps.
func (f Form) FocusNext(ctx context.Context) error {
	inner := innerPage(f.page)
	if inner == nil {
		return ErrNilPage
	}

	return inner.FocusNext(ctx)
}

// FocusPrev focuses the previous control in document order and wraps.
func (f Form) FocusPrev(ctx context.Context) error {
	inner := innerPage(f.page)
	if inner == nil {
		return ErrNilPage
	}

	return inner.FocusPrev(ctx)
}

// SetFocusScope limits Tab traversal to the given controls in document order.
// No IDs removes the scope. Call on the page loop and clear it when closing a modal.
func (f Form) SetFocusScope(ids ...string) {
	inner := innerPage(f.page)
	if inner == nil {
		return
	}
	page.SetFocusScope(inner, ids)
}

// RequestFocus applies focus during the next redraw, after new controls exist.
// An empty, missing, or disabled ID clears focus. This method does not draw.
func (f Form) RequestFocus(id string) {
	inner := innerPage(f.page)
	if inner == nil {
		return
	}
	page.RequestFocus(inner, id)
}
