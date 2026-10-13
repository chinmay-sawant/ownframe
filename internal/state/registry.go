package state

// StableRegistry keeps stable IDs for repeated items across re-renders.
type StableRegistry struct {
	ids map[string]int
}

// NewStableRegistry returns an empty registry.
func NewStableRegistry() *StableRegistry {
	return &StableRegistry{ids: map[string]int{}}
}

// ID returns a stable id for key, allocating if new.
func (r *StableRegistry) ID(key string) string {
	if r == nil {
		return key
	}
	if key == "" {
		return ""
	}
	if _, ok := r.ids[key]; !ok {
		r.ids[key] = len(r.ids)
	}
	return key
}
