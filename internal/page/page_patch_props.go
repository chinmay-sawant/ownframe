package page

// ChangedProperties returns a snapshot, falling back to the bound store when empty.
func (p *Page) ChangedProperties() []string {
	if p == nil {
		return nil
	}
	return p.stateBinding.ChangedProperties()
}

// TakeChangedProperties drains local metadata; the store fallback remains a snapshot.
func (p *Page) TakeChangedProperties() []string {
	if p == nil {
		return nil
	}
	return p.stateBinding.TakeChangedProperties()
}

// SetChangedProperties copies handler metadata into the binding.
func (p *Page) SetChangedProperties(props []string) {
	if p == nil {
		return
	}
	p.stateBinding.SetChangedProperties(props)
}

// AddChangedProperty appends a nonempty property path.
func (p *Page) AddChangedProperty(prop string) {
	if p == nil {
		return
	}
	p.stateBinding.AddChangedProperties(prop)
}

// AddChangedProperties appends nonempty property paths.
func (p *Page) AddChangedProperties(props []string) {
	if p == nil {
		return
	}
	p.stateBinding.AddChangedProperties(props...)
}

// ClearChangedProperties clears local metadata without clearing the store.
func (p *Page) ClearChangedProperties() {
	if p == nil {
		return
	}
	p.stateBinding.ClearChangedProperties()
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
