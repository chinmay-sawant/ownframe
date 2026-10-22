package page

// ChangedProperties returns changed-property metadata surfaced to the page.
// When no local props are set it falls back to the bound store snapshot.
func (p *Page) ChangedProperties() []string {
	if p == nil {
		return nil
	}
	p.changedPropsMu.RLock()
	empty := len(p.changedProps) == 0
	local := make([]string, len(p.changedProps))
	copy(local, p.changedProps)
	store := p.stateStore
	p.changedPropsMu.RUnlock()
	if empty && store != nil {
		return store.ChangedProperties()
	}
	return local
}

// TakeChangedProperties returns a snapshot and drains local metadata.
// A store fallback is snapshot-only; clear it via State().ClearChangedProperties.
func (p *Page) TakeChangedProperties() []string {
	if p == nil {
		return nil
	}
	p.changedPropsMu.Lock()
	local := make([]string, len(p.changedProps))
	copy(local, p.changedProps)
	empty := len(p.changedProps) == 0
	p.changedProps = p.changedProps[:0]
	store := p.stateStore
	p.changedPropsMu.Unlock()
	if empty && store != nil {
		return store.ChangedProperties()
	}
	return local
}

// SetChangedProperties records property metadata surfaced from a handler.
func (p *Page) SetChangedProperties(props []string) {
	if p == nil {
		return
	}
	p.changedPropsMu.Lock()
	p.changedProps = append([]string(nil), props...)
	p.changedPropsMu.Unlock()
}

// Deprecated: use SetChangedProperties.
func (p *Page) SetChangedProps(props []string) { p.SetChangedProperties(props) }

// AddChangedProperty appends one property path.
func (p *Page) AddChangedProperty(prop string) {
	if p == nil || prop == "" {
		return
	}
	p.changedPropsMu.Lock()
	p.changedProps = append(p.changedProps, prop)
	p.changedPropsMu.Unlock()
}

// AddChangedProperties appends multiple property paths.
func (p *Page) AddChangedProperties(props []string) {
	if p == nil {
		return
	}
	p.changedPropsMu.Lock()
	for _, prop := range props {
		if prop != "" {
			p.changedProps = append(p.changedProps, prop)
		}
	}
	p.changedPropsMu.Unlock()
}

// Deprecated: use AddChangedProperty.
func (p *Page) RecordChangedProperty(prop string) { p.AddChangedProperty(prop) }

// Deprecated: use AddChangedProperties.
func (p *Page) RecordChangedProperties(props []string) { p.AddChangedProperties(props) }

// ClearChangedProperties clears surfaced metadata.
func (p *Page) ClearChangedProperties() {
	if p == nil {
		return
	}
	p.changedPropsMu.Lock()
	p.changedProps = p.changedProps[:0]
	p.changedPropsMu.Unlock()
}

// withChangedProps fills the canonical field once; deprecated
// fields share the same backing array. Treat them as read-only.
func (p *Page) withChangedProps(patch DisplayPatch) DisplayPatch {
	props := p.ChangedProperties()
	clone := append([]string(nil), props...)
	patch.ChangedProperties = clone
	patch.Properties = clone
	patch.Props = clone
	patch.ChangedProps = clone
	return patch
}
