package page

// setPressUsed records whether a handler consumed the current press.
func (p *Page) setPressUsed(used bool) { p.ff.pressUsed = used }

// takePress clears the press hit and returns the held id, or "".
func (p *Page) takePress() string {
	id := p.ff.active
	p.ff.active = ""
	p.ff.pressUsed = false
	return id
}

// consumePress drops a handler-consumed press hit. It is true when Click
// lands on the pressed id and should only redraw.
func (p *Page) consumePress(id string) bool {
	if p.ff.pressUsed && p.ff.active == id {
		p.ff.pressUsed = false
		return true
	}
	return false
}

// a11yState returns the owned accessibility snapshot, or nil.
func (p *Page) a11yState() *accessibilityState { return p.ff.accessibility }

// setA11yState replaces the owned accessibility snapshot.
func (p *Page) setA11yState(s *accessibilityState) { p.ff.accessibility = s }
