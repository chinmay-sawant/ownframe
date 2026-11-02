package page

import "slices"

// SetFocusScope restricts keyboard traversal to ids; an empty list clears it.
func SetFocusScope(p *Page, ids []string) {
	p.ff.scope = slices.Clone(ids)
}

// RequestFocus queues focus until the next form sync, when new controls exist.
// Empty id blurs. An absent or disabled control clears the old focus.
func RequestFocus(p *Page, id string) {
	p.ff.requestFocus, p.ff.hasFocusRequest = id, true
	p.markFormDirty()
}

func (p *Page) inFocusScope(id string) bool {
	return len(p.ff.scope) == 0 || slices.Contains(p.ff.scope, id)
}

func (p *Page) applyFocusRequest() {
	if !p.ff.hasFocusRequest {
		return
	}
	p.ff.hasFocusRequest = false
	fs := p.formState()
	c, ok := fs.ByID[p.ff.requestFocus]
	if !ok || c.Disabled || !p.inFocusScope(c.ID) {
		fs.FocusID = ""
		fs.ClearRange()
		return
	}
	fs.TakeFocus(c.ID, c, true)
}
