package page

import "sync"

// changedPropsMu guards Page.changedProps. State store listeners fire on
// arbitrary goroutines while the page loop drains, so every access here
// and in page_patch_set.go holds this lock.
var changedPropsMu sync.RWMutex

// ChangedProperties returns changed-property metadata surfaced to the page.
// When no local props are set it falls back to the bound store snapshot.
func (p *Page) ChangedProperties() []string {
	if p == nil {
		return nil
	}
	changedPropsMu.RLock()
	empty := len(p.changedProps) == 0
	local := make([]string, len(p.changedProps))
	copy(local, p.changedProps)
	changedPropsMu.RUnlock()
	if empty && p.stateStore != nil {
		return p.stateStore.ChangedProperties()
	}
	return local
}

// ChangedProps is an alias for ChangedProperties.
func (p *Page) ChangedProps() []string { return p.ChangedProperties() }

// LastChangedProperties returns the most recently surfaced properties.
func (p *Page) LastChangedProperties() []string { return p.ChangedProperties() }

// LastChangedProps is an alias.
func (p *Page) LastChangedProps() []string { return p.LastChangedProperties() }

// TakeChangedProperties returns a snapshot and drains local metadata.
// A store fallback is snapshot-only; clear it via State().ClearChangedProperties.
func (p *Page) TakeChangedProperties() []string {
	if p == nil {
		return nil
	}
	changedPropsMu.Lock()
	local := make([]string, len(p.changedProps))
	copy(local, p.changedProps)
	empty := len(p.changedProps) == 0
	p.changedProps = p.changedProps[:0]
	changedPropsMu.Unlock()
	if empty && p.stateStore != nil {
		return p.stateStore.ChangedProperties()
	}
	return local
}

// TakeChangedProps is an alias.
func (p *Page) TakeChangedProps() []string { return p.TakeChangedProperties() }

// PropertiesChanged is an alias.
func (p *Page) PropertiesChanged() []string { return p.ChangedProperties() }

// Changed returns surfaced metadata.
func (p *Page) Changed() []string { return p.ChangedProperties() }
