package ownframe

// IMEContext returns the focused field's box and caret text.
func (t Testing) IMEContext() (Box, int, string, string, bool) {
	p := innerPage(t.page)
	if p == nil {
		return Box{}, 0, "", "", false
	}

	return p.IMEContext()
}
