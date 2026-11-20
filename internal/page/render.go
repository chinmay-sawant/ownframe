package page

import (
	"context"
	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"html/template"
	"io"
	"runtime"
	"strings"
	"time"
)

func (p *Page) redraw(ctx context.Context, skipSame bool) error {
	return draw.Run(ctx, p.surf, skipSame)
}

// redrawSurface adapts Page to draw.RedrawSurface.
// Each method delegates to the page step the sequence already used.
type redrawSurface struct{ p *Page }

func (s redrawSurface) AfterBusy() bool { return s.p.afterRedraw.Busy() }

func (s redrawSurface) AfterAsk() { s.p.afterRedraw.Ask() }

func (s redrawSurface) QueueBatch(skipSame bool) bool {
	return draw.QueueBatch(
		s.p.redrawBatch,
		&s.p.redrawPending,
		&s.p.redrawForce,
		skipSame,
	)
}

func (s redrawSurface) BeginSample() (draw.Sample, bool) {
	s.p.stats.lastTemplate = 0
	s.p.stats.styleTime = 0
	s.p.stats.layoutTime = 0
	s.p.stats.paintTime = 0
	s.p.stats.otherTime = 0
	if !s.p.perf {
		return draw.BeginSample(false, false)
	}

	return draw.BeginSample(true, s.p.allocDue(time.Now()))
}

func (s redrawSurface) ClearPerf() { s.p.clearPerf() }

func (s redrawSurface) CountRedraw() { s.p.stats.redraws++ }

func (s redrawSurface) FinishRedraw(track, memOK bool, mem runtime.MemStats, start time.Time) {
	s.p.finishRedraw(track, memOK, mem, start)
}

func (s redrawSurface) ExecTemplate(track bool) (string, error) {
	return s.p.execTemplate(track)
}

func (s redrawSurface) SyncForm(executed string) string {
	out := s.p.syncForm(executed)
	s.p.source = out
	// A clean fast-path sync leaves the source byte-identical, so the
	// memoized hash, length, and revision from the last real sync are
	// still valid and the rehash is pure overhead on the skip path.
	if !s.p.syncClean {
		s.p.noteSource(out)
	}

	return out
}

func (s redrawSurface) RenderState() render.State { return s.p.renderState() }

func (s redrawSurface) Unchanged(state render.State) bool { return s.p.unchanged(state) }

func (s redrawSurface) DropPNG() { s.p.draw.DropPNG() }

func (s redrawSurface) StyleStaged(
	ctx context.Context,
	track bool,
	source string,
	state render.State,
) (*css.Document, error) {
	return s.p.styleStaged(ctx, track, source, state)
}

func (s redrawSurface) DrawStaged(
	ctx context.Context,
	track bool,
	styled *css.Document,
	state render.State,
	executed string,
) error {
	return s.p.drawStaged(ctx, track, styled, state, executed)
}

func (s redrawSurface) AfterRedrawDraw(ctx context.Context) error {
	return s.p.afterRedrawDraw(ctx)
}

var _ draw.RedrawSurface = redrawSurface{}

// execTemplate runs the trusted template, timing it when tracked.
//
// The "{{" probe is memoized to template-write time: the scan runs once
// when p.tpl changes and the verdict is reused, so repeat frames pay no
// scan. A clean form reuses the last merged body with no Execute. A dirty
// frame executes into a pooled buffer and keeps one copy; its hash is taken
// once in the sync path, which pairs it with the body for the next reuse.
func (p *Page) execTemplate(track bool) (string, error) {
	p.applyScrollWindow()
	if p.tplProbeTpl != p.tpl {
		p.tplProbeHas = strings.Contains(p.nav.Current(), "{{")
		p.tplProbeTpl = p.tpl
	}

	if !p.tplProbeHas {
		p.stats.lastTemplate = 0
		p.execSkipped = true

		return p.nav.Current(), nil
	}

	p.dataMu.RLock()
	defer p.dataMu.RUnlock()

	// A Redraw with no new data, form input, or changed row window reuses
	// the last merged template body. Every path that runs application code
	// marks the form dirty first (handlers can mutate the data in place
	// through the SetData pointer), so a clean flag means the body is
	// still current.
	if !p.formDirty && p.lastSyncBody != "" {
		p.stats.lastTemplate = 0
		p.execSkipped = true

		return p.lastSyncBody, nil
	}

	p.execSkipped = false

	out, err := draw.ExecuteWriter(
		track,
		len(p.source),
		&p.stats.lastTemplate,
		func(w io.Writer) error {
			return p.tpl.Execute(w, p.data)
		},
	)
	if err != nil {
		return "", err
	}

	body := string(out)

	return body, nil
}

