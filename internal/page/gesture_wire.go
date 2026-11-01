package page

import (
	"context"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/gesture"
)

// pageDrag is the claimed pointer capture stored on the page.
type pageDrag = gesture.Drag

// Drop is one file or directory dropped on the window.
type Drop = host.Drop

// HoverHandler may paint a hover change directly into the retained display.
// Returning true skips CSS relayout; the page invalidates both boxes. Return
// false for any change that needs layout or ordinary CSS state handling.
// The handler must not change box geometry. An error leaves hover unchanged.
type HoverHandler func(ctx context.Context, previous, next Box) (handled bool, err error)

// KeyDown calls the key-down handler. When a text field is focused, Escape
// clears the focus and the caret keys move the caret; both draw. A nil
// handler does nothing else.
func (p *Page) KeyDown(ctx context.Context, key string) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if p.handlers.KeyDown != nil {
		if err := p.handlers.KeyDown(ctx, key); err != nil {
			return err
		}
		p.markFormDirty()
	}

	var field Control
	if fs := p.formState(); fs != nil {
		field, _ = p.control(fs.FocusID)
	}

	if activationKey(key) && p.activationTarget(field) {
		_, err := p.fireTriggers(ctx, p.boxByID(field.ID), true)
		if err != nil {
			return err
		}
	}

	switch gesture.KeyDown(field, field.Value, key) {
	case gesture.EffectBlur:
		return p.Focus(ctx, "")
	case gesture.EffectCaret:
		return p.caretKey(ctx, key)
	default:
		return nil
	}
}

// KeyUp calls the key-up handler. It does not draw. A nil handler does
// nothing.
func (p *Page) KeyUp(ctx context.Context, key string) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if p.handlers.KeyUp == nil {
		return nil
	}

	if err := p.handlers.KeyUp(ctx, key); err != nil {
		return err
	}
	p.markFormDirty()

	return nil
}

// Type calls the type handler and edits a focused text field.
// It draws when the handler ran or the field changed.
// A nil handler with nothing to edit does nothing.
// The mobile window shows a preedit by typing and backspacing around it,
// so a platform text input session leaves its committed text here.
func (p *Page) Type(ctx context.Context, text string) error {
	var fn func() error
	if p.handlers.Type != nil {
		fn = func() error {
			return p.handlers.Type(ctx, text)
		}
	}

	return p.editField(ctx, fn, func(c Control, caret, anchor int) (Control, int, int, bool) {
		return gesture.Edit(c, text, gesture.OpInsert, caret, anchor)
	})
}

// Backspace calls the backspace handler and edits a focused text field.
// It draws when the handler ran or the field changed.
// A nil handler with nothing to edit does nothing.
func (p *Page) Backspace(ctx context.Context) error {
	var fn func() error
	if p.handlers.Backspace != nil {
		fn = func() error {
			return p.handlers.Backspace(ctx)
		}
	}

	return p.editField(ctx, fn, func(c Control, caret, anchor int) (Control, int, int, bool) {
		return gesture.Edit(c, "", gesture.OpBackspace, caret, anchor)
	})
}

// DeleteWord calls the delete-word handler and edits a focused text field.
// It draws when the handler ran or the field changed.
// A nil handler with nothing to edit does nothing.
func (p *Page) DeleteWord(ctx context.Context) error {
	var fn func() error
	if p.handlers.DeleteWord != nil {
		fn = func() error {
			return p.handlers.DeleteWord(ctx)
		}
	}

	return p.editField(ctx, fn, func(c Control, caret, anchor int) (Control, int, int, bool) {
		return gesture.Edit(c, "", gesture.OpDeleteWord, caret, anchor)
	})
}

// Copy returns the text the page wants on the clipboard.
// A focused text field returns its selected range, or its whole value when
// the range is empty, and does not call the handler.
// ok is false when there is nothing to copy. Copy does not draw.
func (p *Page) Copy(ctx context.Context) (string, bool, error) {
	if err := useContext(ctx); err != nil {
		return "", false, err
	}

	if c, ok := p.focusedEditable(); ok {
		return gesture.Selected(c.Value, p.formState().Anchor, p.formState().Caret), true, nil
	}

	if p.handlers.Copy == nil {
		return "", false, nil
	}

	return p.handlers.Copy(ctx)
}

