package page

import "github.com/chinmay-sawant/ownframe/internal/state"

// State returns the bound state store.
func (p *Page) State() *state.Store { return p.stateStore }

// StateStore is an alias for State.
func (p *Page) StateStore() *state.Store { return p.stateStore }

// Store returns the bound state store.
func (p *Page) Store() *state.Store { return p.stateStore }

// SetState binds a state store and surfaces its change metadata.
func (p *Page) SetState(s *state.Store) {
	p.stateStore = s
	p.BindState(s)
}

// SetStateStore is an alias.
func (p *Page) SetStateStore(s *state.Store) { p.SetState(s) }

// SetStore is an alias.
func (p *Page) SetStore(s *state.Store) { p.SetState(s) }

// AttachState is an alias.
func (p *Page) AttachState(s *state.Store) { p.SetState(s) }

// BindState registers a listener that surfaces changed properties.
func (p *Page) BindState(s *state.Store) {
	if p == nil {
		return
	}
	p.stateStore = s
	if s == nil {
		return
	}
	s.OnChange(func(c state.Change) {
		if c.Path != "" {
			p.AddChangedProperty(c.Path)
		}
		if len(c.Properties) > 0 {
			p.AddChangedProperties(c.Properties)
		}
	})
}

// BindStore is an alias.
func (p *Page) BindStore(s *state.Store) { p.BindState(s) }

// AttachStore is an alias.
func (p *Page) AttachStore(s *state.Store) { p.BindState(s) }

// SyncState refreshes surfaced metadata from the bound store.
func (p *Page) SyncState() {
	if p == nil || p.stateStore == nil {
		return
	}
	p.AddChangedProperties(p.stateStore.ChangedProperties())
}

// SyncChangedProperties is an alias.
func (p *Page) SyncChangedProperties() { p.SyncState() }

// ChangedPropertiesFromStore returns store metadata.
func (p *Page) ChangedPropertiesFromStore() []string {
	if p == nil || p.stateStore == nil {
		return nil
	}
	return p.stateStore.ChangedProperties()
}