// styleStaged styles the source, timing it when tracked.
//
// A theme swap forces a fresh Apply: css.Relayout keeps the cached
// document's extra sheets, so the cacheFresh relayout would keep painting
// the old theme. A state-only move the sheets cannot match reuses the cached
// document with no relayout at all.
//
// Otherwise, when the style cache already reflects the current source memo,
// the page relayouts directly: the revision, hash, and length compare with no
// cache lock, and the relayout takes one. Only a new source (or a dropped
// cache) pays the Source memcmp and the fresh parse inside StyleDocument.
func (p *Page) styleStaged(
	ctx context.Context,
	track bool,
	source string,
	state render.State,
) (*css.Document, error) {
	if doc, ok := p.styleNoop(state); ok {
		return doc, nil
	}

	changed := p.themeChanged(state)
	if !changed && p.cacheFresh() {
		// An element patch the hover handler just committed already
		// relayed the cache for this state, so reuse its document.
		if doc, ok := p.relayedDoc(state); ok {
			return doc, nil
		}
		ds := draw.Clock(track)

		doc, err := p.cache.Relayout(ctx, p.width, p.height, state)
		draw.Note(track, &p.stats.styleTime, ds)
		if err != nil {
			return nil, err
		}

		p.noteRelayed(state)
		return doc, nil
	}

	styled, err := draw.StyleDocument(
		ctx,
		track,
		&p.stats.styleTime,
		styleCache(p.cache, changed),
		source,
		p.viewport(),
		state,
	)
	if err != nil {
		return nil, err
	}

	p.cache = styled.Cache
	if styled.Fresh {
		p.statePseudos = styled.Pseudos
		p.stats.parses++
		p.stats.cascades++
	} else if changed {
		// A relayout keeps the old extra sheet, so refresh what the new
		// document can match even without a fresh cascade.
		p.statePseudos = p.cache.StatePseudos()
	}

	p.noteCacheSource()
	p.noteRelayed(state)

	return styled.Doc, nil
}

// styleCache selects the cache a style call starts from. A theme change
// passes nil so the call takes the fresh Apply: css.Relayout keeps the
// cached document's extra sheets and would keep the old theme.
func styleCache(cache *render.Cache, themeChanged bool) *render.Cache {
	if themeChanged {
		return nil
	}

	return cache
}

// themeChanged reports that the theme sheet differs from the one the last
// committed frame was styled with. The cache still reflects the old sheet,
// so the caller must not take the cacheFresh relayout for it.
func (p *Page) themeChanged(state render.State) bool {
	return p.cache != nil && p.last.Theme != state.Theme
}

// styleNoop returns the cached styled document when a redraw changed only
// pointer or focus state the sheets cannot match, so a relayout would
// reproduce the same styles. ok is false when the frame changed anything
// else, or when a changed state can match, and the caller restyles.
func (p *Page) styleNoop(state render.State) (doc *css.Document, ok bool) {
	if p.cache == nil || !p.hasLast {
		return nil, false
	}

	prev := render.State{Hover: p.last.Hover, Active: p.last.Active, Focus: p.last.Focus}
	if !state.SameStyleState(prev, p.statePseudos) {
		return nil, false
	}

	next, last := p.keyFor(state), p.last
	next.Focus, next.Hover, next.Active = "", "", ""
	last.Focus, last.Hover, last.Active = "", "", ""

	if next != last {
		return nil, false
	}

	doc = p.cache.Styled()
	if doc == nil {
		return nil, false
	}

	return doc, true
}

