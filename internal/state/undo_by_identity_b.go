package state

// RedoFor reapplies the most recent undone patch for identity.
func (s *Store) RedoFor(identity string) bool {
	return s.redoByIdentity(identity)
}

// RedoIdentity is an alias.
func (s *Store) RedoIdentity(identity string) bool { return s.redoByIdentity(identity) }

// RedoStable is an alias.
func (s *Store) RedoStable(identity string) bool { return s.redoByIdentity(identity) }

func (s *Store) redoByIdentity(identity string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.redo) - 1; i >= 0; i-- {
		if !changeIdentityMatches(s.redo[i], identity) {
			continue
		}
		c := s.redo[i]
		s.redo = append(s.redo[:i], s.redo[i+1:]...)
		if nd, ok := setPath(s.data, c.Path, c.New); ok {
			s.data = nd
		}
		s.undo = append(s.undo, c)
		return true
	}
	return false
}

// CanUndoFor reports whether identity has undo work.
func (s *Store) CanUndoFor(identity string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.undo {
		if changeIdentityMatches(c, identity) {
			return true
		}
	}
	return false
}

// CanRedoFor reports whether identity has redo work.
func (s *Store) CanRedoFor(identity string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.redo {
		if changeIdentityMatches(c, identity) {
			return true
		}
	}
	return false
}

func changeIdentityMatches(c Change, identity string) bool {
	if identity == "" {
		return c.Identity == "" && c.ElementID == "" && c.StableID == ""
	}
	return c.Identity == identity || c.ElementID == identity || c.StableID == identity
}
