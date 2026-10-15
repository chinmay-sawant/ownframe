package page

// PressedID is the control held by the current pointer gesture.
// It is empty after Release.
func (p *Page) PressedID() string { return p.activeID() }
