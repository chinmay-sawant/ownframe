package page

// SetChangedProperties records property metadata surfaced from a handler.
func (p *Page) SetChangedProperties(props []string) {
	if p == nil {
		return
	}
	p.changedProps = append([]string(nil), props...)
}

// SetChangedProps is an alias.
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

// RecordChangedProperty appends one property path.
func (p *Page) RecordChangedProperty(prop string) { p.AddChangedProperty(prop) }

// RecordChangedProperties appends multiple property paths.
func (p *Page) RecordChangedProperties(props []string) { p.AddChangedProperties(props) }

// ClearChangedProperties clears surfaced metadata.
func (p *Page) ClearChangedProperties() {
	if p == nil {
		return
	}
	p.changedProps = p.changedProps[:0]
}

func (p *Page) withChangedProps(patch DisplayPatch) DisplayPatch {
	props := p.ChangedProperties()
	patch.Properties = append([]string(nil), props...)
	patch.Props = append([]string(nil), props...)
	patch.ChangedProperties = append([]string(nil), props...)
	patch.ChangedProps = append([]string(nil), props...)
	return patch
}
