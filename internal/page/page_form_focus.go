package page

// formFocusState owns the form document, the pointer hit state, and the
// accessibility snapshot. Only this file and page_form_focus_hit.go write
// its fields; Page delegates below. No lock: the page loop serializes it.
type formFocusState struct {
	form           *formState
	hover          string
	hoverX, hoverY float64
	active         string
	pressUsed      bool
	accessibility  *accessibilityState
}

// formState returns the owned form document, or nil before first sync.
func (p *Page) formState() *formState { return p.ff.form }

// setFormState replaces the owned form document.
func (p *Page) setFormState(s *formState) { p.ff.form = s }

// hoverID is the hovered element id, or "" when none.
func (p *Page) hoverID() string { return p.ff.hover }

// setHoverID records the hovered element id.
func (p *Page) setHoverID(id string) { p.ff.hover = id }

// clearHoverID drops the hovered element id.
func (p *Page) clearHoverID() { p.ff.hover = "" }

// hoverPoint is the last hover position in CSS pixels.
func (p *Page) hoverPoint() (float64, float64) { return p.ff.hoverX, p.ff.hoverY }

// setHoverPoint records the last hover position in CSS pixels.
func (p *Page) setHoverPoint(x, y float64) { p.ff.hoverX, p.ff.hoverY = x, y }

// activeID is the pressed element id, or "" when none.
func (p *Page) activeID() string { return p.ff.active }

// setActiveID records the pressed element id.
func (p *Page) setActiveID(id string) { p.ff.active = id }
