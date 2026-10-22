package state

import "time"

// PatchChangeFor records a change bound to a stable element identity.
//
// Deprecated: use TryPatchChangeFor, which reports setPath failure.
func (s *Store) PatchChangeFor(identity, path string, value any) Change {
	c, _ := s.patchChangeWithIdentity(identity, path, value)
	return c
}

// PatchChangeWithIdentity is an alias.
//
// Deprecated: use TryPatchChangeWithIdentity.
func (s *Store) PatchChangeWithIdentity(identity, path string, value any) Change {
	c, _ := s.patchChangeWithIdentity(identity, path, value)
	return c
}

// PatchChangeWithStableID is an alias.
//
// Deprecated: use TryPatchChangeWithStableID.
func (s *Store) PatchChangeWithStableID(identity, path string, value any) Change {
	c, _ := s.patchChangeWithIdentity(identity, path, value)
	return c
}

// PatchChangeElement is an alias.
//
// Deprecated: use TryPatchChangeElement.
func (s *Store) PatchChangeElement(identity, path string, value any) Change {
	c, _ := s.patchChangeWithIdentity(identity, path, value)
	return c
}

func (s *Store) patchChangeWithIdentity(identity, path string, value any) (Change, bool) {
	if s == nil {
		return Change{}, false
	}
	s.mu.Lock()
	old := valueAtPath(s.data, path)
	c := Change{
		Path:       path,
		Old:        old,
		New:        value,
		Origin:     s.origin,
		At:         time.Now(),
		Identity:   identity,
		ElementID:  identity,
		StableID:   identity,
		Properties: []string{path},
		Changed:    []string{path},
		Props:      []string{path},
	}
	nd, ok := setPath(s.data, path, value)
	if !ok {
		s.mu.Unlock()
		return c, false
	}
	s.data = nd
	s.changes = append(s.changes, c)
	s.undo = append(s.undo, c)
	s.redo = nil
	ls := s.listenersLocked()
	s.mu.Unlock()
	for _, l := range ls {
		l(c)
	}
	return c, true
}
