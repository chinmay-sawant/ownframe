package state

import "time"

// TryPatchChange applies value at path, recording only on success.
func (s *Store) TryPatchChange(path string, value any) (Change, bool) {
	if s == nil {
		return Change{}, false
	}
	s.mu.Lock()
	old := valueAtPath(s.data, path)
	c := Change{
		Path: path, Old: old, New: value, Origin: s.origin,
		At: time.Now(), Properties: []string{path},
		Changed: []string{path}, Props: []string{path},
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
	ls := make([]func(Change), len(s.listeners))
	copy(ls, s.listeners)
	s.mu.Unlock()
	for _, l := range ls {
		l(c)
	}
	return c, true
}

// Deprecated: use TryPatchChange, which reports setPath failure.
func (s *Store) PatchChange(path string, value any) Change {
	c, _ := s.TryPatchChange(path, value)
	return c
}
