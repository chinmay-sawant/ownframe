package page

import "github.com/chinmay-sawant/ownframe/internal/state"

// State returns the bound state store.
func (p *Page) State() *state.Store { return p.stateBinding.Store() }

// BindState replaces the subscription. Bind nil to detach.
func (p *Page) BindState(s *state.Store) {
	if p != nil {
		p.stateBinding.Bind(s)
	}
}

// SyncState refreshes surfaced metadata from the current binding only.
func (p *Page) SyncState() {
	if p != nil {
		p.stateBinding.Sync()
	}
}
