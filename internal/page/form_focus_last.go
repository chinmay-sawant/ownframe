package page

// lastFocusID reports the last control cleared by a blur.
func (p *Page) lastFocusIDValue() string {
	if p.formState() == nil {
		return ""
	}
	return p.formState().lastFocusID
}

// LastFocusID returns the control id cleared by the most recent blur.
func (p *Page) LastFocusID() string { return p.lastFocusIDValue() }
