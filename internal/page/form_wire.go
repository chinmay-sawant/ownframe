package page

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// Control is one form field taken from the executed HTML.
type Control = form.Control

// Option is one choice inside a select.
type Option = form.Option

// controlSpan is one control and where it sits in the executed HTML.
// Start is the index of '<'. End is the index after the element.
type controlSpan = form.Span

// formState keeps the field names page_form_focus, gesture_wire, a11y_wire,
// and page_perf already read. snap and store move those fields to form.State,
// which owns the caret, focus, and value changes.
type formState struct {
	byID         map[string]Control
	order        []string
	focusID      string
	lastFocusID  string
	caret        int
	anchor       int
	all          bool
	doc          int
	blinkOn      bool
	blinkAt      time.Time
	preeditID    string
	preedit      string
	preeditStart int
	preeditEnd   int
}

const (
	commentOpenLen  = form.CommentOpenLen
	commentTotalLen = form.CommentTotalLen
	// caretBlinkInterval is how long the caret stays shown or hidden.
	caretBlinkInterval = form.BlinkInterval
	// DefaultMaxFieldLen caps one field value when the template sets no
	// maxlength (GEN-28).
	DefaultMaxFieldLen = form.DefaultMaxFieldLen
	// MaxPasteRunes caps one paste or type call before it merges.
	MaxPasteRunes = form.MaxPasteRunes
)

func (s *formState) snap() form.State {
	return form.State{
		ByID:         s.byID,
		Order:        s.order,
		FocusID:      s.focusID,
		LastFocusID:  s.lastFocusID,
		Caret:        s.caret,
		Anchor:       s.anchor,
		All:          s.all,
		Doc:          s.doc,
		BlinkOn:      s.blinkOn,
		BlinkAt:      s.blinkAt,
		PreeditID:    s.preeditID,
		Preedit:      s.preedit,
		PreeditStart: s.preeditStart,
		PreeditEnd:   s.preeditEnd,
	}
}

func (s *formState) store(st *form.State) {
	s.byID = st.ByID
	s.order = st.Order
	s.focusID = st.FocusID
	s.lastFocusID = st.LastFocusID
	s.caret = st.Caret
	s.anchor = st.Anchor
	s.all = st.All
	s.doc = st.Doc
	s.blinkOn = st.BlinkOn
	s.blinkAt = st.BlinkAt
	s.preeditID = st.PreeditID
	s.preedit = st.Preedit
	s.preeditStart = st.PreeditStart
	s.preeditEnd = st.PreeditEnd
}

func (p *Page) editState(fn func(*form.State)) {
	fs := p.formState()
	st := fs.snap()
	fn(&st)
	fs.store(&st)
}

// bounds normalizes the form range to [start, end) clamped to n runes.
func (s *formState) bounds(n int) (int, int) {
	st := s.snap()

	return st.Bounds(n)
}

func textLike(kind string) bool { return form.TextLike(kind) }

func canEdit(c Control) bool { return form.CanEdit(c) }

func runeLen(s string) int { return form.RuneLen(s) }

func readOpen(s string, i int) (string, string, int, bool) {
	return form.ReadOpen(s, i)
}

func closeSpan(s string, i int, tag string) (int, int) {
	return form.CloseSpan(s, i, tag)
}

func attrValue(raw, name string) string {
	return form.AttrValue(raw, name)
}

func openTagDrop(tag, raw string, extra []string, drop func(string) bool) string {
	return form.OpenTagDrop(tag, raw, extra, drop)
}

func rewriteControls(source string, spans []controlSpan, live map[string]Control, focusID string, selected bool) string {
	return form.Rewrite(source, spans, live, focusID, selected)
}

// insertValue inserts text at the caret and replaces the selection.
func (p *Page) insertValue(c Control, caret, anchor int, text string) (Control, int, int, bool) {
	return form.InsertValue(c, caret, anchor, text)
}

// valueEdit computes one edit to a field: the new value, the caret and the
// anchor to store, and whether anything changed.
type valueEdit func(c Control, caret, anchor int) (Control, int, int, bool)

// PickFunc opens a file dialog and returns the path the person chose.
type PickFunc func(ctx context.Context, title string) (path string, err error)

