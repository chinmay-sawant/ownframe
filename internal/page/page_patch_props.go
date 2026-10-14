package page

// ChangedProperties returns changed-property metadata surfaced to the page.
func (p *Page) ChangedProperties() []string {
	if p == nil {
		return nil
	}
	if len(p.changedProps) == 0 && p.stateStore != nil {
		return p.stateStore.ChangedProperties()
	}
	out := make([]string, len(p.changedProps))
	copy(out, p.changedProps)
	return out
}

// ChangedProps is an alias for ChangedProperties.
func (p *Page) ChangedProps() []string { return p.ChangedProperties() }

// LastChangedProperties returns the most recently surfaced properties.
func (p *Page) LastChangedProperties() []string { return p.ChangedProperties() }

// LastChangedProps is an alias.
func (p *Page) LastChangedProps() []string { return p.LastChangedProperties() }

// TakeChangedProperties returns a snapshot of surfaced metadata.
func (p *Page) TakeChangedProperties() []string { return p.ChangedProperties() }

// TakeChangedProps is an alias.
func (p *Page) TakeChangedProps() []string { return p.TakeChangedProperties() }

// PropertiesChanged is an alias.
func (p *Page) PropertiesChanged() []string { return p.ChangedProperties() }

// Changed returns surfaced metadata.
func (p *Page) Changed() []string { return p.ChangedProperties() }
