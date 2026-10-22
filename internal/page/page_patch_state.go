package page

import "github.com/chinmay-sawant/ownframe/internal/state"

// State returns the bound state store.
func (p *Page) State() *state.Store {
	p.changedPropsMu.RLock()
	defer p.changedPropsMu.RUnlock()
	return p.stateStore
}

// BindState replaces the subscription and clears metadata from the former
// binding. Binding the same store again is a no-op. Bind nil to detach.
func (p *Page) BindState(s *state.Store) {
	if p == nil {
		return
	}
	p.changedPropsMu.Lock()
	defer p.changedPropsMu.Unlock()
	if p.stateStore == s {
		return
	}
	if p.stateCancel != nil {
		p.stateCancel()
		p.stateCancel = nil
	}
	p.stateEpoch++
	p.stateStore = s
	p.changedProps = nil
	if s == nil {
		return
	}
	epoch := p.stateEpoch
	p.stateCancel = s.AddListener(func(c state.Change) {
		p.recordStateChange(s, epoch, c)
	})
}

// recordStateChange rejects callbacks already dispatched by a former binding.
func (p *Page) recordStateChange(store *state.Store, epoch uint64, c state.Change) {
	p.changedPropsMu.Lock()
	defer p.changedPropsMu.Unlock()
	if p.stateStore != store || p.stateEpoch != epoch {
		return
	}
	if c.Path != "" {
		p.changedProps = append(p.changedProps, c.Path)
	}
	for _, prop := range c.Properties {
		if prop != "" {
			p.changedProps = append(p.changedProps, prop)
		}
	}
}

// SyncState refreshes surfaced metadata from the current binding only.
func (p *Page) SyncState() {
	if p == nil {
		return
	}
	p.changedPropsMu.RLock()
	store, epoch := p.stateStore, p.stateEpoch
	p.changedPropsMu.RUnlock()
	if store != nil {
		p.recordStateChange(store, epoch, state.Change{Properties: store.ChangedProperties()})
	}
}
