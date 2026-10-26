package page

import (
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
	// lastRetain and lastBlink are the layout-mode and caret inputs the key
	// does not carry: retained layout state changes the placement the engine
	// produces, and a blink phase change repaints the caret.
	lastRetain bool
	lastBlink  bool
	// statePseudos records the interactive pseudo-classes the sheets use. All
	// true means unknown, which keeps the old redraw behavior.
	statePseudos render.StatePseudos
	// hoverGuardHash keys the hover risk scan input: the committed source
	// plus the theme. A hash keeps no extra copy of either string alive.
	// hoverGuardRisky records whether those sheets can style outside the
	// hovered subtrees, which disables the scoped hover patch.
	hoverGuardHash  uint64
	hoverGuardRisky bool
	// redrawBatch defers draws across a pointer sequence. Any Redraw in the
	// batch sets pending; a Redraw (not RedrawIfChanged) also sets force.
	redrawBatch   int
	redrawPending bool
	redrawForce   bool
	// afterRedraw is the hook SetAfterRedraw registers, with the guards
	// that keep it from re-entering a redraw.
	afterRedraw draw.After
}

func (p *Page) viewport() draw.Viewport {
	return draw.Viewport{Width: p.width, Height: p.height}
}

func (p *Page) keyFor(state render.State) draw.Key {
	return draw.KeyFor(p.source, p.viewport(), state, p.statePseudos, p.imagesRev)
}

func (p *Page) unchanged(state render.State) bool {
	// A retained-state debt, or a retained flag the last commit did not lay
	// out with, needs a fresh placement. So does a running animation or an
	// in-flight transition, whose pose lives in the operations a publish
	// would replace, and a blink phase change, which repaints the caret.
	// Everything else the display shows is in the key.
	if p.restampRedraw || p.retainState != p.lastRetain {
		return false
	}
	if p.motion != nil && p.motion.Active() {
		return false
	}
	blink := false
	if fs := p.formState(); fs != nil {
		blink = fs.BlinkOn
	}
	if blink != p.lastBlink {
		return false
	}
	return draw.Unchanged(p.hasLast, p.last, p.keyFor(state))
}

func (p *Page) remember(state render.State) {
	p.last = p.keyFor(state)
	p.hasLast = true
	p.lastRetain = p.retainState
	p.lastBlink = false
	if fs := p.formState(); fs != nil {
		p.lastBlink = fs.BlinkOn
	}
}