// SetPicker installs the dialog a file control opens.
func (p *Page) SetPicker(pick PickFunc) {
	if p == nil {
		return
	}

	p.picker = pick
}

// InstallPicker sets the dialog a click on a file control opens.
// Run installs the picker for the current system; tests install a fake.
func InstallPicker(p interface{ SetPicker(PickFunc) }, pick PickFunc) {
	if p != nil {
		p.SetPicker(pick)
	}
}

// FocusNext focuses the next control in document order and wraps. Disabled
// controls and tabindex="-1" are skipped. It draws when the focus moved.
func (p *Page) FocusNext(ctx context.Context) error {
	return p.focusStep(ctx, 1)
}

// FocusPrev focuses the previous control in document order and wraps.
// Disabled controls and tabindex="-1" are skipped. It draws when the focus
// moved.
func (p *Page) FocusPrev(ctx context.Context) error {
	return p.focusStep(ctx, -1)
}

// Focus focuses the control id and draws. An empty id blurs. A disabled
// control or an unknown id is left alone.
func (p *Page) Focus(ctx context.Context, id string) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if id == "" {
		fs := p.formState()
		if fs == nil || fs.focusID == "" {
			return nil
		}

		p.blurForm()

		return p.Redraw(ctx)
	}

	c, ok := p.control(id)
	if !ok || c.Disabled {
		return nil
	}

	var moved, reveal bool
	p.editState(func(st *form.State) {
		moved, reveal = st.TakeFocus(id, c, ok)
	})
	if !moved {
		return nil
	}

	if reveal {
		p.revealFocus(id)
	}

	return p.Redraw(ctx)
}

// FocusID returns the focused control id, or "" when none is focused.
func (p *Page) FocusID() string {
	fs := p.formState()
	if fs == nil {
		return ""
	}

	return fs.focusID
}

// LastFocusID returns the control id cleared by the most recent blur.
func (p *Page) LastFocusID() string {
	fs := p.formState()
	if fs == nil {
		return ""
	}

	return fs.lastFocusID
}

func (p *Page) focusStep(ctx context.Context, step int) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id, ok := form.NextID(p.focusOrder(), p.FocusID(), step)
	if !ok {
		return nil
	}

	return p.Focus(ctx, id)
}

// focusOrder is the tab order: the document order of the control boxes,
// with disabled controls and tabindex="-1" left out.
func (p *Page) focusOrder() []string {
	ids := make([]string, 0, len(p.boxes))
	for _, b := range p.boxes {
		ids = append(ids, b.ID)
	}

	return form.Order(ids, p.control)
}

// SelectAt focuses the control under the point and puts the caret at the
// measured rune offset. A point outside every control blurs the form. A
// disabled control is left alone. A control that takes no text just takes
// focus.
func (p *Page) SelectAt(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id := p.boxIDAt(x, y)
	c, ok := p.control(id)
	if !ok {
		fs := p.formState()
		if fs == nil || fs.focusID == "" {
			return nil
		}

		p.blurForm()

		return p.Redraw(ctx)
	}

	if c.Disabled {
		return nil
	}

	if !canEdit(c) {
		return p.Focus(ctx, id)
	}

	return p.focusCaret(ctx, c, p.offsetAt(p.boxByID(id), c, x, y))
}

// Drag extends the focused field's range to the point. It does nothing when
// no text field is focused.
func (p *Page) Drag(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	c, ok := p.typingTarget()
	if !ok {
		return nil
	}

	off := p.offsetAt(p.boxByID(c.ID), c, x, y)
	changed := false
	p.editState(func(st *form.State) {
		changed = st.DragTo(off)
	})
	if !changed {
		return nil
	}

	return p.Redraw(ctx)
}

// SelectWordAt selects the word under the point and draws.
func (p *Page) SelectWordAt(ctx context.Context, x, y float64) error {
	return p.selectSpan(ctx, x, y, form.WordBounds)
}

// SelectLineAt selects the line under the point and draws.
func (p *Page) SelectLineAt(ctx context.Context, x, y float64) error {
	return p.selectSpan(ctx, x, y, form.LineBounds)
}