// drawStaged builds the list and commits either path.
func (p *Page) drawStaged(
	ctx context.Context,
	track bool,
	styled *css.Document,
	state render.State,
	executed string,
) error {
	drawStart := draw.Clock(true)
	ds := draw.Clock(track)
	display, err := p.layoutStaged(ctx, styled, state)
	draw.Note(track, &p.stats.layoutTime, ds)
	if err != nil {
		return err
	}

	p.paintCaret(display)
	if draw.Choose(display) == draw.PathList {
		p.commitDisplay(display, executed, drawStart, state)

		return nil
	}
	if err := p.paintFallback(display, drawStart, track, state); err != nil {
		return err
	}

	p.commitAccessibility(executed)

	return nil
}

// commitDisplay installs a replayable list and its dirty region.
func (p *Page) commitDisplay(display *Display, executed string, drawStart time.Time, state render.State) {
	previous := p.draw.Display
	previousFallback := p.draw.FallbackDisplay
	p.updateMotionIndex(previous, display)
	captures := p.captureTransitions(previous, display)
	p.dirtyFromDisplay(previous, display)
	p.noteCommit(drawStart)
	p.draw.CommitList(display)
	p.setDisplayBoxes(display)
	p.generation++
	p.layerGen++
	p.bindMotion(previous, display)
	p.bindTransitions(captures)
	if previous != nil && previous != display {
		previous.Release()
	}
	if previousFallback != nil && previousFallback != display && previousFallback != previous {
		previousFallback.Release()
	}
	p.remember(state)
	p.commitAccessibility(executed)
	p.applyPending()
	p.runAfterRedraw()
}

func (p *Page) noteCommit(drawStart time.Time) {
	draw.NoteCommit(&p.stats.layouts, &p.stats.repaints, &p.stats.lastDraw, drawStart)
}

func (p *Page) finishRedraw(track, memOK bool, m0 runtime.MemStats, start time.Time) {
	defer p.notifyFrame()
	p.stats.lastRedraw = time.Since(start)
	if track {
		p.stats.otherTime = max(0, p.stats.lastRedraw-p.stats.lastTemplate-
			p.stats.styleTime-p.stats.layoutTime-p.stats.paintTime)
	}
	if track && memOK {
		p.stats.allocFrame = draw.AllocUsed(m0)
	}
}

// paintFallback stores a bitmap for a page the vector path declined. The
// picture is a replay of that same list. blinkless does not paint a page
// bitmap of its own. The hit map is built here too, so a fixed box still
// wins over the content it covers on the bitmap path.
func (p *Page) paintFallback(display *Display, drawStart time.Time, track bool, state render.State) error {
	bitmap, err := draw.PaintBitmap(track, &p.stats.paintTime, display)
	if err != nil {
		return err
	}

	previous := p.draw.Display
	previousFallback := p.draw.FallbackDisplay
	p.updateMotionIndex(previous, nil)
	p.captureTransitions(previous, display)
	p.noteCommit(drawStart)
	p.draw.CommitBitmap(display, bitmap.Image)
	p.setDisplayBoxes(display)
	p.generation++
	p.layerGen++
	p.bindMotion(previous, nil)
	p.bindTransitions(nil)
	if previous != nil && previous != display {
		previous.Release()
	}
	if previousFallback != nil && previousFallback != display && previousFallback != previous {
		previousFallback.Release()
	}
	p.remember(state)
	p.markFull()
	p.lastPatch = DisplayPatch{
		Changed: bitmap.Changed,
		Added:   bitmap.Added,
		Regions: bitmap.Regions,
		Full:    bitmap.Full,
	}
	if track {
		p.recordFull(len(display.Boxes))
	}
	p.applyPending()
	p.runAfterRedraw()

	return nil
}

