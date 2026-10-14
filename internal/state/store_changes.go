package state

// Changes returns a copy of every recorded change.
func (s *Store) Changes() []Change {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Change, len(s.changes))
	copy(out, s.changes)
	return out
}

// CanUndo reports whether Undo has work to do.
func (s *Store) CanUndo() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.undo) > 0
}

// CanRedo reports whether Redo has work to do.
func (s *Store) CanRedo() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.redo) > 0
}
