package scroll

// Windowing is the scroll offset the window last reported and the
// row-window callback. The window owns scrolling; the callback slices
// app rows before each draw so only visible rows become boxes.
type Windowing struct {
	X  int
	Y  int
	On bool
	Fn func(offsetY, viewH int) bool
}

// SetOn opts into redraws on scroll changes.
func (s *State) SetOn(on bool) { s.Windowing.On = on }

// On reports whether scroll changes redraw this page.
func (s *State) On() bool { return s.Windowing.On }

// Offset is the last offset the window reported, in CSS pixels.
func (s *State) Offset() (int, int) { return s.Windowing.X, s.Windowing.Y }

// SetOffset records the window offset.
func (s *State) SetOffset(x, y int) { s.Windowing.X, s.Windowing.Y = x, y }

// SetFn installs a row-window callback.
func (s *State) SetFn(fn func(offsetY, viewH int) bool) { s.Windowing.Fn = fn }

// Apply runs the row-window callback. A nil callback, or a page that
// opted out, is a no-op.
func (s *State) Apply(height int) {
	w := s.Windowing
	if !w.On || w.Fn == nil {
		return
	}
	w.Fn(w.Y, height)
}

// Step prepares the visible rows and reports whether they changed. A
// page without a callback keeps the ordinary scroll redraw path.
func (s *State) Step(height int) bool {
	if s.Windowing.Fn == nil {
		return true
	}
	return s.Windowing.Fn(s.Windowing.Y, height)
}