func (p *Page) selectSpan(ctx context.Context, x, y float64, bounds func(string, int) (int, int)) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id := p.boxIDAt(x, y)
	c, ok := p.control(id)
	if !ok || c.Disabled || !canEdit(c) {
		return nil
	}

	off := p.offsetAt(p.boxByID(id), c, x, y)
	start, end := bounds(c.Value, off)
	p.editState(func(st *form.State) {
		st.SelectSpan(c.ID, start, end)
	})

	return p.Redraw(ctx)
}

// offsetAt turns a point into a rune offset in c's value. A run that cannot
// be measured, or one whose drawn text does not line up with the shown
// value, lands at the end of the value.
func (p *Page) offsetAt(box Box, c Control, x, y float64) int {
	return form.OffsetAt(p.draw.Display, box, c, x, y)
}

// focusCaret focuses c and puts the caret and the anchor at off.
func (p *Page) focusCaret(ctx context.Context, c Control, off int) error {
	changed := false
	p.editState(func(st *form.State) {
		changed = st.FocusCaret(c.ID, off)
	})
	if !changed {
		return nil
	}

	return p.Redraw(ctx)
}

func (p *Page) revealFocus(id string) {
	box := p.boxByID(id)
	if p.isFixed(box) {
		// A box the window pins to the viewport is always visible, so
		// revealing it must not scroll the page away from the content.
		return
	}

	x, y := p.ScrollOffset()
	x, y = form.RevealOrigin(box.X, box.Y, box.W, box.H, x, y, p.width, p.height)
	p.ScrollTo(x, y)
}

// setCaret moves the caret to pos, keeping the anchor when extend is true.
func (p *Page) setCaret(pos int, extend bool) {
	if p.formState() == nil {
		return
	}

	n := 0
	if c, ok := p.focusedEditable(); ok {
		n = runeLen(c.Value)
	}

	p.editState(func(st *form.State) {
		st.SetCaret(pos, n, extend)
	})
}

// hasSelection reports whether the focused field has a non-empty range.
func (p *Page) hasSelection() bool {
	c, ok := p.typingTarget()
	if !ok {
		return false
	}

	start, end := p.formState().bounds(runeLen(c.Value))

	return start != end
}

// caretKey moves the caret for a focused text field and draws. Arrow keys
// move by rune, Home and End move to the ends, and a ctrl+ or alt+ prefix
// jumps by word. A shift+ prefix extends the range from the anchor.
func (p *Page) caretKey(ctx context.Context, key string) error {
	c, ok := p.typingTarget()
	if !ok {
		return nil
	}

	fs := p.formState()
	st := fs.snap()
	n := runeLen(c.Value)
	caret := form.ClampPos(st.Caret, n)
	start, end := st.Bounds(n)
	pos, extend, ok := form.MoveCaret(key, c.Value, caret, start, end)
	if !ok {
		return nil
	}

	if !form.CaretMoved(st.Caret, st.Anchor, pos, st.All, extend) {
		return nil
	}

	st.SetCaret(pos, n, extend)
	fs.store(&st)

	return p.Redraw(ctx)
}

// resetCaretBlink shows the caret and restarts the blink clock. Every caret
// move and edit calls it, so the caret is solid while the person types.
func (p *Page) resetCaretBlink() {
	fs := p.formState()
	if fs == nil {
		return
	}

	fs.blinkOn, fs.blinkAt = form.ResetBlink()
}

// caretBlink toggles the focused field's caret on the blink clock and draws
// when the phase changes. A range selection paints no caret, so it does
// nothing. A file input never shows a caret, so it does nothing.
func (p *Page) caretBlink(ctx context.Context) error {
	c, ok := p.focusedEditable()
	fs := p.formState()
	if fs == nil {
		return nil
	}

	n := 0
	if ok {
		n = runeLen(c.Value)
	}

	st := fs.snap()
	draw := st.StepBlink(time.Now(), n, ok, form.IsFileInput(c))
	fs.store(&st)
	if !draw {
		return nil
	}

	return p.Redraw(ctx)
}

func (p *Page) paintCaret(d *layout.Display) {
	fs := p.formState()
	if fs == nil {
		return
	}

	c, ok := p.focusedEditable()
	st := fs.snap()
	v := st.View()
	if ok {
		v.Control = c
	}

	form.PaintCaret(d, v, ok, form.IsFileInput(c), fs.blinkOn)
}

