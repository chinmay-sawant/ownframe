package ownframe

// Boxes returns a copy of the last hit-test boxes.
func (p *Page) Boxes() []Box { return innerPage(p).Boxes() }