// Cut returns the text the page wants on the clipboard and draws again.
// A focused text field returns its selected range, or its whole value when
// the range is empty, removes that text, writes the bound field, and fires
// Change before the redraw. It does not call the handler.
// ok is false when there is nothing to cut.
func (p *Page) Cut(ctx context.Context) (string, bool, error) {
	if err := useContext(ctx); err != nil {
		return "", false, err
	}

	if c, ok := p.focusedEditable(); ok {
		id := p.formState().FocusID
		if err := p.beforeEdit(ctx, id); err != nil {
			return "", false, err
		}

		c, text, pos := gesture.Cut(c, p.formState().Anchor, p.formState().Caret)
		p.formState().Caret, p.formState().Anchor = pos, pos
		p.formState().All = false
		p.resetCaretBlink()
		p.formState().ByID[id] = c
		p.markPending(id)
		bindWrite(p, c)
		if err := p.change(ctx, id); err != nil {
			return text, true, err
		}

		err := p.Redraw(ctx)

		return text, true, err
	}

	if p.handlers.Cut == nil {
		return "", false, nil
	}

	text, ok, err := p.handlers.Cut(ctx)
	if err != nil || !ok {
		return text, ok, err
	}

	if err := p.Redraw(ctx); err != nil {
		return text, ok, err
	}

	return text, ok, nil
}

// Paste calls the paste handler and edits a focused text field.
// It draws when the handler ran or the field changed.
// A nil handler with nothing to edit does nothing.
func (p *Page) Paste(ctx context.Context, text string) error {
	var fn func() error
	if p.handlers.Paste != nil {
		fn = func() error {
			return p.handlers.Paste(ctx, text)
		}
	}

	// GEN-28: cap the paste before it merges into the field. The handler
	// closure reads text, so the cap has to land before editField runs it.
	text = gesture.ClampPaste(text)

	return p.editField(ctx, fn, func(c Control, caret, anchor int) (Control, int, int, bool) {
		return gesture.Edit(c, text, gesture.OpPaste, caret, anchor)
	})
}

// SelectAll selects the whole value of a focused text field, or calls the
// select-all handler. A focused field does not call the handler. It draws
// after selecting. It does nothing when no field is focused and no handler
// is registered.
func (p *Page) SelectAll(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if c, ok := p.typingTarget(); ok {
		sel := gesture.SelectAll(c.Value)
		p.formState().Anchor, p.formState().Caret = sel.Anchor, sel.Caret
		p.formState().All = sel.All
		p.resetCaretBlink()

		return p.Redraw(ctx)
	}

	return p.after(ctx, p.handlers.SelectAll == nil, func() error {
		return p.handlers.SelectAll(ctx)
	})
}

// Undo calls the undo handler and draws the page again.
// It does nothing when no undo handler is registered.
func (p *Page) Undo(ctx context.Context) error {
	return p.after(ctx, p.handlers.Undo == nil, func() error {
		return p.handlers.Undo(ctx)
	})
}

// Redo calls the redo handler and draws the page again.
// It does nothing when no redo handler is registered.
func (p *Page) Redo(ctx context.Context) error {
	return p.after(ctx, p.handlers.Redo == nil, func() error {
		return p.handlers.Redo(ctx)
	})
}

// Submit calls the submit handler and draws the page again.
// It does nothing when no submit handler is registered.
func (p *Page) Submit(ctx context.Context) error {
	return p.after(ctx, p.handlers.Submit == nil, func() error {
		return p.handlers.Submit(ctx)
	})
}

func (p *Page) after(ctx context.Context, skip bool, fn func() error) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if skip {
		return nil
	}

	if err := fn(); err != nil {
		return err
	}

	// The handler can mutate the template data in place through the
	// SetData pointer, so force the next Redraw to re-execute.
	p.markFormDirty()

	return p.Redraw(ctx)
}