func (p *Page) syncForm(body string) string {
	doc := p.nav.Doc()
	fs := p.formState()
	if fs == nil || fs.doc != doc {
		fs = &formState{}
		fresh := form.NewDoc(doc)
		fs.store(&fresh)
		p.setFormState(fs)
	}

	body = rewriteButtons(body)
	spans := form.Scan(body)
	st := fs.snap()
	byID, order := form.Merge(st.ByID, spans)
	form.ApplyBinds(byID, func(c Control) (Control, bool) {
		return bindRead(p, c)
	})
	st.ByID = byID
	st.Order = order
	if _, ok := byID[st.FocusID]; !ok {
		st.FocusID = ""
		st.ClearRange()
	}

	st.ClampToFocus()
	fs.store(&st)

	return emojiPass(form.RewriteState(body, spans, byID, form.CaretOf(st.View())))
}

func (p *Page) clickControl(ctx context.Context, box Box) error {
	c, ok := p.control(box.ID)
	if ok && c.Disabled {
		return nil
	}

	if c.Type == "file" {
		return p.clickFile(ctx, box, c)
	}

	if !c.Disabled && form.WillChange(c) {
		if err := p.beforeEdit(ctx, box.ID); err != nil {
			return err
		}
	}

	oldFocus := p.currentFocus()
	changed := false
	if !c.Disabled {
		changed = p.activate(box.ID)
	}

	if p.currentFocus() != oldFocus {
		p.markPair(oldFocus, p.currentFocus())
	}

	if changed {
		if c.Type == "checkbox" || c.Type == "radio" {
			p.markToggle(box.ID)
		}

		if next, ok := p.control(box.ID); ok {
			bindWrite(p, next)
		}

		if err := p.change(ctx, box.ID); err != nil {
			return err
		}
	}

	if p.handlers.Click != nil {
		if err := p.handlers.Click(ctx, box); err != nil {
			return err
		}
	}

	painted, err := p.clickPainted(ctx, box)
	if err != nil {
		return err
	}

	if painted {
		return nil
	}

	return p.Redraw(ctx)
}

// clickFile focuses a file control and stores the path the picker returns.
// Cancel or an empty path keeps the old value, still runs Click, and
// redraws. Any other picker error redraws the focus and returns wrapped.
func (p *Page) clickFile(ctx context.Context, box Box, c Control) error {
	oldFocus := p.currentFocus()
	p.activate(box.ID)
	if p.currentFocus() != oldFocus {
		p.markPair(oldFocus, p.currentFocus())
	}

	if p.picker != nil {
		if err := p.applyPicked(ctx, box, c); err != nil {
			return err
		}
	}

	if p.handlers.Click != nil {
		if err := p.handlers.Click(ctx, box); err != nil {
			return err
		}
	}

	return p.Redraw(ctx)
}

func (p *Page) applyPicked(ctx context.Context, box Box, c Control) error {
	path, err := p.picker(ctx, p.title)
	if err != nil && !errors.Is(err, filepick.ErrCanceled) {
		if rerr := p.Redraw(ctx); rerr != nil {
			err = errors.Join(err, rerr)
		}

		return fmt.Errorf("ownframe: %w", err)
	}

	if err != nil || path == "" || path == c.Value {
		return nil
	}

	if err := p.beforeEdit(ctx, box.ID); err != nil {
		return err
	}

	c.Value = path
	p.formState().byID[box.ID] = c
	bindWrite(p, c)
	p.markPending(box.ID)

	return p.change(ctx, box.ID)
}

func (p *Page) blurForm() {
	fs := p.formState()
	if fs == nil {
		return
	}

	st := fs.snap()
	old := st.Blur()
	fs.store(&st)
	p.markPair(old, "")
}

// currentFocus is the focused control id, or "" when none is focused.
func (p *Page) currentFocus() string {
	return p.FocusID()
}

func (p *Page) formControl(id string) bool {
	if id == "" {
		return false
	}

	_, ok := p.control(id)

	return ok
}

func (p *Page) activate(id string) bool {
	fs := p.formState()
	if fs == nil {
		return false
	}

	changed := false
	p.editState(func(st *form.State) {
		changed = st.Activate(id)
	})

	return changed
}

