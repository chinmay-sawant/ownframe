package state

// Undo reverts the most recent patch, if any.
func (s *Store) Undo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.undo) == 0 {
		return
	}
	c := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]

	if nd, ok := setPath(s.data, c.Path, c.Old); ok {
		s.data = nd
	}
	s.redo = append(s.redo, c)
}

// Redo reapplies the most recently undone patch, if any.
func (s *Store) Redo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.redo) == 0 {
		return
	}
	c := s.redo[len(s.redo)-1]
	s.redo = s.redo[:len(s.redo)-1]

	if nd, ok := setPath(s.data, c.Path, c.New); ok {
		s.data = nd
	}
	s.undo = append(s.undo, c)
}