// BeginDrag offers the box under the point to Handlers.DragStart.
func (p *Page) BeginDrag(ctx context.Context, x, y float64) (bool, error) {
	if err := useContext(ctx); err != nil {
		return false, err
	}

	if p.handlers.DragStart == nil {
		return false, nil
	}

	box := p.boxByID(p.boxIDAt(x, y))
	claimed, err := p.handlers.DragStart(ctx, box)
	if gesture.Claimed(err, claimed) {
		p.markFormDirty()
		p.gesture = p.gesture.Begin(x, y)
		err = p.Press(ctx, x, y)
	}

	return claimed, err
}

// MoveDrag sends movement while a claimed gesture is active.
func (p *Page) MoveDrag(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	next, dx, dy, send := p.gesture.Move(x, y)
	p.gesture = next
	if p.handlers.DragMove == nil || !send {
		return nil
	}

	if err := p.handlers.DragMove(ctx, dx, dy); err != nil {
		return err
	}
	p.markFormDirty()

	return nil
}

// EndDrag releases the claimed gesture without synthesizing a click.
func (p *Page) EndDrag(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	next, active := p.gesture.End()
	p.gesture = next
	if !active {
		return nil
	}

	if p.handlers.DragEnd != nil {
		if err := p.handlers.DragEnd(ctx); err != nil {
			return err
		}
		p.markFormDirty()
	}

	return p.Release(ctx)
}

// Swipe reports a one-finger swipe after the finger lifts. A page without
// a Swipe handler ignores the gesture. It does not draw: change state and
// let a SetTick callback paint, or call Redraw.
func (p *Page) Swipe(ctx context.Context, dx, dy float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	dx, dy, ok := gesture.Swipe(p.handlers.Swipe != nil, dx, dy)
	if !ok {
		return nil
	}

	if err := p.handlers.Swipe(ctx, dx, dy); err != nil {
		return err
	}
	p.markFormDirty()

	return nil
}

// LongPress handles a press held in place at a point. A text control under
// the point selects the word there, the way a double click does, and claims
// the press. Otherwise a Handlers.LongPress runs on the innermost box under
// the point, found the way Click finds it, and the press is claimed. A nil
// handler reports false, so the window keeps its ordinary gesture.
func (p *Page) LongPress(ctx context.Context, x, y float64) (bool, error) {
	if err := useContext(ctx); err != nil {
		return false, err
	}

	id := p.boxIDAt(x, y)
	if c, ok := p.control(id); gesture.ClaimsWord(c, ok) {
		return true, p.SelectWordAt(ctx, x, y)
	}

	box, ok := p.hitBox(x, y)
	if !ok || p.handlers.LongPress == nil {
		return false, nil
	}

	if box.ID == "" {
		if id := p.boxIDAt(x, y); id != "" {
			box = p.boxByID(id)
		}
	}

	if err := p.handlers.LongPress(ctx, box); err != nil {
		return true, err
	}
	p.markFormDirty()

	return true, p.Redraw(ctx)
}

// Drop calls the drop handler with the files dropped on the window and runs
// the redraw sequence again. An error returns before the redraw. When the
// handler changed no page input, the frame is identical and the redraw
// commits nothing.
func (p *Page) Drop(ctx context.Context, files []Drop) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	files = gesture.Limit(files)
	if p.handlers.Drop != nil {
		if err := p.handlers.Drop(ctx, files); err != nil {
			return err
		}
		p.markFormDirty()
	}

	return p.Redraw(ctx)
}

// Hover sets the hovered element from a point and draws when it changed.
// The innermost box with an id under the point wins; an empty point clears it.
func (p *Page) Hover(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	p.setHoverPoint(x, y)

	return p.setHover(ctx, p.boxIDAt(x, y))
}

// Press sets the pressed element from a point and draws when it changed.
func (p *Page) Press(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id := p.boxIDAt(x, y)
	if gesture.SameHit(p.activeID(), id) {
		return nil
	}

	if box, ok := p.activationBox(x, y); ok {
		if err := p.pressTriggers(ctx, box); err != nil {
			return err
		}
	}

	p.markPair(p.activeID(), id)
	p.setActiveID(id)
	p.setPressUsed(false)
	c, ok := p.control(id)
	if gesture.RunPress(p.handlers.Press != nil, ok, c.Disabled) {
		used, err := p.handlers.Press(ctx, p.boxByID(id), x, y)
		if err != nil {
			return err
		}
		p.markFormDirty()

		p.setPressUsed(used)
	}

	// No :active paint to show, so wait for the click handler's redraw.
	if !p.statePseudos.Active {
		return nil
	}

	return p.Redraw(ctx)
}