// updateMotionIndex updates the cached *cssmotion.Index for display.
// If display.Keyframes matches the previous display's keyframes (pointer equality
// or length == 0), it reuses the previous *cssmotion.Index instead of calling
// NewIndex on every commit.
func (p *Page) updateMotionIndex(previous, display *Display) *motion.Index {
	if display == nil || len(display.Keyframes) == 0 {
		p.motionIndex = motion.EmptyIndex()
		return p.motionIndex
	}
	if previous != nil && sameKeyframes(display.Keyframes, previous.Keyframes) && p.motionIndex != nil && p.motionIndex != motion.EmptyIndex() {
		if p.motionIndex.Display() == display || len(display.Ops) == p.motionIndex.NumOps() {
			p.motionIndex.Repoint(display)
			return p.motionIndex
		}
	}
	p.motionIndex = motion.NewIndex(display)
	return p.motionIndex
}

// sameKeyframes reports whether two keyframe rule slices share the same rules
// by length or pointer equality.
func sameKeyframes(a, b []layout.KeyframesRule) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	return &a[0] == &b[0]
}

// Redraw renders the current size: a display list for a replayable page,
// a bitmap otherwise. Parsed trees and sheets are reused. When the frame
// matches the last committed one, Redraw returns after the template and
// form sync without laying the page out again: the retained list, boxes,
// and picture already show it. A skipped call still counts the redraw but
// no layout or repaint. Stage timing and
// allocs track only with Perf on. One Page has one owner: never call Redraw
// concurrently on the same Page; keep Tick, MarkRect, and Redraw on the
// page loop the way Run and Serve each own their Page.
func (p *Page) Redraw(ctx context.Context) error {
	return p.redraw(ctx, false)
}

// RedrawIfChanged renders unless the executed source, viewport, focus,
// hover, theme, and images all match the last committed frame. It shares
// Redraw's unchanged check, so an identical frame skips the layout either
// way. It is for callers that redraw after every event and would otherwise
// lay the same frame out again. A skipped call counts no layout or repaint.
func (p *Page) RedrawIfChanged(ctx context.Context) error {
	return p.redraw(ctx, true)
}

// Stats returns one snapshot of the page counters and the last frame sizes.
// LastDraw is the window's draw time once SetDrawTime has recorded one, and
// the display-list or paint stage of the last Redraw before that.
func (p *Page) Stats() host.Stats {
	ops := 0
	if p.draw.Display != nil {
		ops = len(p.draw.Display.Ops)
	}

	counts := draw.Counts{
		FullLayouts:      p.stats.fullLayouts,
		HoverPatches:     p.stats.hoverPatches,
		HoverFallbacks:   p.stats.hoverFallbacks,
		HoverPatchErrors: p.stats.hoverPatchErrors,

		Redraws:         p.stats.redraws,
		Parses:          p.stats.parses,
		Cascades:        p.stats.cascades,
		Layouts:         p.stats.layouts,
		Repaints:        p.stats.repaints,
		LastRedraw:      p.stats.lastRedraw,
		LastDraw:        p.stats.lastDraw,
		FrameDraw:       p.stats.frameDraw,
		LastTemplate:    p.stats.lastTemplate,
		StyleTime:       p.stats.styleTime,
		LayoutTime:      p.stats.layoutTime,
		PaintTime:       p.stats.paintTime,
		OtherTime:       p.stats.otherTime,
		DirtyOps:        p.stats.dirtyOps,
		DirtyRegions:    p.stats.dirtyRegions,
		DirtyArea:       p.stats.dirtyArea,
		ChangedOps:      p.stats.changedOps,
		AllocFrame:      p.stats.allocFrame,
		Reloads:         p.stats.reloads,
		LastReloadError: p.stats.reloadErr,
	}

	stats := draw.Snapshot(&counts, len(p.boxes), ops)
	stats.HoverPatchFailures = p.hoverPatchFailureSnapshot()
	return stats
}

// hashedBody pairs one template body with the FNV-1a hash taken from it.
// The publish and adopt paths write the pair under dataMu; the sync path
// reuses hash when the body about to be synced has the same length and
// backing array as source, so a clean redraw skips folding the bytes. The
// pair is self-consistent even when a failed ReplaceHTML rolls lastSyncBody
// back to an older value: the next body simply misses and is hashed again.
type hashedBody struct {
	source string
	hash   uint64
}

