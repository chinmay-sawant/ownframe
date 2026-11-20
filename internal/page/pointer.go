package page

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/click"
	"github.com/chinmay-sawant/ownframe/internal/page/hit"
	"net/url"
)

// Click hit-tests the last picture and applies the click. One box scan answers
// the animation-trigger pass and the click sequence.
func (p *Page) Click(ctx context.Context, x, y float64) error {
	spot := boxAt(p, x, y)
	if box, ok := activationSpot(p, spot); ok && !triggerActivated(p, box) {
		if _, err := fireTriggers(p, ctx, box, true); err != nil {
			return err
		}
	}

	return click.Run(ctx, clickSurface{p}, click.Point{X: x, Y: y, Spot: spot.hitSpot()})
}

// hitSpot converts the page's point query to the hit package's spot.
func (s boxSpot) hitSpot() hit.Spot {
	return hit.Spot{Box: s.box, OK: s.ok, ID: s.id}
}

// triggerActivated reports whether the press that started this gesture already
// fired the entry actions for box, so the click does not repeat them.
func triggerActivated(p *Page, box Box) bool {
	return p.triggerPress.armed && sameBox(p.triggerPress.box, box)
}

// clickSurface adapts Page to click.Surface.
// Each method delegates to the page method the sequence already used.
type clickSurface struct{ p *Page }

func (s clickSurface) BoxByID(id string) Box {
	return boxByID(s.p, id)
}

func (s clickSurface) ConsumePress(id string) bool {
	return consumePress(s.p, id)
}

func (s clickSurface) FormControl(id string) bool {
	return formControl(s.p, id)
}

func (s clickSurface) Control(id string) (Control, bool) {
	return control(s.p, id)
}

func (s clickSurface) Lookup(action string) (string, bool) {
	return s.p.nav.Lookup(action)
}

func (s clickSurface) Load(ctx context.Context, html string) error {
	return s.p.Load(ctx, html)
}

func (s clickSurface) ClickControl(ctx context.Context, box Box) error {
	return clickControl(s.p, ctx, box)
}

func (s clickSurface) BlurForm() {
	blurForm(s.p)
}

func (s clickSurface) HasOpenLink() bool {
	return s.p.handlers.OpenLink != nil
}

func (s clickSurface) OpenLinkAt(ctx context.Context, x, y float64) (bool, error) {
	return openLinkAt(s.p, ctx, x, y)
}

func (s clickSurface) HasClick() bool {
	return s.p.handlers.Click != nil
}

func (s clickSurface) HandlerClick(ctx context.Context, box Box) error {
	// Mark before the handler runs: an incremental update that adopts its
	// patched source clears the flag, while a handler that only mutates data
	// leaves it set for the next redraw's form sync.
	markFormDirty(s.p)

	return s.p.handlers.Click(ctx, box)
}

func (s clickSurface) ClickPainted(ctx context.Context, box Box) (bool, error) {
	return clickPainted(s.p, ctx, box)
}

func (s clickSurface) Redraw(ctx context.Context) error {
	return s.p.Redraw(ctx)
}

func (s clickSurface) RedrawIfChanged(ctx context.Context) error {
	return s.p.RedrawIfChanged(ctx)
}

func (s clickSurface) LabelControl(box Box) string {
	id := click.LabelControl(s.p.source, s.p.boxes, box)
	if id == "" {
		return ""
	}

	c, ok := control(s.p, id)
	if !ok || c.Disabled || !inFocusScope(s.p, id) {
		return ""
	}

	return id
}

func (s clickSurface) FocusControl(ctx context.Context, id string) error {
	return s.p.Focus(ctx, id)
}

var _ host.Screen = (*Page)(nil)

var _ click.Surface = clickSurface{}

// clickPainted runs the paint-only click handler. It reports whether the
// handler painted the frame itself. A true result clears the pending and
// force flags of the redraw batch, so the batch waits for no draw that
// will not come, and the next real redraw still happens.
//
// A frame the page would draw differently is not painted. The handler
// painted the pixels, but the state the sheets match still needs a layout,
// so the click keeps its redraw.
func clickPainted(p *Page, ctx context.Context, box Box) (bool, error) {
	if p.handlers.ClickPaint == nil {
		return false, nil
	}

	painted, err := p.handlers.ClickPaint(ctx, box)
	if err != nil {
		return false, err
	}
	// The handler can mutate the template data in place through the
	// SetData pointer, so force the next Redraw to re-execute.
	markFormDirty(p)
	if painted {
		if !unchanged(p, renderState(p)) {
			return false, nil
		}
		paintedFrame(p)
	}

	return painted, nil
}

