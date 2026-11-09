package page

import (
	"context"
	"errors"
	"hash/fnv"
	"slices"
	"strings"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
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

type hoverPatchAttemptError struct {
	target string
	err    error
}

type hoverFailureKey struct {
	stage  string
	reason string
	target string
}

func (e *hoverPatchAttemptError) Error() string { return e.err.Error() }
func (e *hoverPatchAttemptError) Unwrap() error { return e.err }

// errHoverPatchEmpty reports a patch that returned no display. The caller
// falls back to the full layout, so the page draws what it drew before.
var errHoverPatchEmpty = errors.New("page: hover patch returned no display")

const maxHoverFailureKeys = 256

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
	hoverChanged := p.hoverPatchEligible(state) && p.last.Hover != state.Hover
	if roots, ok := p.hoverPatchRoots(styled, state); ok {
		patched, err := p.listHoverPatched(ctx, styled, state, roots)
		if err == nil {
			p.stats.hoverPatches++
			return patched, nil
		}
		p.stats.hoverPatchErrors++
		p.noteHoverPatchFailure("selector", classifyHoverPatchError(err), hoverFailureTarget(err, roots))
		// If the cascade roots failed (e.g. an element transitioned to display:none
		// and placed no box), fall back to geometric roots only when the sheets
		// contain no escaping selector at all. The geometric roots cannot prove
		// which scope a :has or sibling-after-:hover rule belongs to.
		if !p.hoverRiskAny() {
			if geoRoots, geoOK := draw.HoverPatchRoots(p.boxes, p.last.Hover, state.Hover); geoOK {
				patched, retryErr := p.listHoverPatched(ctx, styled, state, geoRoots)
				if retryErr == nil {
					p.stats.hoverPatches++
					return patched, nil
				}
				p.stats.hoverPatchErrors++
				p.noteHoverPatchFailure("geometry", classifyHoverPatchError(retryErr), hoverFailureTarget(retryErr, geoRoots))
			}
		}
	}

	if hoverChanged {
		p.stats.hoverFallbacks++
		if _, ok := p.hoverPatchRoots(styled, state); !ok {
			reason := "no-patch-roots"
			if p.hoverIsRisky(p.last.Hover, state.Hover) {
				reason = "escaping-hover-selector"
			}
			p.noteHoverPatchFailure("selection", reason, []string{p.last.Hover, state.Hover})
		}
	}
	p.stats.fullLayouts++
	return draw.List(ctx, styled, state)
}

func classifyHoverPatchError(err error) string {
	switch {
	case errors.Is(err, errHoverPatchEmpty):
		return "empty-display"
	case errors.Is(err, layout.ErrNoElement):
		return "target-missing"
	case errors.Is(err, layout.ErrNoRetainedState):
		return "retained-state-missing"
	case errors.Is(err, layout.ErrOverrideReflow):
		return "reflow-escapes-patch"
	case errors.Is(err, layout.ErrInvalidOverrideValue):
		return "invalid-override"
	case errors.Is(err, layout.ErrUnknownOverrideProperty):
		return "unsupported-override"
	case errors.Is(err, context.Canceled):
		return "context-canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "context-deadline"
	default:
		return "patch-error"
	}
}

func hoverFailureTarget(err error, roots []string) []string {
	var attempt *hoverPatchAttemptError
	if errors.As(err, &attempt) {
		return []string{attempt.target}
	}
	return roots
}

func (p *Page) noteHoverPatchFailure(stage, reason string, targets []string) {
	if p.hoverFailureCounts == nil {
		p.hoverFailureCounts = make(map[hoverFailureKey]uint64)
	}
	if len(targets) == 0 {
		targets = []string{"<none>"}
	}
	for _, target := range targets {
		key := hoverFailureKey{stage: stage, reason: reason, target: target}
		if _, exists := p.hoverFailureCounts[key]; !exists && len(p.hoverFailureCounts) >= maxHoverFailureKeys-1 {
			key.target = "<other>"
		}
		p.hoverFailureCounts[key]++
	}
}