// renderTrack remembers the last committed frame so a Redraw that changes
// none of its inputs can return without laying the page out again.
type renderTrack struct {
	generation uint64
	imagesRev  uint64
	// last is the committed frame key. Once the source memo carries a
	// nonzero revision the key drops the source string (see
	// draw.KeyForHashed), so last pins hash, length, and revision only.
	last    draw.Key
	hasLast bool
	// lastRetain, lastBlink, lastCaret, and lastAnchor are inputs the layout
	// key does not carry. Retained state changes the placement the engine
	// produces. A blink phase change repaints the caret, and so does a move
	// of the caret offset or the anchor. Those offsets stay off the layout
	// key: a collapsed caret does not change the source the key covers.
	lastRetain bool
	lastBlink  bool
	lastCaret  int
	lastAnchor int
	// statePseudos records the interactive pseudo-classes the sheets use. All
	// true means unknown, which keeps the old redraw behavior.
	statePseudos render.StatePseudos
	// hoverGuardHash keys the hover risk scan input: the committed source
	// plus the theme. A hash keeps no extra copy of either string alive.
	hoverGuardHash  uint64
	hoverRiskGlobal bool
	hoverRiskScopes []string
	// hoverScopeKey is built by keyFor like last, so it pins no source
	// copy once the source memo carries a nonzero revision.
	hoverScopeKey      draw.Key
	hoverScopes        map[hoverScopePair]hoverScopeResult
	hoverFailureCounts map[hoverFailureKey]uint64
	// redrawBatch defers draws across a pointer sequence. Any Redraw in the
	// batch sets pending; a Redraw (not RedrawIfChanged) also sets force.
	redrawBatch   int
	redrawPending bool
	redrawForce   bool
	// afterRedraw is the hook SetAfterRedraw registers, with the guards
	// that keep it from re-entering a redraw.
	afterRedraw draw.After
	// sourceHash, sourceLen, and sourceRev memoize p.source. The draw
	// adapter refreshes them when the synced source changes, so keyFor
	// builds the next key with no source scan and Unchanged compares u64s
	// first, memcmping only on a hash hit.
	sourceHash uint64
	sourceLen  int
	sourceRev  uint64
	// lastBody pins the last template body the sync path hashed together
	// with that hash. publishForm and AdoptPatchedSource write it under
	// dataMu, and syncForm copies it into formSnap under the same lock, so
	// a clean redraw that reuses the cached body proves byte identity by
	// backing pointer and skips the FNV fold.
	lastBody hashedBody
	// cacheRev, cacheHash, and cacheLen record the source the style cache
	// was built or relayed for. styleStaged relayouts directly when they
	// match the current source memo, so a same-source restyle takes one
	// cache lock and no Source memcmp.
	cacheRev  uint64
	cacheHash uint64
	cacheLen  int
	// relayed records the pointer/focus state, viewport, theme, and image
	// revision the style cache was last relayed for, by an element patch or
	// a committed relayout. A Redraw asking for the same state reuses the
	// cached styled document instead of cascading again. A hover handler
	// patch followed by the hover's own Redraw is the hot path: one cascade
	// serves both commits instead of one per commit.
	relayedOK          bool
	relayedW, relayedH int
	relayedHover       string
	relayedActive      string
	relayedFocus       string
	relayedTheme       *css.Sheet
	relayedImages      uint64
	// tplProbeTpl and tplProbeHas memoize the template probe. The probe
	// moves to template-write time: execTemplate rescans nav.Current for
	// "{{" only when p.tpl changed, and reuses the verdict otherwise. A
	// builder document keeps both at zero (staticSource), so execTemplate
	// returns nav.Current with no scan and no Execute.
	tplProbeTpl *template.Template
	tplProbeHas bool
	// execSkipped reports that execTemplate returned without executing:
	// either the template is static or the form is clean and the cached
	// body was reused. The SyncForm adapter consumes it.
	execSkipped bool
}

