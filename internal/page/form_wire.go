package page

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// Control is one form field taken from the executed HTML.
type Control = form.Control

// Option is one choice inside a select.
type Option = form.Option

// controlSpan is one control and where it sits in the executed HTML.
// Start is the index of '<'. End is the index after the element.
type controlSpan = form.Span

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
		if fs == nil || fs.FocusID == "" {
			return nil
		}

		p.blurForm()

		return p.Redraw(ctx)
	}

	c, ok := p.control(id)
	if !ok || c.Disabled || !p.inFocusScope(id) {
		return nil
	}

	moved, reveal := p.formState().TakeFocus(id, c, ok)
	if !moved {
		return nil
	}

	p.tickMutated.Store(true)
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

	return fs.FocusID
}

// LastFocusID returns the control id cleared by the most recent blur.
func (p *Page) LastFocusID() string {
	fs := p.formState()
	if fs == nil {
		return ""
	}

	return fs.LastFocusID
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
		if p.inFocusScope(b.ID) {
			ids = append(ids, b.ID)
		}
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
		if fs == nil || fs.FocusID == "" {
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
	changed := p.formState().DragTo(off)
	if !changed {
		return nil
	}

	p.tickMutated.Store(true)
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
	p.formState().SelectSpan(c.ID, start, end)
	p.tickMutated.Store(true)

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
	changed := p.formState().FocusCaret(c.ID, off)
	if !changed {
		return nil
	}

	p.tickMutated.Store(true)
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

	p.tickMutated.Store(true)
	n := 0
	if c, ok := p.focusedEditable(); ok {
		n = runeLen(c.Value)
	}

	p.formState().SetCaret(pos, n, extend)
}

// hasSelection reports whether the focused field has a non-empty range.
func (p *Page) hasSelection() bool {
	c, ok := p.typingTarget()
	if !ok {
		return false
	}

	start, end := p.formState().Bounds(runeLen(c.Value))

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
	n := runeLen(c.Value)
	caret := form.ClampPos(fs.Caret, n)
	start, end := fs.Bounds(n)
	pos, extend, ok := form.MoveCaret(key, c.Value, caret, start, end)
	if !ok {
		return nil
	}

	if !form.CaretMoved(fs.Caret, fs.Anchor, pos, fs.All, extend) {
		return nil
	}

	fs.SetCaret(pos, n, extend)
	p.tickMutated.Store(true)

	return p.Redraw(ctx)
}

// resetCaretBlink shows the caret and restarts the blink clock. Every caret
// move and edit calls it, so the caret is solid while the person types.
func (p *Page) resetCaretBlink() {
	p.tickMutated.Store(true)
	fs := p.formState()
	if fs == nil {
		return
	}

	fs.RestartBlink()
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

	draw := fs.StepBlink(time.Now(), n, ok, form.IsFileInput(c))
	if !draw {
		return nil
	}

	p.tickMutated.Store(true)
	return p.Redraw(ctx)
}

func (p *Page) paintCaret(d *layout.Display) {
	fs := p.formState()
	if fs == nil {
		return
	}

	c, ok := p.focusedEditable()
	v := fs.View()
	if ok {
		v.Control = c
	}

	form.PaintCaret(d, v, ok, form.IsFileInput(c), fs.BlinkOn)
}

// markFormDirty forces the next Redraw to re-sync form controls even when
// the template output is unchanged. It takes dataMu because syncForm reads
// and clears the flag under the same lock; a plain write here would race a
// concurrent Redraw.
func (p *Page) markFormDirty() {
	p.dataMu.Lock()
	p.formDirty = true
	p.dataMu.Unlock()
}

// its pass counter. Tests read the counter to prove the rewrite is skipped
// for span-free bodies. Phase 2 and the publish path are the only callers,
// so the count is exact.
//
//nolint:gochecknoglobals // Single process-wide call counter for tests
var rewriteButtonCalls atomic.Uint64

// syncHashCalls counts HashSource calls made by the sync path, so tests can
// prove the executed body is hashed once per changed frame and never on a
// clean frame, without timing. The memo caches stay hash-keyed; the single
// hash is threaded through the scan, the rewrite, the emoji pass, and the
// source memo instead of hashing the same body up to three times.
//
//nolint:gochecknoglobals // Single process-wide call counter for tests
var syncHashCalls atomic.Uint64

// syncHash hashes source for the sync path and counts the call for tests.
// HashSource allocates nothing, so the counted path stays alloc-free.
func syncHash(source string) uint64 {
	syncHashCalls.Add(1)

	return form.HashSource(source)
}

// formSnap pins the sync inputs snapshot under the phase-1 lock. The dirty
// flag, the last synced body, and the published source are shared with
// SetData; the live map, the order, and the caret scalars are owned by the
// page loop and only retained, never mutated, after the unlock.
type formSnap struct {
	dirty        bool
	source       string
	lastSync     string
	focusID      string
	caret        int
	anchor       int
	all          bool
	preeditID    string
	preedit      string
	preeditStart int
	preeditEnd   int
	live         map[string]Control
	order        []string
}

// formMerge carries one lock-free merge from the scan to the publish. body
// is the post-rewrite source stored as lastSyncBody; out is the published
// source (built marks it ready, so the skip path can leave it empty and let
// the emoji pass run after the publish unlocks); outHash is the hash of out
// (the input hash when the emoji pass leaves the source unchanged, otherwise
// a single hash of the final output), so the publish can note the source
// without rehashing; bound is the phase-2 bind snapshot the publish
// re-checks.
type formMerge struct {
	body    string
	hash    uint64
	spans   []controlSpan
	byID    map[string]Control
	order   []string
	st      form.Caret
	bound   map[string]Control
	out     string
	outHash uint64
	built   bool
}

// noteSourceHashed records a newly synced source with a caller-computed hash.
// It mirrors noteSource exactly (equal hash and length keep the revision, so
// an unchanged frame compares u64s with no memcmp; a change bumps the
// revision the key carries), letting the sync path thread one hash through
// the emoji pass into the source memo instead of hashing the same body in
// the emoji pass and in noteSource. Callers set syncClean after noting, so
// the draw adapter skips its own rehash.
func (p *Page) noteSourceHashed(next string, hash uint64) {
	if p.sourceRev != 0 && hash == p.sourceHash && len(next) == p.sourceLen {
		return
	}

	p.sourceHash = hash
	p.sourceLen = len(next)
	p.sourceRev++
}

func (p *Page) syncForm(body string) string {
	p.syncClean = false
	// Phase 1: snapshot the shared state under one short exclusive lock.
	// SetData publishes data and the dirty flag under this lock, so the
	// snapshot and the publish below cannot interleave with it. The live
	// control map and the caret scalars are owned by the page loop (no
	// other goroutine writes them), so retaining them lock-free is safe;
	// every SetData-shared field is copied here.
	p.dataMu.Lock()
	doc := p.nav.Doc()
	fs := p.formState()
	if fs == nil || fs.Doc != doc {
		fresh := form.NewDoc(doc)
		fs = &fresh
		p.setFormState(fs)
		p.formDirty = true
	}
	snap := formSnap{
		dirty:        p.formDirty,
		source:       p.source,
		lastSync:     p.lastSyncBody,
		focusID:      fs.FocusID,
		caret:        fs.Caret,
		anchor:       fs.Anchor,
		all:          fs.All,
		preeditID:    fs.PreeditID,
		preedit:      fs.Preedit,
		preeditStart: fs.PreeditStart,
		preeditEnd:   fs.PreeditEnd,
		live:         fs.ByID,
		order:        fs.Order,
	}
	p.dataMu.Unlock()

	// Fast path: the template output is unchanged and nothing is dirty. When
	// execTemplate reused the cached body, body IS lastSyncBody by
	// construction, so skip the memcmp: Go short-circuits ||, so the string
	// compare runs only when the frame is not already known clean. Reordering
	// !dirty before the compare also skips the memcmp on dirty frames, where
	// the fast path is impossible anyway.
	cleanExec := p.tplProbeHas && p.execSkipped
	if snap.source != "" && !snap.dirty && (cleanExec || body == snap.lastSync) {
		c, hasFocus := snap.live[snap.focusID]
		if !hasFocus || !form.CanEdit(c) {
			p.syncClean = true
			return snap.source
		}
	}

	// One fused pass: the control probe, the control scan, and the
	// button-input flag, memoized per source hash. A repeat body reuses
	// the cached spans and head offsets without walking the source. The hash
	// is computed once here and threaded through the rewrite, the emoji
	// pass, and the source memo, so the same body is never hashed twice.
	hash := syncHash(body)
	hasControls, spans, hasButtons := form.ScanCached(body, hash)

	// Pages without form controls skip the button rewrite, the merge, the
	// binds, and the control rewrite. The live set cached from the last
	// sync is empty and the new body carries no control tags, so every one
	// of those steps is a no-op; only the focus handoff, the bookkeeping,
	// and the emoji pass below can change anything.
	if len(snap.live) == 0 && len(snap.order) == 0 && !hasControls {
		return p.publishForm(formMerge{body: body, hash: hash}, snap)
	}

	// The fused scan flags exactly the inputs the button rewrite replaces
	// (same open-tag read, same type attribute read, same comment and
	// script skipping), so a false skips the rewrite knowing it is the
	// identity. A true rewrite moves spans, hence the rescan.
	if hasButtons {
		rewriteButtonCalls.Add(1)
		body = rewriteButtons(body)
		hash = syncHash(body)
		_, spans, _ = form.ScanCached(body, hash)
	}

	byID, order := form.Merge(snap.live, spans)

	// Phase 2: snapshot the bound fields under a short read lock. The scan
	// and the rewrites above and below run lock-free; bindRead still reads
	// p.data, so it stays under the lock.
	p.dataMu.RLock()
	bound := make(map[string]Control, len(byID))
	for id, c := range byID {
		if c.Bind == "" {
			continue
		}
		if next, ok := bindRead(p, c); ok {
			bound[id] = next
		}
	}
	p.dataMu.RUnlock()
	form.ApplyBinds(byID, func(c Control) (Control, bool) {
		next, ok := bound[c.ID]
		return next, ok
	})

	merge := formMerge{
		body:  body,
		hash:  hash,
		spans: spans,
		byID:  byID,
		order: order,
		st:    caretOfSnap(snap, byID),
		bound: bound,
	}

	// A span-free body (and a body with only plain buttons) needs no
	// control rewrite: RewriteCached returns the source unchanged for it.
	// The span count comes from the single scan above, not a rescan. The
	// rewrite output is hashed once for the emoji gate, so the emoji pass
	// never rehashes its input.
	if len(spans) > 0 && form.HasFieldSpans(spans) {
		rewritten := form.RewriteCached(body, hash, spans, byID, merge.st)
		rewriteHash := syncHash(rewritten)
		out, outHash := draw.EmojiPassWithHash(rewritten, rewriteHash)
		merge.out, merge.outHash = out, outHash
		merge.built = true
	}

	return p.publishForm(merge, snap)
}

// publishForm installs the merged controls under a short exclusive lock,
// re-checks the bound fields against a concurrent SetData, and stores the
// sync flags. The re-read is cheap (one reflected field per bound control,
// no source pass): when nothing moved, the lock-free merge publishes as is.
// When a publish landed after the phase-2 snapshot, the binds refresh and
// the output rebuilds under this lock from the cached spans and head
// offsets (no scan repeats), and the page stays dirty so the next redraw
// re-verifies against re-executed template output. The final source is noted
// with its threaded hash and syncClean is set, so the draw adapter skips its
// own rehash; the memo update mirrors noteSource exactly.
func (p *Page) publishForm(merge formMerge, snap formSnap) string {
	p.dataMu.Lock()
	fs := p.formState()
	dirty := false
	if merge.byID != nil {
		if refreshBindsLocked(p, &merge, snap) {
			dirty = true
		}
		fs.ByID = merge.byID
		fs.Order = merge.order
	}
	if _, ok := fs.ByID[fs.FocusID]; !ok {
		fs.FocusID = ""
		fs.ClearRange()
	}

	p.applyFocusRequest()
	fs.ClampToFocus()

	p.lastSyncBody = merge.body
	p.formDirty = dirty
	out, outHash, built := merge.out, merge.outHash, merge.built
	body, bodyHash := merge.body, merge.hash
	p.dataMu.Unlock()

	if !built {
		var oh uint64
		out, oh = draw.EmojiPassWithHash(body, bodyHash)
		outHash = oh
	}

	p.noteSourceHashed(out, outHash)
	p.syncClean = true

	return out
}

// refreshBindsLocked re-reads every bound control under the publish lock
// and reports whether any bound value moved under a concurrent publish. A
// move refreshes the merge and rebuilds the output; a control whose binding
// vanished (its data no longer binds) also counts as a move so the next
// redraw re-merges against the new data. bindRead keeps reading p.data, so
// it stays correct: it runs only under dataMu here and in phase 2.
func refreshBindsLocked(p *Page, merge *formMerge, snap formSnap) bool {
	moved := false
	for id, c := range merge.byID {
		if c.Bind == "" {
			continue
		}
		next, ok := bindRead(p, c)
		if !ok {
			if _, was := merge.bound[id]; was {
				moved = true
			}

			continue
		}
		if bindValueChanged(c, next) {
			merge.byID[id] = next
			moved = true
		}
	}
	if !moved {
		return false
	}

	merge.st = caretOfSnap(snap, merge.byID)
	merge.out = merge.body
	merge.outHash = merge.hash
	merge.built = false
	if len(merge.spans) > 0 && form.HasFieldSpans(merge.spans) {
		rewritten := form.RewriteCached(merge.body, merge.hash, merge.spans, merge.byID, merge.st)
		merge.out = rewritten
		merge.outHash = syncHash(rewritten)
		merge.built = true
	}
	out, outHash := draw.EmojiPassWithHash(merge.out, merge.outHash)
	merge.out, merge.outHash = out, outHash
	merge.built = true

	return true
}

// bindValueChanged reports whether a refreshed bind differs from the merge
// in exactly the fields bindRead sets: the check for a checkbox, the check
// for a radio, and the value for every other bound control.
func bindValueChanged(old, next Control) bool {
	switch bindKind(old) {
	case 'b':
		return old.Checked != next.Checked
	case 's':
		if old.Type == "radio" {
			return old.Checked != next.Checked
		}

		return old.Value != next.Value
	default:
		return false
	}
}

// caretOfSnap builds the rewrite caret from the merged controls and the
// snapshotted caret range, mirroring State.View after a merge.
func caretOfSnap(snap formSnap, byID map[string]Control) form.Caret {
	c, ok := byID[snap.focusID]

	return form.CaretOf(form.FocusView{
		FocusID:      snap.focusID,
		Caret:        snap.caret,
		Anchor:       snap.anchor,
		All:          snap.all,
		Control:      c,
		Present:      ok,
		PreeditID:    snap.preeditID,
		Preedit:      snap.preedit,
		PreeditStart: snap.preeditStart,
		PreeditEnd:   snap.preeditEnd,
	})
}

func (p *Page) clickControl(ctx context.Context, box Box) error {
	c, ok := p.control(box.ID)
	if ok && c.Disabled {
		return nil
	}

	if c.Type == "file" {
		p.markFormDirty()
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
		p.markFormDirty()
		p.markPair(oldFocus, p.currentFocus())
	}

	if changed {
		p.markFormDirty()
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
		p.markFormDirty()
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
		p.markFormDirty()
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
	p.formState().ByID[box.ID] = c
	bindWrite(p, c)
	p.markPending(box.ID)

	return p.change(ctx, box.ID)
}

func (p *Page) blurForm() {
	p.markFormDirty()
	p.tickMutated.Store(true)
	fs := p.formState()
	if fs == nil {
		return
	}

	old := fs.Blur()
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
	p.tickMutated.Store(true)
	fs := p.formState()
	if fs == nil {
		return false
	}

	return fs.Activate(id)
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
		// The handler can mutate the template data in place through the
		// SetData pointer, so force the next Redraw to re-execute.
		p.markFormDirty()
	}

	changed, step, err := p.applyEdit(ctx, edit)
	if err != nil {
		return err
	}

	if !changed && step == 0 && !ran {
		return nil
	}

	edited := ""
	if fs := p.formState(); fs != nil {
		edited = fs.FocusID
	}

	if changed {
		p.markPending(edited)
		if err := p.change(ctx, edited); err != nil {
			return err
		}
	}

	if step != 0 {
		p.focusSlotStep(edited, step)
	}

	return p.Redraw(ctx)
}

func (p *Page) applyEdit(ctx context.Context, edit valueEdit) (bool, int, error) {
	c, ok := p.typingTarget()
	if !ok {
		return false, 0, nil
	}

	fs := p.formState()
	next, caret, anchor, changed := edit(c, fs.Caret, fs.Anchor)
	step := next.SlotStep
	next.SlotStep = 0
	if !changed && step == 0 {
		return false, 0, nil
	}

	if !changed {
		return false, step, nil
	}

	// GEN-28: every edit funnels here, so clamp once for all writers.
	next, caret, anchor = form.ClampEdited(next, caret, anchor)
	if err := p.beforeEdit(ctx, fs.FocusID); err != nil {
		return false, 0, err
	}

	fs = p.formState()
	fs.ByID[fs.FocusID] = next
	fs.Caret, fs.Anchor = caret, anchor
	fs.All = false
	p.resetCaretBlink()
	bindWrite(p, next)

	return true, step, nil
}

// focusSlotStep moves focus from a one-digit slot by step. A missing
// neighbor leaves the focus where it is.
func (p *Page) focusSlotStep(from string, step int) {
	fs := p.formState()
	if fs == nil || step == 0 {
		return
	}

	id := form.SlotNeighbor(fs.ByID, from, step)
	if id == "" || !p.inFocusScope(id) {
		return
	}

	c, ok := fs.Lookup(id)
	moved, reveal := fs.TakeFocus(id, c, ok)
	if !moved {
		return
	}

	p.tickMutated.Store(true)
	p.markPair(from, id)
	if reveal {
		p.revealFocus(id)
	}
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

	c, ok := p.control(fs.FocusID)
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

	if err := p.handlers.BeforeEdit(ctx, p.boxByID(id)); err != nil {
		return err
	}
	p.markFormDirty()

	return nil
}

// change calls the Change handler for id.
// A nil handler does nothing.
func (p *Page) change(ctx context.Context, id string) error {
	if p.handlers.Change == nil {
		return nil
	}

	if err := p.handlers.Change(ctx, p.boxByID(id)); err != nil {
		return err
	}
	p.markFormDirty()

	return nil
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
	p.tickMutated.Store(true)
	fs := p.formState()
	if fs == nil {
		return
	}

	fs.WriteValue(id, value)
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
	p.tickMutated.Store(true)
	fs := p.formState()
	if fs == nil {
		return
	}

	fs.WriteChecked(id, checked)
}

// FocusedField returns the focused field id, excluding buttons.
// FocusID also reports button focus.
func (p *Page) FocusedField() string {
	fs := p.formState()
	if fs == nil {
		return ""
	}

	return fs.FieldID()
}

func (p *Page) control(id string) (Control, bool) {
	fs := p.formState()
	if fs == nil {
		return Control{}, false
	}

	return fs.Lookup(id)
}

// FormSelected reports whether id is the focused control and its whole
// value is selected. It is false when no control is focused, when the
// page has no form yet, or when another control is focused.
func (p *Page) FormSelected(id string) bool {
	fs := p.formState()
	if fs == nil || fs.FocusID != id {
		return false
	}

	return form.WholeSelected(fs.View(), id)
}