func (p *Page) editField(ctx context.Context, fn func() error, edit valueEdit) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	ran := false
	if fn != nil {
		if err := fn(); err != nil {
			return err
		}

		ran = true
	}

	changed, err := p.applyEdit(ctx, edit)
	if err != nil {
		return err
	}

	if !changed && !ran {
		return nil
	}

	if changed {
		p.markPending(p.formState().focusID)
		if err := p.change(ctx, p.formState().focusID); err != nil {
			return err
		}
	}

	return p.Redraw(ctx)
}

func (p *Page) applyEdit(ctx context.Context, edit valueEdit) (bool, error) {
	c, ok := p.typingTarget()
	if !ok {
		return false, nil
	}

	fs := p.formState()
	next, caret, anchor, changed := edit(c, fs.caret, fs.anchor)
	if !changed {
		return false, nil
	}

	// GEN-28: every edit funnels here, so clamp once for all writers.
	next, caret, anchor = form.ClampEdited(next, caret, anchor)
	if err := p.beforeEdit(ctx, fs.focusID); err != nil {
		return false, err
	}

	fs = p.formState()
	fs.byID[fs.focusID] = next
	fs.caret, fs.anchor = caret, anchor
	fs.all = false
	p.resetCaretBlink()
	bindWrite(p, next)

	return true, nil
}

func (p *Page) typingTarget() (Control, bool) {
	c, ok := p.focusedEditable()
	if !ok || c.Disabled {
		return Control{}, false
	}

	return c, true
}

func (p *Page) focusedEditable() (Control, bool) {
	fs := p.formState()
	if fs == nil {
		return Control{}, false
	}

	c, ok := p.control(fs.focusID)
	if !ok || !canEdit(c) {
		return Control{}, false
	}

	return c, true
}

// beforeEdit calls the BeforeEdit handler for id.
// A nil handler does nothing.
func (p *Page) beforeEdit(ctx context.Context, id string) error {
	if p.handlers.BeforeEdit == nil {
		return nil
	}

	return p.handlers.BeforeEdit(ctx, p.boxByID(id))
}

// change calls the Change handler for id.
// A nil handler does nothing.
func (p *Page) change(ctx context.Context, id string) error {
	if p.handlers.Change == nil {
		return nil
	}

	return p.handlers.Change(ctx, p.boxByID(id))
}

// FormValue returns the current value of id, or "" when id is absent.
func (p *Page) FormValue(id string) string {
	c, _ := p.control(id)

	return c.Value
}

// SetFormValue stores value for id and does not draw.
// It does nothing when that control is not on the page yet.
// A select marks each option whose value matches.
// The focused control loses its selection.
func (p *Page) SetFormValue(id, value string) {
	fs := p.formState()
	if fs == nil {
		return
	}

	p.editState(func(st *form.State) {
		st.WriteValue(id, value)
	})
}

// FormChecked reports whether id is checked.
// A missing control returns false.
func (p *Page) FormChecked(id string) bool {
	c, _ := p.control(id)

	return c.Checked
}

// SetFormChecked sets a checkbox or radio and does not draw.
// It does nothing when id is missing.
// A radio set true clears other radios with the same non-empty name.
// A radio set false clears only that radio.
func (p *Page) SetFormChecked(id string, checked bool) {
	fs := p.formState()
	if fs == nil {
		return
	}

	p.editState(func(st *form.State) {
		st.WriteChecked(id, checked)
	})
}

// FocusedField returns the focused field id, excluding buttons.
// FocusID also reports button focus.
func (p *Page) FocusedField() string {
	fs := p.formState()
	if fs == nil {
		return ""
	}

	st := fs.snap()

	return st.FieldID()
}

func (p *Page) control(id string) (Control, bool) {
	fs := p.formState()
	if fs == nil {
		return Control{}, false
	}

	st := fs.snap()

	return st.Lookup(id)
}

// FormSelected reports whether id is the focused control and its whole
// value is selected. It is false when no control is focused, when the
// page has no form yet, or when another control is focused.
func (p *Page) FormSelected(id string) bool {
	fs := p.formState()
	if fs == nil || fs.focusID != id {
		return false
	}

	st := fs.snap()

	return form.WholeSelected(st.View(), id)
}
