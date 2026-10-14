package state

// PatchFields applies one projected value per path and returns the changes.
func (s *Store) PatchFields(values map[string]any) []Change {
	return s.PatchProjected(values)
}

// PatchProjected applies a projected state slice (path -> value).
func (s *Store) PatchProjected(values map[string]any) []Change {
	if s == nil {
		return nil
	}
	out := make([]Change, 0, len(values))
	for path, value := range values {
		out = append(out, s.PatchChange(path, value))
	}
	return out
}

// PatchSlice is an alias for PatchChange.
func (s *Store) PatchSlice(path string, value any) Change {
	return s.PatchChange(path, value)
}

// Patch applies a Patch's changed paths.
func (s *Store) Patch(p Patch) []Change {
	if s == nil {
		return nil
	}
	if p.IsWhole() {
		if p.Values != nil {
			if v, ok := p.Values[""]; ok {
				s.SetData(v)
			}
		}
		return nil
	}
	s.mu.Lock()
	s.origin = p.Origin
	s.mu.Unlock()
	out := make([]Change, 0, len(p.Changed))
	for _, path := range p.Changed {
		value, ok := p.Values[path]
		if !ok {
			continue
		}
		out = append(out, s.PatchChange(path, value))
	}
	s.origin = p.Origin
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
