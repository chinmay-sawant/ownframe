package page

import (
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
	"html"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"time"
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
func insertValue(p *Page, c Control, caret, anchor int, text string) (Control, int, int, bool) {
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
	return focusStep(p, ctx, 1)
}

// FocusPrev focuses the previous control in document order and wraps.
// Disabled controls and tabindex="-1" are skipped. It draws when the focus
// moved.
func (p *Page) FocusPrev(ctx context.Context) error {
	return focusStep(p, ctx, -1)
}

// Focus focuses the control id and draws. An empty id blurs. A disabled
// control or an unknown id is left alone.
func (p *Page) Focus(ctx context.Context, id string) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if id == "" {
		fs := formState(p)
		if fs == nil || fs.FocusID == "" {
			return nil
		}

		blurForm(p)

		return p.Redraw(ctx)
	}

	c, ok := control(p, id)
	if !ok || c.Disabled || !inFocusScope(p, id) {
		return nil
	}

	moved, reveal := formState(p).TakeFocus(id, c, ok)
	if !moved {
		return nil
	}

	p.tickMutated.Store(true)
	if reveal {
		revealFocus(p, id)
	}

	return p.Redraw(ctx)
}

// FocusID returns the focused control id, or "" when none is focused.
func (p *Page) FocusID() string {
	fs := formState(p)
	if fs == nil {
		return ""
	}

	return fs.FocusID
}

// LastFocusID returns the control id cleared by the most recent blur.
func (p *Page) LastFocusID() string {
	fs := formState(p)
	if fs == nil {
		return ""
	}

	return fs.LastFocusID
}

func focusStep(p *Page, ctx context.Context, step int) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id, ok := form.NextID(focusOrder(p), p.FocusID(), step)
	if !ok {
		return nil
	}

	return p.Focus(ctx, id)
}

// focusOrder is the tab order: the document order of the control boxes,
// with disabled controls and tabindex="-1" left out.
func focusOrder(p *Page) []string {
	ids := make([]string, 0, len(p.boxes))
	for _, b := range p.boxes {
		if inFocusScope(p, b.ID) {
			ids = append(ids, b.ID)
		}
	}

	return form.Order(ids, func(id string) (Control, bool) { return control(p, id) })
}

// SelectAt focuses the control under the point and puts the caret at the
// measured rune offset. A point outside every control blurs the form. A
// disabled control is left alone. A control that takes no text just takes
// focus.
func (p *Page) SelectAt(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id := boxIDAt(p, x, y)
	c, ok := control(p, id)
	if !ok {
		fs := formState(p)
		if fs == nil || fs.FocusID == "" {
			return nil
		}

		blurForm(p)

		return p.Redraw(ctx)
	}

	if c.Disabled {
		return nil
	}

	if !canEdit(c) {
		return p.Focus(ctx, id)
	}

	return focusCaret(p, ctx, c, offsetAt(p, boxByID(p, id), c, x, y))
}

// Drag extends the focused field's range to the point. It does nothing when
// no text field is focused.
func (p *Page) Drag(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	c, ok := typingTarget(p)
	if !ok {
		return nil
	}

	off := offsetAt(p, boxByID(p, c.ID), c, x, y)
	changed := formState(p).DragTo(off)
	if !changed {
		return nil
	}

	p.tickMutated.Store(true)
	return p.Redraw(ctx)
}

// SelectWordAt selects the word under the point and draws.
func (p *Page) SelectWordAt(ctx context.Context, x, y float64) error {
	return selectSpan(p, ctx, x, y, form.WordBounds)
}

// SelectLineAt selects the line under the point and draws.
func (p *Page) SelectLineAt(ctx context.Context, x, y float64) error {
	return selectSpan(p, ctx, x, y, form.LineBounds)
}

