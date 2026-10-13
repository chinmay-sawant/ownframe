package runtimebehavior

// HasSaved reports whether focus state is saved.
func (f *FocusRestorer) HasSaved() bool {
	if f == nil {
		return false
	}
	return f.ElementID != ""
}

// SetSelection updates the saved selection.
func (f *FocusRestorer) SetSelection(sel [2]int) {
	if f == nil {
		return
	}
	f.Selection = sel
}

// SetCaret updates the saved caret.
func (f *FocusRestorer) SetCaret(caret int) {
	if f == nil {
		return
	}
	f.Caret = caret
}