// Release clears the pressed element and draws when it changed.
func (p *Page) Release(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if err := p.releaseTriggers(ctx); err != nil {
		return err
	}

	old := p.takePress()
	if old == "" {
		return nil
	}

	p.markPair(old, "")
	if !p.statePseudos.Active {
		return nil
	}

	return p.Redraw(ctx)
}

func (p *Page) setHover(ctx context.Context, id string) error {
	if gesture.SameHit(p.HoverID(), id) {
		return nil
	}

	handled, err := p.paintHover(ctx, id)
	if err != nil {
		return err
	}

	p.markPair(p.HoverID(), id)
	p.setHoverID(id)
	if gesture.SkipHoverRedraw(handled, p.statePseudos.Hover) {
		return nil
	}

	return p.Redraw(ctx)
}

func (p *Page) paintHover(ctx context.Context, id string) (bool, error) {
	if !gesture.CanPaintHover(p.handlers.Hover != nil, p.draw.Display != nil) {
		return false, nil
	}

	handled, err := p.handlers.Hover(ctx, p.boxByID(p.HoverID()), p.boxByID(id))
	if err != nil {
		return false, err
	}
	// The handler ran and can mutate the template data in place, whether
	// or not it claimed the hover.
	p.markFormDirty()

	return handled, nil
}

// boxIDAt returns the innermost box with an id that contains the point, or
// "". It delegates to the shared boxAt scan so Hover and Press agree with
// Click on overlapping boxes.
func (p *Page) boxIDAt(x, y float64) string {
	return p.boxAt(x, y).id
}

// PressedID is the control held by the current pointer gesture.
// It is empty after Release.
func (p *Page) PressedID() string { return p.activeID() }

// CancelTouches queues gesture cancellation for the window loop. A native
// host may call it from its UI thread before clearing platform touch IDs.
func (p *Page) CancelTouches() { gesture.QueueCancel(&p.touchCancel) }

// TakeTouchCancel consumes a queued cancellation. Custom screens can expose
// this method to let the window drop active gestures without delivering a tap.
func (p *Page) TakeTouchCancel() bool { return gesture.TakeCancel(&p.touchCancel) }

// SetTouchZoomAllowed controls whether two-finger touch gestures zoom the page.
// It does not change scrolling, wheel zoom, keyboard zoom, or fitting.
func (p *Page) SetTouchZoomAllowed(on bool) {
	if p != nil {
		p.allowTouchZoom = on
	}
}

// TouchZoomAllowed reports whether two-finger touch gestures may zoom the page.
func (p *Page) TouchZoomAllowed() bool {
	return gesture.TouchZoom(p == nil, p != nil && p.allowTouchZoom)
}

// WheelZoomAllowed reports whether Ctrl+wheel or Command+wheel may scale the
// page. A nil page keeps the default, which is on.
func (p *Page) WheelZoomAllowed() bool {
	return p == nil || !p.disableWheelZoom
}

var _ host.WheelZoomPolicy = (*Page)(nil)

// KeyZoomAllowed reports whether Ctrl+=, Ctrl++, Ctrl+-, and Ctrl+0 may
// scale the page. A nil page keeps the default, which is on.
func (p *Page) KeyZoomAllowed() bool {
	return p == nil || !p.disableKeyZoom
}

var _ host.KeyZoomPolicy = (*Page)(nil)

// ContextMenu returns the context-menu rows for the current state. Cut and
// copy need a selection, paste needs a focused editable field, select all
// needs either, and undo and redo need their handler.
func (p *Page) ContextMenu() []host.MenuItem {
	_, editable := p.typingTarget()

	return gesture.ContextMenu(gesture.Menu{
		Selection: p.hasSelection(),
		Editable:  editable,
		SelectAll: editable || p.handlers.SelectAll != nil,
		Undo:      p.handlers.Undo != nil,
		Redo:      p.handlers.Redo != nil,
	})
}