func (p *Page) viewport() draw.Viewport {
	return draw.Viewport{Width: p.width, Height: p.height}
}

func (p *Page) keyFor(state render.State) draw.Key {
	// The hash, length, and revision memoized beside p.source travel in
	// the key, and the hashed constructor drops the source string once
	// the revision is nonzero, so neither the transient key nor the
	// committed last pins an extra copy of the page.
	return draw.KeyForHashed(
		p.source,
		p.sourceHash,
		p.sourceLen,
		p.sourceRev,
		p.viewport(),
		state,
		p.statePseudos,
		p.imagesRev,
	)
}

func (p *Page) unchanged(state render.State) bool {
	// A retained-state debt, or a retained flag the last commit did not lay
	// out with, needs a fresh placement. So does a running animation or an
	// in-flight transition, whose pose lives in the operations a publish
	// would replace. A blink phase change repaints the caret, and a caret
	// or anchor move does too, without a new layout key. Everything else
	// the display shows is in the key.
	if p.restampRedraw || p.retainState != p.lastRetain {
		return false
	}
	if p.motion != nil && p.motion.Active() {
		return false
	}
	blink := false
	caret := 0
	anchor := 0
	if fs := p.formState(); fs != nil {
		blink = fs.BlinkOn
		caret = fs.Caret
		anchor = fs.Anchor
	}
	if blink != p.lastBlink {
		return false
	}
	if caret != p.lastCaret || anchor != p.lastAnchor {
		return false
	}
	return draw.Unchanged(p.hasLast, p.last, p.keyFor(state))
}

func (p *Page) remember(state render.State) {
	p.last = p.keyFor(state)
	// The committed key keeps identity, not bytes: with a nonzero
	// revision Unchanged never reads Source, so drop the second copy of
	// the page. A zero revision (no noteSourceHashed yet) keeps the string
	// for the memcmp fallback.
	if p.last.SourceRev != 0 {
		p.last.Source = ""
	}
	p.hasLast = true
	p.lastRetain = p.retainState
	p.lastBlink = false
	p.lastCaret = 0
	p.lastAnchor = 0
	if fs := p.formState(); fs != nil {
		p.lastBlink = fs.BlinkOn
		p.lastCaret = fs.Caret
		p.lastAnchor = fs.Anchor
	}
}

// noteSource records a newly synced source with a fresh hash. noteSourceHashed
// carries the store so a caller that already hashed the bytes skips the fold.
func (p *Page) noteSource(next string) {
	p.noteSourceHashed(next, draw.HashString(next))
}

// noteCacheSource records the source the style cache now reflects.
func (p *Page) noteCacheSource() {
	p.cacheRev = p.sourceRev
	p.cacheHash = p.sourceHash
	p.cacheLen = p.sourceLen
}

// cacheFresh reports that the style cache reflects the current source memo,
// so a restyle can relayout directly: one cache lock, no Source memcmp.
func (p *Page) cacheFresh() bool {
	return p.cache != nil && p.sourceRev != 0 && p.cacheRev == p.sourceRev &&
		p.cacheHash == p.sourceHash && p.cacheLen == p.sourceLen
}

// noteRelayed records the state the style cache now reflects. Call it after
// every path that (re)cascades the cache: an element patch relay and a
// committed styleStaged relayout alike, so a later Redraw for the same state
// reuses the document instead of cascading again.
func (p *Page) noteRelayed(state render.State) {
	p.relayedOK = true
	p.relayedW, p.relayedH = p.width, p.height
	p.relayedHover, p.relayedActive, p.relayedFocus = state.Hover, state.Active, state.Focus
	p.relayedTheme = state.Theme
	p.relayedImages = p.imagesRev
}

