package state

// ChangedProperties returns the property paths recorded so far.
func (s *Store) ChangedProperties() []string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.changes))
	seen := map[string]bool{}
	for _, c := range s.changes {
		if c.Path == "" || seen[c.Path] {
			continue
		}
		seen[c.Path] = true
		out = append(out, c.Path)
	}
	return out
}

// Deprecated: use ChangedProperties.
func (s *Store) ChangedProps() []string { return s.ChangedProperties() }

// Deprecated: use ChangedProperties.
func (s *Store) PropertiesChanged() []string { return s.ChangedProperties() }

// LastChangedProperties returns the paths from the most recent change.
func (s *Store) LastChangedProperties() []string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.changes) == 0 {
		return nil
	}
	c := s.changes[len(s.changes)-1]
	if c.Path == "" {
		return nil
	}
	return []string{c.Path}
}

// Deprecated: use LastChangedProperties.
func (s *Store) LastChangedProps() []string { return s.LastChangedProperties() }

// TakeChangedProperties returns a snapshot of the recorded paths.
func (s *Store) TakeChangedProperties() []string { return s.ChangedProperties() }

// Deprecated: use TakeChangedProperties.
func (s *Store) TakeChangedProps() []string { return s.TakeChangedProperties() }
