package state

import "sync"

// Binding owns one store subscription and its changed-property metadata.
// The zero value is detached and ready to use. Methods are safe for concurrent
// use. A Binding must not be copied after first use; Bind(nil) detaches it.
type Binding struct {
	mu         sync.RWMutex
	props      []string
	store      *Store
	cancel     func()
	generation uint64
}

// Store returns the bound state store.
func (b *Binding) Store() *Store {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.store
}

// Bind replaces the subscription and clears metadata from the former
// binding. Binding the same store again is a no-op. Bind nil to detach.
func (b *Binding) Bind(s *Store) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.store == s {
		return
	}
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
	b.generation++
	b.store = s
	b.props = nil
	if s == nil {
		return
	}
	epoch := b.generation
	b.cancel = s.AddListener(func(c Change) {
		b.recordStateChange(s, epoch, c)
	})
}

// recordStateChange rejects callbacks already dispatched by a former binding.
func (b *Binding) recordStateChange(store *Store, epoch uint64, c Change) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.store != store || b.generation != epoch {
		return
	}
	if c.Path != "" {
		b.props = append(b.props, c.Path)
	}
	for _, prop := range c.Properties {
		if prop != "" {
			b.props = append(b.props, prop)
		}
	}
}

// Sync refreshes surfaced metadata from the current binding only.
func (b *Binding) Sync() {
	if b == nil {
		return
	}
	b.mu.RLock()
	store, epoch := b.store, b.generation
	b.mu.RUnlock()
	if store != nil {
		b.recordStateChange(store, epoch, Change{Properties: store.ChangedProperties()})
	}
}

// ChangedProperties returns changed-property metadata surfaced to the binding.
// When no local props are set it falls back to the bound store snapshot.
func (b *Binding) ChangedProperties() []string {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	empty := len(b.props) == 0
	local := make([]string, len(b.props))
	copy(local, b.props)
	store := b.store
	b.mu.RUnlock()
	if empty && store != nil {
		return store.ChangedProperties()
	}
	return local
}

// TakeChangedProperties returns a snapshot and drains local metadata.
// A store fallback is snapshot-only; clear it via Store().ClearChangedProperties.
func (b *Binding) TakeChangedProperties() []string {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	local := make([]string, len(b.props))
	copy(local, b.props)
	empty := len(b.props) == 0
	b.props = b.props[:0]
	store := b.store
	b.mu.Unlock()
	if empty && store != nil {
		return store.ChangedProperties()
	}
	return local
}

// SetChangedProperties records property metadata surfaced from a handler.
func (b *Binding) SetChangedProperties(props []string) {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.props = append([]string(nil), props...)
	b.mu.Unlock()
}

// AddChangedProperties appends multiple property paths.
func (b *Binding) AddChangedProperties(props ...string) {
	if b == nil {
		return
	}
	b.mu.Lock()
	for _, prop := range props {
		if prop != "" {
			b.props = append(b.props, prop)
		}
	}
	b.mu.Unlock()
}

// ClearChangedProperties clears surfaced metadata.
func (b *Binding) ClearChangedProperties() {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.props = b.props[:0]
	b.mu.Unlock()
}
