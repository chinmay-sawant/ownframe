package state

import "time"

// PatchChange records a field-level Change at path and applies value.
func (s *Store) PatchChange(path string, value any) Change {
	s.mu.Lock()
	old := valueAtPath(s.data, path)
	c := Change{
		Path:       path,
		Old:        old,
		New:        value,
		Origin:     s.origin,
		At:         time.Now(),
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
