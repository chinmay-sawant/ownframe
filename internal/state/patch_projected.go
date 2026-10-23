package state

import (
	"sort"
	"time"
)

// PatchFields applies one projected value per path and returns the changes.
func (s *Store) PatchFields(values map[string]any) []Change {
	return s.PatchProjected(values)
}

// PatchProjected applies values in sorted key order under one lock.
func (s *Store) PatchProjected(values map[string]any) []Change {
	if s == nil || len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s.mu.Lock()
	out := make([]Change, 0, len(keys))
	for _, path := range keys {
		v := values[path]
		c := Change{Path: path, Old: valueAtPath(s.data, path),
			New: v, Origin: s.origin, At: time.Now(),
			Properties: []string{path}, Changed: []string{path},
			Props: []string{path}}
		nd, ok := setPath(s.data, path, v)
		if !ok {
			continue
		}
		s.data = nd
		s.changes = append(s.changes, c)
		s.undo = append(s.undo, c)
		out = append(out, c)
	}
	if len(out) > 0 {
		s.redo = nil
	}
	ls := s.listenersLocked()
	s.mu.Unlock()
	for _, c := range out {
		for _, l := range ls {
			l(c)
		}
	}
	return out
}

// PatchSlice is an alias for PatchChange.
func (s *Store) PatchSlice(path string, value any) Change {
	return s.PatchChange(path, value)
}

// Patch applies a Patch's changed paths in sorted order.
func (s *Store) Patch(p Patch) []Change {
	out, _ := s.PatchWithSkipped(p)
	return out
}

// ApplyPatch is an alias for Patch.
func (s *Store) ApplyPatch(p Patch) []Change { return s.Patch(p) }

// ApplyProjected is an alias for PatchProjected.
func (s *Store) ApplyProjected(values map[string]any) []Change {
	return s.PatchProjected(values)
}

// Projected returns the current value at path.
func (s *Store) Projected(path string) any {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return valueAtPath(s.data, path)
}
