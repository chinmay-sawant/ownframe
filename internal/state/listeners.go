package state

import "reflect"

// OnChange registers a listener invoked for every field-level change.
func (s *Store) OnChange(fn func(Change)) {
	s.AddListener(fn)
}

// AddListener registers a change listener.
func (s *Store) AddListener(fn func(Change)) {
	if s == nil || fn == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, fn)
}

// Subscribe is an alias for AddListener.
func (s *Store) Subscribe(fn func(Change)) { s.AddListener(fn) }

// RemoveListener removes a previously registered listener.
func (s *Store) RemoveListener(fn func(Change)) bool {
	if s == nil || fn == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, l := range s.listeners {
		if fnPtrEqual(l, fn) {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			return true
		}
	}
	return false
}

// Listeners returns a copy of registered listeners.
func (s *Store) Listeners() []func(Change) {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]func(Change), len(s.listeners))
	copy(out, s.listeners)
	return out
}

// OnFieldChange registers a listener that only fires for path and its children.
func (s *Store) OnFieldChange(path string, fn func(Change)) {
	s.OnChange(func(c Change) {
		if c.Path == path || hasPrefixPath(c.Path, path) {
			fn(c)
		}
	})
}

// BindField is an alias for OnFieldChange.
func (s *Store) BindField(path string, fn func(Change)) { s.OnFieldChange(path, fn) }

func fnPtrEqual(a, b func(Change)) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

func hasPrefixPath(path, prefix string) bool {
	if prefix == "" {
		return true
	}
	return len(path) > len(prefix) && path[:len(prefix)] == prefix && path[len(prefix)] == '.'
}
