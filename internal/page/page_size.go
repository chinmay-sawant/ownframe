package page

// Handle registers the input functions. A later call replaces the whole
// set; use AddHandlers to merge one callback while keeping the rest.
func (p *Page) Handle(h Handlers) {
	p.handlers = h
}

// SetData stores the value the template prints on the next Redraw.
func (p *Page) SetData(data any) {
	p.dataMu.Lock()
	p.data = data
	p.dataMu.Unlock()
}

// Title returns the window title.
func (p *Page) Title() string {
	return p.title
}

// Size returns the frame size in CSS pixels.
func (p *Page) Size() (int, int) {
	return p.width, p.height
}

// MinSize returns the configured desktop minimum. Responsive mobile pages
// follow the native viewport instead.
func (p *Page) MinSize() (int, int) {
	return p.minWidth, p.minHeight
}

// MaxSize returns the largest window the page asks its host for. The
// window clamps to it; the layout does not.
func (p *Page) MaxSize() (int, int) {
	return p.maxWidth, p.maxHeight
}

// Clamp applies the desktop minimum or the native mobile viewport. LockView
// retains its canvas minimum on mobile. The maximum only bounds the window.
func (p *Page) Clamp(width, height int) (int, int) {
	if p.mobileViewport && !p.lockView {
		return clampMin(width, 1), clampMin(height, 1)
	}
	return clampMin(width, p.minWidth), clampMin(height, p.minHeight)
}

// SetSize stores the frame size used by the next Redraw.
// Clamp selects the minimum for the current host and canvas policy.
func (p *Page) SetSize(width, height int) {
	p.width, p.height = p.Clamp(width, height)
}

// Generation increases by one on every successful Redraw.
func (p *Page) Generation() uint64 {
	return p.generation
}

// LayerGeneration keys the compositor's retained element layers. It changes
// when a publish or an element patch changes display content, and stays put
// through a re-projection pass, because that pass rewrites operation matrices
// in place. A layer texture reuses its bake across a 3D animation; the layer
// signature still catches a colour or text change.
func (p *Page) LayerGeneration() uint64 {
	return p.layerGen
}

func clampMin(v, min int) int {
	if v < min {
		return min
	}

	return v
}
