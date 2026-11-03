package page

import (
	"context"
	"errors"
	"hash/fnv"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// hoverPatchFunc re-lays one element's subtree. It mirrors
// render.RelayoutElement so tests can observe the roots a hover change
// patches without touching the engine.
type hoverPatchFunc func(
	ctx context.Context,
	doc *css.Document,
	prev *Display,
	elementID string,
	options layout.Options,
) (*layout.ElementPatch, error)

// errHoverPatchEmpty reports a patch that returned no display. The caller
// falls back to the full layout, so the page draws what it drew before.
var errHoverPatchEmpty = errors.New("page: hover patch returned no display")

// layoutStaged lays out the styled document for one redraw. A change that
// only moves the hover re-places the subtrees the new hover state can reach
// instead of laying the whole page out again. Anything else, and any patch
// the engine refuses, lays the page out exactly as before, so a page the
// patch cannot represent draws what a full layout would draw.
func (p *Page) layoutStaged(
	ctx context.Context,
	styled *css.Document,
	state render.State,
) (*Display, error) {
	if roots, ok := p.hoverPatchRoots(styled, state); ok {
		if patched, err := p.listHoverPatched(ctx, styled, state, roots); err == nil {
			return patched, nil
		}
	}

	return draw.List(ctx, styled, state)
}

// hoverPatchRoots returns the subtree roots a hover-only change must
// re-lay out. ok is false when the frame changed more than the hover, the
// page cannot patch, or the sheets can style outside the hovered subtrees.
// maxHoverScopes caps the hover scope cache before it is cleared.
const maxHoverScopes = 256

func (p *Page) hoverPatchRoots(styled *css.Document, state render.State) ([]string, bool) {
	if !p.hoverPatchEligible(state) || p.last.Hover == state.Hover {
		return nil, false
	}

	if key := hoverGuardHash(p.source, p.themeSrc); key != p.hoverGuardHash {
		p.hoverGuardHash = key
		p.hoverGuardRisky = draw.HoverPatchRisky(p.source, p.themeSrc)
	}

	if p.hoverGuardRisky {
		return nil, false
	}

	key := p.keyFor(state)
	key.Hover = ""
	if p.hoverScopes == nil || p.hoverScopeKey != key {
		p.hoverScopeKey = key
		p.hoverScopes = make(map[hoverScopePair]hoverScopeResult)
	}
	pair := hoverScopePair{p.last.Hover, state.Hover}
	if cached, found := p.hoverScopes[pair]; found {
		return cached.roots, cached.ok
	}
	roots, ok := p.findHoverScopes(styled, state)
	if len(p.hoverScopes) >= maxHoverScopes {
		clear(p.hoverScopes)
	}
	p.hoverScopes[pair] = hoverScopeResult{roots: roots, ok: ok}
	return roots, ok
}

type hoverScopePair struct{ previous, current string }
type hoverScopeResult struct {
	roots []string
	ok    bool
}

func (p *Page) findHoverScopes(styled *css.Document, state render.State) ([]string, bool) {
	// The cascade knows which selectors changed. Prefer those scopes over
	// the common ancestor of two controls, which can include a whole toolbar.
	if roots, ok := styled.HoverChangedScopes(p.last.Hover); ok && draw.HoverScopesFit(p.boxes, roots) {
		return roots, true
	}
	return draw.HoverPatchRoots(p.boxes, p.last.Hover, state.Hover)
}

// hoverPatchEligible reports whether this redraw changed only the hover:
// the committed frame key matches the next one once the hover is ignored,
// the committed display carries the retained layout state the patch splices
// into, and no retained-state debt is pending.
func (p *Page) hoverPatchEligible(state render.State) bool {
	if !p.hasLast || p.draw.Display == nil || p.cache == nil {
		return false
	}

	if !p.retainState || !p.lastRetain || p.restampRedraw {
		return false
	}

	prev, next := p.last, p.keyFor(state)
	prev.Hover, next.Hover = "", ""

	return prev == next
}

// hoverGuardHash keys the risk scan input without keeping the source alive.
func hoverGuardHash(source, theme string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(source))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(theme))

	return h.Sum64()
}

// listHoverPatched re-lays each root in turn and returns the patched page.
// The engine splices the patches into the retained placement, so the
// returned display carries the whole page. An error leaves the committed
// display for the full layout the caller falls back to: every fallible check
// runs before the splice, so a refused patch changes nothing.
func (p *Page) listHoverPatched(
	ctx context.Context,
	styled *css.Document,
	state render.State,
	roots []string,
) (*Display, error) {
	current := p.draw.Display
	options := layout.Options{Images: state.Images, RetainState: true}

	for _, root := range roots {
		patch, err := p.hoverPatch(ctx, styled, current, root, options)
		if err != nil {
			return nil, err
		}

		if patch == nil || patch.Display == nil {
			return nil, errHoverPatchEmpty
		}

		current = patch.Display
	}

	return current, nil
}
