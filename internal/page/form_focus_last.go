package page

// lastFocusID reports the last control cleared by a blur.
func (p *Page) lastFocusIDValue() string {
	if p.form == nil {
		return ""
	}
	return p.form.lastFocusID
}

// LastFocusID returns the control id cleared by the most recent blur.
func (p *Page) LastFocusID() string { return p.lastFocusIDValue() }

// PreviousFocusID is an alias for LastFocusID.
func (p *Page) PreviousFocusID() string { return p.lastFocusIDValue() }

// BlurredFocusID is an alias for LastFocusID.
func (p *Page) BlurredFocusID() string { return p.lastFocusIDValue() }

// LastFocused is an alias for LastFocusID.
func (p *Page) LastFocused() string { return p.lastFocusIDValue() }
