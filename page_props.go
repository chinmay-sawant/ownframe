package ownframe

// ChangedProperties returns changed-property metadata surfaced to the page.
func (p *Page) ChangedProperties() []string {
	return innerPage(p).ChangedProperties()
}

// ClearChangedProperties clears surfaced metadata.
func (p *Page) ClearChangedProperties() {
	innerPage(p).ClearChangedProperties()
}

// AddChangedProperty appends one property path.
func (p *Page) AddChangedProperty(prop string) {
	innerPage(p).AddChangedProperty(prop)
}

// AddChangedProperties appends multiple property paths.
func (p *Page) AddChangedProperties(props []string) {
	innerPage(p).AddChangedProperties(props)
}

// SetChangedProperties records property metadata surfaced from a handler.
func (p *Page) SetChangedProperties(props []string) {
	innerPage(p).SetChangedProperties(props)
}

// TakeChangedProperties returns a snapshot and drains local metadata.
func (p *Page) TakeChangedProperties() []string {
	return innerPage(p).TakeChangedProperties()
}
