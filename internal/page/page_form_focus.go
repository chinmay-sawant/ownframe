package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// formFocusState owns the form document, the pointer hit state, and the
// accessibility snapshot. Only this file and page_form_focus_hit.go write
// its fields; Page delegates below. No lock: the page loop serializes it.
type formFocusState struct {
	form           *form.State
	hover          string
	hoverX, hoverY float64
	active         string
	pressUsed      bool
	accessibility  *accessibilityState
}

// formState returns the owned form document, or nil before first sync.
func (p *Page) formState() *form.State { return p.ff.form }

// setFormState replaces the owned form document.
func (p *Page) setFormState(s *form.State) { p.ff.form = s }

// HoverID returns the element the page paints as hovered, or "" when none.
// The window reads it through host.Hoverer so a hover the page set itself
// survives a frame that did not move the pointer.
func (p *Page) HoverID() string { return p.ff.hover }

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
