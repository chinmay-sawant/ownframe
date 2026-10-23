package state

// listener identity belongs to the registration, not a closure's code address.
type listener struct{ fn func(Change) }

// Deprecated: use AddListener.
func (s *Store) OnChange(fn func(Change)) {
	s.AddListener(fn)
}

// AddListener registers fn and returns an idempotent cancellation function.
// Cancellation removes future notifications. A callback already copied by a
// dispatch may finish; listeners run outside the store lock and may reenter it.
func (s *Store) AddListener(fn func(Change)) (cancel func()) {
	if s == nil || fn == nil {
		return func() {}
	}
	entry := &listener{fn: fn}
	s.mu.Lock()
	s.listeners = append(s.listeners, entry)
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, current := range s.listeners {
			if current != entry {
				continue
			}
			copy(s.listeners[i:], s.listeners[i+1:])
			s.listeners[len(s.listeners)-1] = nil
			s.listeners = s.listeners[:len(s.listeners)-1]
			entry.fn = nil
			return
		}
	}
}

// Deprecated: use AddListener and retain its cancellation function.
func (s *Store) Subscribe(fn func(Change)) { s.AddListener(fn) }

// Listeners returns a copy of registered listeners.
func (s *Store) Listeners() []func(Change) {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listenersLocked()
}

// OnFieldChange registers a listener that only fires for path and its children.
func (s *Store) OnFieldChange(path string, fn func(Change)) {
	s.AddListener(func(c Change) {
		if c.Path == path || hasPrefixPath(c.Path, path) {
			fn(c)
		}
	})
}

// Deprecated: use OnFieldChange.
func (s *Store) BindField(path string, fn func(Change)) { s.OnFieldChange(path, fn) }

// listenersLocked copies callbacks while the caller holds s.mu.
func (s *Store) listenersLocked() []func(Change) {
	out := make([]func(Change), len(s.listeners))
	for i, entry := range s.listeners {
		out[i] = entry.fn
	}
	return out
}

func hasPrefixPath(path, prefix string) bool {
	if prefix == "" {
		return true
	}
	return len(path) > len(prefix) && path[:len(prefix)] == prefix && path[len(prefix)] == '.'
}
