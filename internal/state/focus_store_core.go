package state

func (s *Store) FocusStateValue() FocusState {
	if s == nil {
		return FocusState{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.focus
}
func (s *Store) SetFocusStateValue(f FocusState) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.focus = KeepFocusState(s.focus, f)
}
func (s *Store) ScrollStateValue() ScrollState {
	if s == nil {
		return ScrollState{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scroll
}
func (s *Store) SetScrollStateValue(sc ScrollState) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scroll = KeepScrollState(s.scroll, sc)
}
func (s *Store) UIStateValue() UIState {
	return UIState{Focus: s.FocusStateValue(), Scroll: s.ScrollStateValue()}
}
func (s *Store) SetUIStateValue(u UIState) {
	s.SetFocusStateValue(u.Focus)
	s.SetScrollStateValue(u.Scroll)
}
