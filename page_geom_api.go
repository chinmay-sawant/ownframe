package ownframe

// Title returns the window title.
func (page *Page) Title() string {
	p := innerPage(page)
	if p == nil {
		return ""
	}

	return p.Title()
}

// Boxes returns a copy of the last hit-test boxes.
func (page *Page) Boxes() []Box {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Boxes()
}