// paintedFrame records that the frame is on screen already. The pending and
// force flags of the redraw batch go, so the batch does not wait for a draw
// the handler already made.
func paintedFrame(p *Page) {
	p.redrawPending = false
	p.redrawForce = false
}

// Handlers are optional callbacks.
type Handlers struct {
	Hover HoverHandler
	Click func(ctx context.Context, box Box) error
	// ClickPaint runs after Click and a control's own change. It reports
	// that the handler painted the frame, so the page skips the layout.
	ClickPaint func(ctx context.Context, box Box) (bool, error)
	// Press can claim the pointer-down.
	Press func(ctx context.Context, box Box, x, y float64) (bool, error)
	// LongPress handles a hold.
	LongPress func(ctx context.Context, box Box) error
	// KeyDown and KeyUp take lowercase names.
	KeyDown func(ctx context.Context, key string) error
	KeyUp   func(ctx context.Context, key string) error
	Swipe   func(ctx context.Context, dx, dy float64) error
	// DragStart claims a gesture.
	DragStart func(ctx context.Context, box Box) (bool, error)
	// DragMove takes CSS deltas.
	DragMove func(ctx context.Context, dx, dy float64) error
	DragEnd  func(ctx context.Context) error
	// Wheel offers a wheel event before the window pans the page.
	Wheel func(ctx context.Context, x, y, dx, dy float64) (bool, error)
	// BeforeEdit may reject an edit.
	BeforeEdit func(ctx context.Context, box Box) error
	// Change runs after edits; errors skip redraw.
	Change     func(ctx context.Context, box Box) error
	Type       func(ctx context.Context, text string) error
	Backspace  func(ctx context.Context) error
	DeleteWord func(ctx context.Context) error
	Submit     func(ctx context.Context) error
	Copy       func(ctx context.Context) (text string, ok bool, err error)
	Cut        func(ctx context.Context) (text string, ok bool, err error)
	Paste      func(ctx context.Context, text string) error
	SelectAll  func(ctx context.Context) error
	Undo       func(ctx context.Context) error
	Redo       func(ctx context.Context) error
	// ContextMenuFor returns custom rows for one hit-tested element.
	ContextMenuFor func(id string) []MenuItem
	// MenuAction runs a custom row the page reported.
	MenuAction func(ctx context.Context, id string) error
	// PresentMenu paints the context menu in the page. A true result tells the
	// window to skip its chrome menu.
	PresentMenu func(x, y int, items []MenuItem) bool
	// Drop readers work in the callback.
	Drop func(ctx context.Context, files []Drop) error
	// OpenLink takes validated links.
	OpenLink func(ctx context.Context, target *url.URL) error
	// DeepLink takes host URLs.
	DeepLink func(ctx context.Context, target *url.URL) error
}

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
	if h.ContextMenuFor != nil {
		p.handlers.ContextMenuFor = h.ContextMenuFor
	}
	if h.MenuAction != nil {
		p.handlers.MenuAction = h.MenuAction
	}
	if h.PresentMenu != nil {
		p.handlers.PresentMenu = h.PresentMenu
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

// hitItem is the hit list entry stored on the page. It is hit.Item so the
// page can pass the slice into package hit without copying it.
type hitItem = hit.Item

// boxSpot is one point query against the box index: the innermost box under
// the point and the innermost id-bearing box under it.
type boxSpot struct {
	box Box
	ok  bool
	id  string
}

func spotFrom(s hit.Spot) boxSpot {
	return boxSpot{box: s.Box, ok: s.OK, id: s.ID}
}

// boxAt answers the point query against the page's boxes. Click, Press,
// and Hover all read it, so overlapping boxes have one winner.
func boxAt(p *Page, x, y float64) boxSpot {
	if len(p.hits) > 0 {
		return spotFrom(hit.ScanList(p.hits, x, y, p.scroll.Windowing.X, p.scroll.Windowing.Y))
	}

	return spotFrom(hit.Scan(p.boxes, x, y))
}

// hitBox returns the box under a content point.
func hitBox(p *Page, x, y float64) (Box, bool) {
	spot := boxAt(p, x, y)

	return spot.box, spot.ok
}

// isFixed reports whether the window pins the box to the viewport.
func isFixed(p *Page, box Box) bool {
	return hit.IsFixed(p.hits, box)
}

// fixedBoxes returns the boxes the window pins to the viewport, such as a
// fixed header. The first operation of a fixed range belongs to that box.
func fixedBoxes(d *Display) []Box {
	return hit.FixedBoxes(d)
}
