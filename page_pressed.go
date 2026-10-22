package ownframe

// PressedID is the control held by the current pointer gesture.
func (p *Page) PressedID() string { return innerPage(p).PressedID() }
