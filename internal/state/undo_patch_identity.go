package state

import "time"

// PatchChangeFor records a change bound to a stable element identity.
func (s *Store) PatchChangeFor(identity, path string, value any) Change {
	return s.patchChangeWithIdentity(identity, path, value)
}

// PatchChangeWithIdentity is an alias.
func (s *Store) PatchChangeWithIdentity(identity, path string, value any) Change {
	return s.patchChangeWithIdentity(identity, path, value)
}

// PatchChangeWithStableID is an alias.
func (s *Store) PatchChangeWithStableID(identity, path string, value any) Change {
	return s.patchChangeWithIdentity(identity, path, value)
}

// PatchChangeElement is an alias.
func (s *Store) PatchChangeElement(identity, path string, value any) Change {
	return s.patchChangeWithIdentity(identity, path, value)
}

func (s *Store) patchChangeWithIdentity(identity, path string, value any) Change {
	c := s.PatchChange(path, value)
	c.Identity = identity
	c.ElementID = identity
	c.StableID = identity
	// Update stored change copy.
	s.mu.Lock()
	if len(s.changes) > 0 {
		s.changes[len(s.changes)-1] = c
	}
	if len(s.undo) > 0 {
		s.undo[len(s.undo)-1] = c
	}
	s.mu.Unlock()
	_ = time.Now
	return c
}
