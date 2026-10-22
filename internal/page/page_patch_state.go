package page

import "github.com/chinmay-sawant/ownframe/internal/state"

// State returns the bound state store.
func (p *Page) State() *state.Store { return p.stateStore }

// SetState binds a state store and surfaces its change metadata.
func (p *Page) SetState(s *state.Store) {
	p.stateStore = s
	p.BindState(s)
}

// BindState registers a listener that surfaces changed properties.
func (p *Page) BindState(s *state.Store) {
	if p == nil {
		return
	}
	p.stateStore = s
	if s == nil {
		return
	}
	s.AddListener(func(c state.Change) {
		if c.Path != "" {
			p.AddChangedProperty(c.Path)
		}
		if len(c.Properties) > 0 {
			p.AddChangedProperties(c.Properties)
		}
	})
}

// SyncState refreshes surfaced metadata from the bound store.
func (p *Page) SyncState() {
	if p == nil || p.stateStore == nil {
		return
	}
	p.AddChangedProperties(p.stateStore.ChangedProperties())
}
