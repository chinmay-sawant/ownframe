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