func selectSpan(p *Page, ctx context.Context, x, y float64, bounds func(string, int) (int, int)) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id := boxIDAt(p, x, y)
	c, ok := control(p, id)
	if !ok || c.Disabled || !canEdit(c) {
		return nil
	}

	off := offsetAt(p, boxByID(p, id), c, x, y)
	start, end := bounds(c.Value, off)
	formState(p).SelectSpan(c.ID, start, end)
	p.tickMutated.Store(true)

	return p.Redraw(ctx)
}

// offsetAt turns a point into a rune offset in c's value. A run that cannot
// be measured, or one whose drawn text does not line up with the shown
// value, lands at the end of the value.
func offsetAt(p *Page, box Box, c Control, x, y float64) int {
	return form.OffsetAt(p.draw.Display, box, c, x, y)
}

// focusCaret focuses c and puts the caret and the anchor at off.
func focusCaret(p *Page, ctx context.Context, c Control, off int) error {
	changed := formState(p).FocusCaret(c.ID, off)
	if !changed {
		return nil
	}

	p.tickMutated.Store(true)
	return p.Redraw(ctx)
}

func revealFocus(p *Page, id string) {
	box := boxByID(p, id)
	if isFixed(p, box) {
		// A box the window pins to the viewport is always visible, so
		// revealing it must not scroll the page away from the content.
		return
	}

	x, y := p.ScrollOffset()
	x, y = form.RevealOrigin(box.X, box.Y, box.W, box.H, x, y, p.width, p.height)
	p.ScrollTo(x, y)
}

// setCaret moves the caret to pos, keeping the anchor when extend is true.
func setCaret(p *Page, pos int, extend bool) {
	if formState(p) == nil {
		return
	}

	p.tickMutated.Store(true)
	n := 0
	if c, ok := focusedEditable(p); ok {
		n = runeLen(c.Value)
	}

	formState(p).SetCaret(pos, n, extend)
}

// hasSelection reports whether the focused field has a non-empty range.
func hasSelection(p *Page) bool {
	c, ok := typingTarget(p)
	if !ok {
		return false
	}

	start, end := formState(p).Bounds(runeLen(c.Value))

	return start != end
}

