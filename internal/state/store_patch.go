package state

import "time"

// PatchChange records a field-level Change at path and applies value.
func (s *Store) PatchChange(path string, value any) Change {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := valueAtPath(s.data, path)
	c := Change{
		Path:   path,
		Old:    old,
		New:    value,
		Origin: s.origin,
		At:     time.Now(),
	}

	if nd, ok := setPath(s.data, path, value); ok {
		s.data = nd
	}

	s.changes = append(s.changes, c)
	s.undo = append(s.undo, c)
	s.redo = nil

	for _, l := range s.listeners {
		l(c)
	}
	return c
}
