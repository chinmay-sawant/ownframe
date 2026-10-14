package state

// LastChange returns the most recent recorded change.
func (s *Store) LastChange() (Change, bool) {
	if s == nil {
		return Change{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.changes) == 0 {
		return Change{}, false
	}
	return s.changes[len(s.changes)-1], true
}

// LastChanged is an alias for LastChange.
func (s *Store) LastChanged() (Change, bool) { return s.LastChange() }

// ClearChangedProperties clears the recorded change log.
func (s *Store) ClearChangedProperties() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.changes = s.changes[:0]
}

// ClearChangedProps is an alias for ClearChangedProperties.
func (s *Store) ClearChangedProps() { s.ClearChangedProperties() }

// Origin returns the current patch origin.
func (s *Store) Origin() Origin {
	if s == nil {
		return OriginUnknown
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.origin
}

// SetOrigin sets the current patch origin.
func (s *Store) SetOrigin(o Origin) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.origin = o
}
