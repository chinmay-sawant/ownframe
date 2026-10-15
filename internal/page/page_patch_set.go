package page

// SetChangedProperties records property metadata surfaced from a handler.
func (p *Page) SetChangedProperties(props []string) {
	if p == nil {
		return
	}
	p.changedProps = append([]string(nil), props...)
}

// Deprecated: use SetChangedProperties.
func (p *Page) SetChangedProps(props []string) { p.SetChangedProperties(props) }

// AddChangedProperty appends one property path.
func (p *Page) AddChangedProperty(prop string) {
	if p == nil || prop == "" {
		return
	}
	p.changedProps = append(p.changedProps, prop)
}

// AddChangedProperties appends multiple property paths.
func (p *Page) AddChangedProperties(props []string) {
	if p == nil {
		return
	}
	for _, prop := range props {
		if prop != "" {
			p.changedProps = append(p.changedProps, prop)
		}
	}
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
	p.changedProps = p.changedProps[:0]
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