// caretKey moves the caret for a focused text field and draws. Arrow keys
// move by rune, Home and End move to the ends, and a ctrl+ or alt+ prefix
// jumps by word. A shift+ prefix extends the range from the anchor.
func caretKey(p *Page, ctx context.Context, key string) error {
	c, ok := typingTarget(p)
	if !ok {
		return nil
	}

	fs := formState(p)
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
func resetCaretBlink(p *Page) {
	p.tickMutated.Store(true)
	fs := formState(p)
	if fs == nil {
		return
	}

	fs.RestartBlink()
}

// caretBlink toggles the focused field's caret on the blink clock and draws
// when the phase changes. A range selection paints no caret, so it does
// nothing. A file input never shows a caret, so it does nothing.
func caretBlink(p *Page, ctx context.Context) error {
	c, ok := focusedEditable(p)
	fs := formState(p)
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

func paintCaret(p *Page, d *layout.Display) {
	fs := formState(p)
	if fs == nil {
		return
	}

	c, ok := focusedEditable(p)
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
func markFormDirty(p *Page) {
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

// sameSourceBytes reports whether a and b are byte-identical without
// scanning either: an equal non-zero length and one backing array make them
// the same bytes. reflect exposes a string's data pointer without an unsafe
// import, which the lint policy keeps to the reviewed syscall shims.
func sameSourceBytes(a, b string) bool {
	if len(a) == 0 || len(a) != len(b) {
		return false
	}

	return reflect.ValueOf(a).UnsafePointer() == reflect.ValueOf(b).UnsafePointer()
}

// formSnap pins the sync inputs snapshot under the phase-1 lock. The dirty
// flag, the last synced body, and the published source are shared with
// SetData; the live map, the order, and the caret scalars are owned by the
// page loop and only retained, never mutated, after the unlock.
type formSnap struct {
	dirty        bool
	source       string
	lastSync     string
	lastBody     hashedBody
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
// Equal hash and length keep the revision, so an unchanged frame compares u64s
// with no memcmp; a change bumps the revision the key carries. It is the store
// behind noteSource, and it lets the sync path and the adopt path thread one
// hash through instead of hashing the same body again. Callers set syncClean
// after noting, so the draw adapter skips its own rehash.
func noteSourceHashed(p *Page, next string, hash uint64) {
	if p.sourceRev != 0 && hash == p.sourceHash && len(next) == p.sourceLen {
		return
	}

	p.sourceHash = hash
	p.sourceLen = len(next)
	p.sourceRev++
}

func syncForm(p *Page, body string) string {
	p.syncClean = false
	// Phase 1: snapshot the shared state under one short exclusive lock.
	// SetData publishes data and the dirty flag under this lock, so the
	// snapshot and the publish below cannot interleave with it. The live
	// control map and the caret scalars are owned by the page loop (no
	// other goroutine writes them), so retaining them lock-free is safe;
	// every SetData-shared field is copied here.
	p.dataMu.Lock()
	doc := p.nav.Doc()
	fs := formState(p)
	if fs == nil || fs.Doc != doc {
		fresh := form.NewDoc(doc)
		fs = &fresh
		setFormState(p, fs)
		p.formDirty = true
	}
	snap := formSnap{
		dirty:        p.formDirty,
		source:       p.source,
		lastSync:     p.lastSyncBody,
		lastBody:     p.lastBody,
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
	cleanExec := p.execSkipped && p.nav.Current().Execute
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
	// is threaded through the rewrite, the emoji pass, and the source memo,
	// so the same body is never hashed twice. A body that still shares the
	// memoized body's backing array is byte-identical by construction, so it
	// reuses that hash instead of folding the bytes again on every clean
	// focused redraw.
	hash := snap.lastBody.hash
	if !sameSourceBytes(body, snap.lastBody.source) {
		hash = syncHash(body)
	}
	hasControls, spans, hasButtons := form.ScanCached(body, hash)

	// Pages without form controls skip the button rewrite, the merge, the
	// binds, and the control rewrite. The live set cached from the last
	// sync is empty and the new body carries no control tags, so every one
	// of those steps is a no-op; only the focus handoff, the bookkeeping,
	// and the emoji pass below can change anything.
	if len(snap.live) == 0 && len(snap.order) == 0 && !hasControls {
		return publishForm(p, formMerge{body: body, hash: hash}, snap)
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

	return publishForm(p, merge, snap)
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
func publishForm(p *Page, merge formMerge, snap formSnap) string {
	p.dataMu.Lock()
	fs := formState(p)
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

	applyFocusRequest(p)
	fs.ClampToFocus()

	p.lastSyncBody = merge.body
	p.lastBody = hashedBody{source: merge.body, hash: merge.hash}
	p.formDirty = dirty
	out, outHash, built := merge.out, merge.outHash, merge.built
	body, bodyHash := merge.body, merge.hash
	p.dataMu.Unlock()

	if !built {
		var oh uint64
		out, oh = draw.EmojiPassWithHash(body, bodyHash)
		outHash = oh
	}

	noteSourceHashed(p, out, outHash)
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

func clickControl(p *Page, ctx context.Context, box Box) error {
	c, ok := control(p, box.ID)
	if ok && c.Disabled {
		return nil
	}

	if c.Type == "file" {
		markFormDirty(p)
		return clickFileControl(p, ctx, box, c)
	}

	if !c.Disabled && form.WillChange(c) {
		if err := beforeEdit(p, ctx, box.ID); err != nil {
			return err
		}
	}

	oldFocus := currentFocus(p)
	changed := false
	if !c.Disabled {
		changed = activate(p, box.ID)
	}

	if currentFocus(p) != oldFocus {
		markFormDirty(p)
		markPair(p, oldFocus, currentFocus(p))
	}

	if changed {
		markFormDirty(p)
		if c.Type == "checkbox" || c.Type == "radio" {
			markToggle(p, box.ID)
		}

		if next, ok := control(p, box.ID); ok {
			bindWrite(p, next)
		}

		if err := change(p, ctx, box.ID); err != nil {
			return err
		}
	}

	if p.handlers.Click != nil {
		if err := p.handlers.Click(ctx, box); err != nil {
			return err
		}
		markFormDirty(p)
	}

	painted, err := clickPainted(p, ctx, box)
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
func clickFileControl(p *Page, ctx context.Context, box Box, c Control) error {
	oldFocus := currentFocus(p)
	activate(p, box.ID)
	if currentFocus(p) != oldFocus {
		markPair(p, oldFocus, currentFocus(p))
	}

	if p.picker != nil {
		if err := applyPicked(p, ctx, box, c); err != nil {
			return err
		}
	}

	if p.handlers.Click != nil {
		if err := p.handlers.Click(ctx, box); err != nil {
			return err
		}
		markFormDirty(p)
	}

	return p.Redraw(ctx)
}

func applyPicked(p *Page, ctx context.Context, box Box, c Control) error {
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

	if err := beforeEdit(p, ctx, box.ID); err != nil {
		return err
	}

	c.Value = path
	formState(p).ByID[box.ID] = c
	bindWrite(p, c)
	markPending(p, box.ID)

	return change(p, ctx, box.ID)
}

func blurForm(p *Page) {
	markFormDirty(p)
	p.tickMutated.Store(true)
	fs := formState(p)
	if fs == nil {
		return
	}

	old := fs.Blur()
	markPair(p, old, "")
}

// currentFocus is the focused control id, or "" when none is focused.
func currentFocus(p *Page) string {
	return p.FocusID()
}

func formControl(p *Page, id string) bool {
	if id == "" {
		return false
	}

	_, ok := control(p, id)

	return ok
}

func activate(p *Page, id string) bool {
	p.tickMutated.Store(true)
	fs := formState(p)
	if fs == nil {
		return false
	}

	return fs.Activate(id)
}

func editField(p *Page, ctx context.Context, fn func() error, edit valueEdit) error {
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
		markFormDirty(p)
	}

	changed, step, err := applyEdit(p, ctx, edit)
	if err != nil {
		return err
	}

	if !changed && step == 0 && !ran {
		return nil
	}

	edited := ""
	if fs := formState(p); fs != nil {
		edited = fs.FocusID
	}

	if changed {
		markPending(p, edited)
		if err := change(p, ctx, edited); err != nil {
			return err
		}
	}

	if step != 0 {
		focusSlotStep(p, edited, step)
	}

	return p.Redraw(ctx)
}

func applyEdit(p *Page, ctx context.Context, edit valueEdit) (bool, int, error) {
	c, ok := typingTarget(p)
	if !ok {
		return false, 0, nil
	}

	fs := formState(p)
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
	if err := beforeEdit(p, ctx, fs.FocusID); err != nil {
		return false, 0, err
	}

	fs = formState(p)
	fs.ByID[fs.FocusID] = next
	fs.Caret, fs.Anchor = caret, anchor
	fs.All = false
	resetCaretBlink(p)
	bindWrite(p, next)

	return true, step, nil
}

// focusSlotStep moves focus from a one-digit slot by step. A missing
// neighbor leaves the focus where it is.
func focusSlotStep(p *Page, from string, step int) {
	fs := formState(p)
	if fs == nil || step == 0 {
		return
	}

	id := form.SlotNeighbor(fs.ByID, from, step)
	if id == "" || !inFocusScope(p, id) {
		return
	}

	c, ok := fs.Lookup(id)
	moved, reveal := fs.TakeFocus(id, c, ok)
	if !moved {
		return
	}

	p.tickMutated.Store(true)
	markPair(p, from, id)
	if reveal {
		revealFocus(p, id)
	}
}

func typingTarget(p *Page) (Control, bool) {
	c, ok := focusedEditable(p)
	if !ok || c.Disabled {
		return Control{}, false
	}

	return c, true
}

func focusedEditable(p *Page) (Control, bool) {
	fs := formState(p)
	if fs == nil {
		return Control{}, false
	}

	c, ok := control(p, fs.FocusID)
	if !ok || !canEdit(c) {
		return Control{}, false
	}

	return c, true
}

// beforeEdit calls the BeforeEdit handler for id.
// A nil handler does nothing.
func beforeEdit(p *Page, ctx context.Context, id string) error {
	if p.handlers.BeforeEdit == nil {
		return nil
	}

	if err := p.handlers.BeforeEdit(ctx, boxByID(p, id)); err != nil {
		return err
	}
	markFormDirty(p)

	return nil
}

// change calls the Change handler for id.
// A nil handler does nothing.
func change(p *Page, ctx context.Context, id string) error {
	if p.handlers.Change == nil {
		return nil
	}

	if err := p.handlers.Change(ctx, boxByID(p, id)); err != nil {
		return err
	}
	markFormDirty(p)

	return nil
}

// FormValue returns the current value of id, or "" when id is absent.
func (p *Page) FormValue(id string) string {
	c, _ := control(p, id)

	return c.Value
}

// SetFormValue stores value for id and does not draw.
// It does nothing when that control is not on the page yet.
// A select marks each option whose value matches.
// The focused control loses its selection.
func (p *Page) SetFormValue(id, value string) {
	p.tickMutated.Store(true)
	fs := formState(p)
	if fs == nil {
		return
	}

	fs.WriteValue(id, value)
}

// FormChecked reports whether id is checked.
// A missing control returns false.
func (p *Page) FormChecked(id string) bool {
	c, _ := control(p, id)

	return c.Checked
}

// SetFormChecked sets a checkbox or radio and does not draw.
// It does nothing when id is missing.
// A radio set true clears other radios with the same non-empty name.
// A radio set false clears only that radio.
func (p *Page) SetFormChecked(id string, checked bool) {
	p.tickMutated.Store(true)
	fs := formState(p)
	if fs == nil {
		return
	}

	fs.WriteChecked(id, checked)
}

// FocusedField returns the focused field id, excluding buttons.
// FocusID also reports button focus.
func (p *Page) FocusedField() string {
	fs := formState(p)
	if fs == nil {
		return ""
	}

	return fs.FieldID()
}

func control(p *Page, id string) (Control, bool) {
	fs := formState(p)
	if fs == nil {
		return Control{}, false
	}

	return fs.Lookup(id)
}

// FormSelected reports whether id is the focused control and its whole
// value is selected. It is false when no control is focused, when the
// page has no form yet, or when another control is focused.
func (p *Page) FormSelected(id string) bool {
	fs := formState(p)
	if fs == nil || fs.FocusID != id {
		return false
	}

	return form.WholeSelected(fs.View(), id)
}

// SetFocusScope restricts keyboard traversal to ids; an empty list clears it.
func SetFocusScope(p *Page, ids []string) {
	p.ff.scope = slices.Clone(ids)
}

// RequestFocus queues focus until the next form sync, when new controls exist.
// Empty id blurs. An absent or disabled control clears the old focus.
func RequestFocus(p *Page, id string) {
	p.ff.requestFocus, p.ff.hasFocusRequest = id, true
	markFormDirty(p)
}

func inFocusScope(p *Page, id string) bool {
	return len(p.ff.scope) == 0 || slices.Contains(p.ff.scope, id)
}

func applyFocusRequest(p *Page) {
	if !p.ff.hasFocusRequest {
		return
	}
	p.ff.hasFocusRequest = false
	fs := formState(p)
	c, ok := fs.ByID[p.ff.requestFocus]
	if !ok || c.Disabled || !inFocusScope(p, c.ID) {
		fs.FocusID = ""
		fs.ClearRange()
		return
	}
	fs.TakeFocus(c.ID, c, true)
}

// formFocusState owns the form document, the pointer hit state, and the
// accessibility snapshot. Only this file and page_form_focus_hit.go write
// its fields; Page delegates below. No lock: the page loop serializes it.
type formFocusState struct {
	scope           []string
	requestFocus    string
	hasFocusRequest bool
	form            *form.State
	hover           string
	hoverX, hoverY  float64
	active          string
	pressUsed       bool
	accessibility   *accessibilityState
}

// formState returns the owned form document, or nil before first sync.
func formState(p *Page) *form.State { return p.ff.form }

// setFormState replaces the owned form document.
func setFormState(p *Page, s *form.State) { p.ff.form = s }

// HoverID returns the element the page paints as hovered, or "" when none.
// The window reads it through host.Hoverer so a hover the page set itself
// survives a frame that did not move the pointer.
func (p *Page) HoverID() string { return p.ff.hover }

// setHoverID records the hovered element id.
func setHoverID(p *Page, id string) { p.ff.hover = id }

// clearHoverID drops the hovered element id.
func clearHoverID(p *Page) { p.ff.hover = "" }

// hoverPoint is the last hover position in CSS pixels.
func hoverPoint(p *Page) (float64, float64) { return p.ff.hoverX, p.ff.hoverY }

// setHoverPoint records the last hover position in CSS pixels.
func setHoverPoint(p *Page, x, y float64) { p.ff.hoverX, p.ff.hoverY = x, y }

// activeID is the pressed element id, or "" when none.
func activeID(p *Page) string { return p.ff.active }

// setActiveID records the pressed element id.
func setActiveID(p *Page, id string) { p.ff.active = id }

// setPressUsed records whether a handler consumed the current press.
func setPressUsed(p *Page, used bool) { p.ff.pressUsed = used }

// takePress clears the press hit and returns the held id, or "".
func takePress(p *Page) string {
	id := p.ff.active
	p.ff.active = ""
	p.ff.pressUsed = false
	return id
}

// consumePress drops a handler-consumed press hit. It is true when Click
// lands on the pressed id and should only redraw.
func consumePress(p *Page, id string) bool {
	if p.ff.pressUsed && p.ff.active == id {
		p.ff.pressUsed = false
		return true
	}
	return false
}

// a11yState returns the owned accessibility snapshot, or nil.
func a11yState(p *Page) *accessibilityState { return p.ff.accessibility }

// setA11yState replaces the owned accessibility snapshot.
func setA11yState(p *Page, s *accessibilityState) { p.ff.accessibility = s }

// rewriteButtons replaces a submit, button, or reset input with a button
// element so the engine gives it a face and a hit box.
func rewriteButtons(source string) string {
	if !strings.Contains(source, "<input") && !strings.Contains(source, "<INPUT") {
		return source
	}
	var b strings.Builder
	i, prev := 0, 0
	for i < len(source) {
		next := strings.IndexByte(source[i:], '<')
		if next < 0 {
			break
		}
		i += next
		if strings.HasPrefix(source[i:], "<!--") {
			end := strings.Index(source[i+commentOpenLen:], "-->")
			if end < 0 {
				break
			}
			i += end + commentTotalLen
			continue
		}
		name, raw, end, ok := readOpen(source, i)
		if !ok {
			i++
			continue
		}
		low := strings.ToLower(name)
		if low == "script" || low == "style" {
			_, i = closeSpan(source, end, low)
			continue
		}
		kind := ""
		if low == "input" {
			kind = strings.ToLower(attrValue(raw, "type"))
		}
		if kind != "submit" && kind != "button" && kind != "reset" {
			i = end
			continue
		}
		b.WriteString(source[prev:i])
		b.WriteString(buttonFromInput(raw, kind))
		prev = end
		i = end
	}
	if prev == 0 {
		return source
	}
	b.WriteString(source[prev:])

	return b.String()
}

func buttonFromInput(raw, kind string) string {
	label := attrValue(raw, "value")
	switch {
	case label != "":
	case kind == "submit":
		label = "Submit"
	case kind == "reset":
		label = "Reset"
	}

	return openTagDrop("button", raw, nil, buttonDrop) + html.EscapeString(label) + "</button>"
}

func buttonDrop(name string) bool {
	return name == "type" || name == "value"
}

// A button gets its face from the layout engine's default stylesheet, so the
// page package no longer scans the source for button elements.
