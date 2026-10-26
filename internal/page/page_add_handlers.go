package page

// AddHandlers merges only non-nil callbacks into the existing set; a
// nil callback keeps the prior handler. Handle replaces the whole set.
func (p *Page) AddHandlers(h Handlers) {
	if h.Hover != nil {
		p.handlers.Hover = h.Hover
	}
	if h.Click != nil {
		p.handlers.Click = h.Click
	}
	if h.ClickPaint != nil {
		p.handlers.ClickPaint = h.ClickPaint
	}
	if h.Press != nil {
		p.handlers.Press = h.Press
	}
	if h.LongPress != nil {
		p.handlers.LongPress = h.LongPress
	}
	if h.KeyDown != nil {
		p.handlers.KeyDown = h.KeyDown
	}
	if h.KeyUp != nil {
		p.handlers.KeyUp = h.KeyUp
	}
	if h.Swipe != nil {
		p.handlers.Swipe = h.Swipe
	}
	if h.DragStart != nil {
		p.handlers.DragStart = h.DragStart
	}
	if h.DragMove != nil {
		p.handlers.DragMove = h.DragMove
	}
	if h.DragEnd != nil {
		p.handlers.DragEnd = h.DragEnd
	}
	if h.Wheel != nil {
		p.handlers.Wheel = h.Wheel
	}
	if h.BeforeEdit != nil {
		p.handlers.BeforeEdit = h.BeforeEdit
	}
	if h.Change != nil {
		p.handlers.Change = h.Change
	}
	if h.Type != nil {
		p.handlers.Type = h.Type
	}
	if h.Backspace != nil {
		p.handlers.Backspace = h.Backspace
	}
	if h.DeleteWord != nil {
		p.handlers.DeleteWord = h.DeleteWord
	}
	if h.Submit != nil {
		p.handlers.Submit = h.Submit
	}
	if h.Copy != nil {
		p.handlers.Copy = h.Copy
	}
	if h.Cut != nil {
		p.handlers.Cut = h.Cut
	}
	if h.Paste != nil {
		p.handlers.Paste = h.Paste
	}
	if h.SelectAll != nil {
		p.handlers.SelectAll = h.SelectAll
	}
	if h.Undo != nil {
		p.handlers.Undo = h.Undo
	}
	if h.Redo != nil {
		p.handlers.Redo = h.Redo
	}
	if h.Drop != nil {
		p.handlers.Drop = h.Drop
	}
	if h.OpenLink != nil {
		p.handlers.OpenLink = h.OpenLink
	}
	if h.DeepLink != nil {
		p.handlers.DeepLink = h.DeepLink
	}
}
