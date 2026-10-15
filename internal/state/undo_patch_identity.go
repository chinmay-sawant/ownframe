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
	if nd, ok := setPath(s.data, path, value); ok {
		s.data = nd
	}
	s.changes = append(s.changes, c)
	s.undo = append(s.undo, c)
	s.redo = nil
	ls := make([]func(Change), len(s.listeners))
	copy(ls, s.listeners)
	s.mu.Unlock()
	for _, l := range ls {
		l(c)
	}
	return c
}