func (p *Page) hoverPatchFailureSnapshot() []host.HoverPatchFailureCount {
	if len(p.hoverFailureCounts) == 0 {
		return nil
	}
	items := make([]host.HoverPatchFailureCount, 0, len(p.hoverFailureCounts))
	for key, count := range p.hoverFailureCounts {
		items = append(items, host.HoverPatchFailureCount{
			Stage: key.stage, Reason: key.reason, Target: key.target, Count: count,
		})
	}
	slices.SortFunc(items, func(a, b host.HoverPatchFailureCount) int {
		if c := strings.Compare(a.Stage, b.Stage); c != 0 {
			return c
		}
		if c := strings.Compare(a.Reason, b.Reason); c != 0 {
			return c
		}
		return strings.Compare(a.Target, b.Target)
	})
	return items
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
		p.hoverRiskGlobal, p.hoverRiskScopes = draw.HoverPatchRiskyScopes(p.source, p.themeSrc)
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
	// The cascade evaluates the full selector per node (including :has and
	// sibling combinators), so its scopes stay safe when the sheets use those
	// patterns. The geometric fallback cannot prove which element a risky
	// selector belongs to, so any escaping selector in the sheets forces a
	// full layout once the cascade scopes do not fit.
	if roots, ok := styled.HoverChangedScopes(p.last.Hover); ok {
		if draw.HoverScopesFit(p.boxes, roots) {
			return roots, true
		}
		if p.hoverRiskAny() {
			return nil, false
		}
	} else if p.hoverRiskAny() {
		return nil, false
	}
	return draw.HoverPatchRoots(p.boxes, p.last.Hover, state.Hover)
}

// hoverRiskAny reports whether the sheets contain any selector that can style
// outside the hovered subtrees. The geometric fallback uses this page-wide
// answer because it cannot map a selector scope to the nodes it affects.
func (p *Page) hoverRiskAny() bool {
	return p.hoverRiskGlobal || len(p.hoverRiskScopes) > 0
}

// hoverIsRisky reports whether an escaping rule (:has or sibling-after-:hover)
// can match when hover moves between oldID and newID.
func (p *Page) hoverIsRisky(oldID, newID string) bool {
	if p.hoverRiskGlobal {
		return true
	}
	if len(p.hoverRiskScopes) == 0 {
		return false
	}

	// Filter down to scopes that actually appear in the HTML source.
	// An unused component stylesheet cannot affect any node in the page.
	// This scoping is advisory: the patch fallback uses hoverRiskAny, which
	// stays page-wide so a missed scope can never accept a narrower patch.
	var activeScopes []string
	for _, scope := range p.hoverRiskScopes {
		if scope != "" && strings.Contains(p.source, scope) {
			activeScopes = append(activeScopes, scope)
		}
	}
	if len(activeScopes) == 0 {
		return false
	}

	return p.elementOrAncestorsMatchScopes(oldID, activeScopes) ||
		p.elementOrAncestorsMatchScopes(newID, activeScopes)
}

func (p *Page) elementOrAncestorsMatchScopes(id string, scopes []string) bool {
	if id == "" || len(p.boxes) == 0 {
		return false
	}

	start := -1
	for i := range p.boxes {
		if p.boxes[i].ID == id {
			start = i
			break
		}
	}
	if start < 0 {
		return p.sourceElementMatchesScopes(id, scopes)
	}

	for idx := start; idx >= 0 && idx < len(p.boxes); {
		boxID := p.boxes[idx].ID
		if boxID != "" {
			for _, scope := range scopes {
				if boxID == scope {
					return true
				}
			}
			if p.sourceElementMatchesScopes(boxID, scopes) {
				return true
			}
		}
		parent := p.boxes[idx].ParentIndex
		if parent < 0 || parent >= len(p.boxes) || parent == idx {
			break
		}
		idx = parent
	}

	return false
}

// sourceElementMatchesScopes checks if the tag declaring id in p.source carries
// any of the scope classes or attributes.
func (p *Page) sourceElementMatchesScopes(id string, scopes []string) bool {
	if id == "" || p.source == "" {
		return false
	}
	needle := `id="` + id + `"`
	pos := strings.Index(p.source, needle)
	if pos < 0 {
		needle = `id='` + id + `'`
		pos = strings.Index(p.source, needle)
	}
	if pos < 0 {
		return false
	}

	tagStart := strings.LastIndexByte(p.source[:pos], '<')
	if tagStart < 0 {
		return false
	}
	tagRelEnd := strings.IndexByte(p.source[pos:], '>')
	if tagRelEnd < 0 {
		return false
	}
	tagText := p.source[tagStart : pos+tagRelEnd+1]

	for _, scope := range scopes {
		if scope != "" && strings.Contains(tagText, scope) {
			return true
		}
	}

	return false
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
			if current != nil && current != p.draw.Display {
				current.Release()
			}
			p.commitRoots = nil
			return nil, &hoverPatchAttemptError{target: root, err: err}
		}

		if patch == nil || patch.Display == nil {
			if current != nil && current != p.draw.Display {
				current.Release()
			}
			p.commitRoots = nil
			return nil, &hoverPatchAttemptError{target: root, err: errHoverPatchEmpty}
		}

		if current != nil && current != p.draw.Display && current != patch.Display {
			current.Release()
		}
		current = patch.Display
	}

	// The commit this display produces is scoped to these roots. An empty
	// non-nil scope means the hover changed no selector: still scoped, so
	// the transition capture can skip when no transitioning box is inside.
	p.commitRoots = append([]string{}, roots...)

	return current, nil
}
