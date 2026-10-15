package state

// TryPatchChangeFor applies value, reporting setPath failure.
func (s *Store) TryPatchChangeFor(id, path string, v any) (Change, bool) {
	return s.patchChangeWithIdentity(id, path, v)
}

// TryPatchChangeWithIdentity is an alias for TryPatchChangeFor.
func (s *Store) TryPatchChangeWithIdentity(id, p string, v any) (Change, bool) {
	return s.patchChangeWithIdentity(id, p, v)
}

// TryPatchChangeWithStableID is an alias for TryPatchChangeFor.
func (s *Store) TryPatchChangeWithStableID(id, p string, v any) (Change, bool) {
	return s.patchChangeWithIdentity(id, p, v)
}

// TryPatchChangeElement is an alias for TryPatchChangeFor.
func (s *Store) TryPatchChangeElement(id, p string, v any) (Change, bool) {
	return s.patchChangeWithIdentity(id, p, v)
}
