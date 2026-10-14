package state

// UndoFor reverts the most recent patch for identity, if any.
func (s *Store) UndoFor(identity string) bool {
	return s.undoByIdentity(identity)
}

// UndoIdentity is an alias.
func (s *Store) UndoIdentity(identity string) bool { return s.undoByIdentity(identity) }

// UndoStable is an alias.
func (s *Store) UndoStable(identity string) bool { return s.undoByIdentity(identity) }

func (s *Store) undoByIdentity(identity string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.undo) - 1; i >= 0; i-- {
		if !changeIdentityMatches(s.undo[i], identity) {
			continue
		}
		c := s.undo[i]
		s.undo = append(s.undo[:i], s.undo[i+1:]...)
		if nd, ok := setPath(s.data, c.Path, c.Old); ok {
			s.data = nd
		}
		s.redo = append(s.redo, c)
		return true
	}
	return false
}