// relayedDoc returns the cached styled document when the cache already
// reflects state: same viewport, pointer/focus ids, theme, and image
// revision as the last relay. The caller must have checked cacheFresh, so
// the tree is also the current source's. A hit skips one css.Relayout.
func (p *Page) relayedDoc(state render.State) (*css.Document, bool) {
	if !p.relayedOK || p.cache == nil {
		return nil, false
	}
	if p.relayedW != p.width || p.relayedH != p.height {
		return nil, false
	}
	if p.relayedHover != state.Hover || p.relayedActive != state.Active || p.relayedFocus != state.Focus {
		return nil, false
	}
	if p.relayedTheme != state.Theme || p.relayedImages != p.imagesRev {
		return nil, false
	}
	doc := p.cache.Styled()
	if doc == nil {
		return nil, false
	}
	return doc, true
}

// invalidateCache drops the parsed tree so the next Redraw rebuilds it. It
// also evicts the box-cache entry for the discarded boxes, so a source
// change on a live page does not pin the old box list until the next commit.
func (p *Page) invalidateCache() {
	p.cache = nil
	p.relayedOK = false
	p.forgetBoxes()
	p.motionIndex = nil
	// Unknown sheets keep every state redraw until the next Redraw reports
	// what the new sheets can match.
	p.statePseudos = draw.UnknownPseudos()
}

// FallbackDisplay returns the retained operations behind a bitmap page.
// Hosts use them to repaint fixed content correctly when the viewport scrolls.
// Callers must not mutate the returned display.
func (p *Page) FallbackDisplay() *layout.Display { return p.draw.FallbackDisplay }

// SetAfterRedraw registers fn to run after a redraw commits a new display.
//
// The callback runs on the page loop, at the end of the commit, so the new
// display and the new hit boxes are installed and the app can read them. An
// app that keeps its state in the parsed tree uses it to reapply that state
// after the template replaced the tree.
//
// The callback cannot deadlock: it runs after the render cache released its
// lock, so an edit of the tree does not wait on itself. It cannot re-enter a
// redraw either: a Redraw asked for inside the callback waits for the commit
// to finish, and the page draws again once the callback returned.
//
// nil removes the callback. Register it from the page loop, like SetTick.
func (p *Page) SetAfterRedraw(fn func()) {
	if p == nil {
		return
	}

	p.afterRedraw.Set(fn)
}

// runAfterRedraw calls the registered callback after a commit installed the
// display and the boxes. A callback that asks for a redraw does not re-enter
// the commit: the request waits for the callback to return.
func (p *Page) runAfterRedraw() {
	p.afterRedraw.Run()
}

// afterRedrawDraw serves the redraw the callback asked for. The draw runs
// after the callback returned, so the commit is never re-entered, and the
// callback runs again in it, so state reapplied to a rebuilt tree lands. One
// draw is enough: a callback that always asks for a draw cannot loop.
func (p *Page) afterRedrawDraw(ctx context.Context) error {
	if !p.afterRedraw.Take() {
		return nil
	}

	defer p.afterRedraw.Done()

	return p.redraw(ctx, false)
}

// BeginRedrawBatch defers page draws until EndRedrawBatch. The window wraps a
// press, focus, click, and release in one batch so the final state draws once.
//
// A redraw that a failed draw left queued stays queued, so the flags survive
// into the next batch and it commits the redraw.
func (p *Page) BeginRedrawBatch() {
	draw.OpenBatch(&p.redrawBatch)
}

// EndRedrawBatch closes a batch and draws the final state once. An inner batch
// leaves the draw to the outermost one. The flush runs the redraw sequence
// once; when the frame is unchanged it counts the redraw without laying the
// page out again, whether the batch held plain Redraw or RedrawIfChanged
// calls.
//
// A draw that fails leaves the redraw queued, so the next batch commits it
// and the committed frame catches up with the state the batch changed.
func (p *Page) EndRedrawBatch(ctx context.Context) error {
	run, force := draw.CloseBatch(&p.redrawBatch, &p.redrawPending, &p.redrawForce)
	if !run {
		return nil
	}

	if err := p.redraw(ctx, !force); err != nil {
		draw.RestoreBatch(&p.redrawPending, &p.redrawForce, force)

		return err
	}

	return nil
}