// CursorShape returns the shape for the hovered box: an I-beam over a text
// control, a hand over a link, a button, or another control, and the default
// arrow everywhere else.
func (p *Page) CursorShape() host.Shape {
	in := gesture.Cursor{}
	if id := p.HoverID(); id != "" {
		in.HoverID = id
		if c, ok := p.control(id); ok {
			in.Control = true
			in.Editable = canEdit(c)
		} else {
			b := p.boxByID(id)
			in.Tag = b.Tag
			in.Action = b.Action
		}
	}

	x, y := p.hoverPoint()
	_, in.Link = p.linkTargetAt(x, y)

	return gesture.CursorShape(in)
}

// Page is an IME target for the window.
var _ host.IME = (*Page)(nil)

// IMEContext returns the focused text control's box, the caret as a byte
// offset into the value, and the value around the caret. ok is false when
// no text control is focused.
func (p *Page) IMEContext() (layout.Box, int, string, string, bool) {
	c, ok := p.typingTarget()
	if !ok {
		return layout.Box{}, 0, "", "", false
	}

	// The window reports the caret to the platform in screen coordinates,
	// so take the scroll offset out of the box. Ebiten pans the canvas when
	// the keyboard would cover this rectangle.
	before, after, n := gesture.SplitCaret(c.Value, p.formState().Caret)
	box := gesture.Unscroll(p.boxByID(c.ID), float64(p.scroll.Windowing.X), float64(p.scroll.Windowing.Y))

	return box, n, before, after, true
}

var _ host.IMECaretAnchor = (*Page)(nil)

// IMEAnchor returns the focused glyph caret in CSS pixels, adjusted for page
// scrolling. False means the current display cannot measure the caret.
func (p *Page) IMEAnchor() (layout.Box, bool) {
	c, ok := p.typingTarget()
	if !ok {
		return layout.Box{}, false
	}

	fs := p.formState()
	pre := gesture.Preedit{
		ID:    fs.PreeditID,
		Text:  fs.Preedit,
		Start: fs.PreeditStart,
		End:   fs.PreeditEnd,
	}

	return gesture.Anchor(
		p.draw.Display,
		c,
		c.Value,
		fs.Caret,
		fs.Anchor,
		pre,
		float64(p.scroll.Windowing.X),
		float64(p.scroll.Windowing.Y),
	)
}

// SetIMEComposition paints temporary IME text without storing or notifying
// the field value. start and end select bytes within text.
func (p *Page) SetIMEComposition(ctx context.Context, text string, start, end int) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	fs := p.formState()
	if fs == nil {
		return nil
	}

	_, has := p.typingTarget()
	cur := gesture.Preedit{
		ID:    fs.PreeditID,
		Text:  fs.Preedit,
		Start: fs.PreeditStart,
		End:   fs.PreeditEnd,
	}
	next, redraw := gesture.Compose(cur, fs.FocusID, text, start, end, has)
	if !redraw {
		return nil
	}

	fs.PreeditID, fs.Preedit = next.ID, next.Text
	fs.PreeditStart, fs.PreeditEnd = next.Start, next.End

	return p.Redraw(ctx)
}

// IMEReplace replaces the focused value's bytes [start, end) with insert
// and puts the caret at the byte offset caret within insert. It draws when
// the value changed.
func (p *Page) IMEReplace(ctx context.Context, start, end int, insert string, caret int) error {
	c, ok := p.typingTarget()
	if !ok {
		return nil
	}

	value, pos, changed := gesture.Replace(c, insert, start, end, caret)
	if !changed {
		return nil
	}

	return p.editField(ctx, nil, func(cur Control, _, _ int) (Control, int, int, bool) {
		cur.Value = value

		return cur, pos, pos, true
	})
}

func useContext(ctx context.Context) error {
	if ctx == nil {
		return errNilContext
	}

	return ctx.Err()
}

var _ host.Dropper = (*Page)(nil)
