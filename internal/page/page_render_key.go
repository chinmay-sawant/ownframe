package page

import (
	"html/template"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

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
	// hoverGuardRisky records whether those sheets can style outside the
	// hovered subtrees, which disables the scoped hover patch.
	hoverGuardHash  uint64
	hoverGuardRisky bool
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
