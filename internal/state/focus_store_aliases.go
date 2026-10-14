package state

// Aliases for likely method names.
func (s *Store) Focus() FocusState             { return s.FocusStateValue() }
func (s *Store) SetFocus(f FocusState)         { s.SetFocusStateValue(f) }
func (s *Store) Scroll() ScrollState           { return s.ScrollStateValue() }
func (s *Store) SetScroll(sc ScrollState)      { s.SetScrollStateValue(sc) }
func (s *Store) FocusState() FocusState        { return s.FocusStateValue() }
func (s *Store) SetFocusState(f FocusState)    { s.SetFocusStateValue(f) }
func (s *Store) ScrollState() ScrollState      { return s.ScrollStateValue() }
func (s *Store) SetScrollState(sc ScrollState) { s.SetScrollStateValue(sc) }
func (s *Store) UIState() UIState              { return s.UIStateValue() }
func (s *Store) SetUIState(u UIState)          { s.SetUIStateValue(u) }
func (s *Store) Caret() int                    { return s.FocusStateValue().Caret }
func (s *Store) SetCaret(v int)                { f := s.FocusStateValue(); f.Caret = v; s.SetFocusStateValue(f) }
func (s *Store) Selection() (int, int) {
	f := s.FocusStateValue()
	return f.SelectionStart, f.SelectionEnd
}
func (s *Store) SetSelection(a, b int) {
	f := s.FocusStateValue()
	f.SelectionStart, f.SelectionEnd = a, b
	s.SetFocusStateValue(f)
}
func (s *Store) KeepFocusStateExport(prev, next FocusState) FocusState {
	return KeepFocusState(prev, next)
}
func (s *Store) PreserveCaretExport(prev, next FocusState) FocusState {
	return PreserveCaret(prev, next)
}
