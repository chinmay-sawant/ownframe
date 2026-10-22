package ownframe

// ChangedProperties returns changed-property metadata surfaced to the page.
func (t Testing) ChangedProperties() []string {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.ChangedProperties()
}

// ClearChangedProperties clears surfaced metadata.
func (t Testing) ClearChangedProperties() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.ClearChangedProperties()
}

// AddChangedProperty appends one property path.
func (t Testing) AddChangedProperty(prop string) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.AddChangedProperty(prop)
}

// AddChangedProperties appends multiple property paths.
func (t Testing) AddChangedProperties(props []string) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.AddChangedProperties(props)
}

// SetChangedProperties records property metadata surfaced from a handler.
func (t Testing) SetChangedProperties(props []string) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetChangedProperties(props)
}

// TakeChangedProperties returns a snapshot and drains local metadata.
func (t Testing) TakeChangedProperties() []string {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.TakeChangedProperties()
}
