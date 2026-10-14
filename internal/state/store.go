package state

import "sync"

// Store keeps one data value and reports field-level changes.
type Store struct {
	mu sync.RWMutex

	data any

	focus  FocusState
	scroll ScrollState

	changes []Change

	listeners []func(Change)

	origin Origin

	undo []Change
	redo []Change

	stable *StableRegistry

	interceptors []RouteInterceptor
}

// NewStore wraps data. A nil data keeps the store usable.
func NewStore(data any) *Store {
	return &Store{data: data, stable: NewStableRegistry()}
}

// Data returns the wrapped value.
func (s *Store) Data() any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// SetData replaces the wrapped value and clears history.
func (s *Store) SetData(data any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
	s.changes = nil
	s.undo = nil
	s.redo = nil
}
