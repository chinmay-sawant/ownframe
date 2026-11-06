package page

import (
	"html/template"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// renderTrack remembers the last committed frame so a Redraw that changes
// none of its inputs can return without laying the page out again.
type renderTrack struct {
	generation uint64
	imagesRev  uint64
	last       draw.Key
	hasLast    bool
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
	hoverScopeKey   draw.Key
	hoverScopes     map[hoverScopePair]hoverScopeResult
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
	// bodyHash and bodyLen memoize the last executed template body for the
	// same reason: the adapter records them when Execute runs, and the
	// clean-form skip reuses the body with no rescan.
	bodyHash uint64
	bodyLen  int
	// cacheRev, cacheHash, and cacheLen record the source the style cache
	// was built or relayed for. styleStaged relayouts directly when they
	// match the current source memo, so a same-source restyle takes one
	// cache lock and no Source memcmp.
	cacheRev  uint64
	cacheHash uint64
	cacheLen  int
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

// noteSource records a newly synced source. Equal hash and length keep the
// revision, so an unchanged frame compares u64s with no memcmp; a change
// bumps the revision the key carries.
func (p *Page) noteSource(next string) {
	hash := draw.HashString(next)
	if p.sourceRev != 0 && hash == p.sourceHash && len(next) == p.sourceLen {
		return
	}

	p.sourceHash = hash
	p.sourceLen = len(next)
	p.sourceRev++
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
