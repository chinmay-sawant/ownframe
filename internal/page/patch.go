package page

import (
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/patch"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/chinmay-sawant/ownframe/internal/state"
	"hash/fnv"
	"slices"
	"strings"
	"sync"
)

// displayOpPool rents operation buffers for display patches and frame ticks.
// It stays in this file so the gochecknoglobals allowlist still covers it.
//
// Retention note: unlike the single-slot dirty arenas, a sync.Pool drains
// across garbage collections, so a giant frame's grown Ops backing array
// pins only transiently between collections. Dropping over-cap buffers
// instead of pooling them would belong in patch.Rent/Free and their
// patch_wire.go call sites, which live outside this file.
var displayOpPool = sync.Pool{
	New: func() any {
		return patch.NewOpBuffer()
	},
}

// DisplayPatch reports how a retained display changed.
type DisplayPatch = patch.DisplayPatch

type opBuffer = patch.OpBuffer

// ElementEdit is one change to an element in the retained tree: an attribute
// to set, or the text to replace. A nil Text and an empty Attribute change
// nothing.
//
// ID names the element the edit applies to. An empty ID means the element
// the patch is for, and an edit on a descendant rides along with the patch.
// Text replaces the text content and drops the children.
//
// Text set this way does not go through the emoji pass. The pass runs on the
// source before it is parsed, and a text edit lands on the parsed tree, so an
// emoji in patched text paints as text rather than as the color image the
// same emoji gets from the template. The engine's tree API only writes a text
// node, so the image cannot be built inside the tree.
type ElementEdit = patch.ElementEdit

// LastPatch returns the last patch report; use it.
func (p *Page) LastPatch() DisplayPatch {
	return p.lastPatch
}

// Deprecated: use LastPatch.
func (p *Page) DisplayPatch() DisplayPatch {
	return p.LastPatch()
}

// PatchDisplay mutates the retained display and records the regions that
// changed. It returns the patch report. A nil display reports nothing.
func (p *Page) PatchDisplay(fn func(d *Display)) DisplayPatch {
	if p.draw.Display == nil || fn == nil {
		return p.lastPatch
	}

	p.tickMutated.Store(true)
	diff, buf := patch.Update(&displayOpPool, p.draw.Display, fn)
	if diff.Patch.Full {
		markFull(p)
	} else {
		for i := range diff.N {
			markRect(p, diff.Rects[i])
		}
	}

	report := patch.WithProps(diff.Patch, p.ChangedProperties())
	p.lastPatch = report
	patch.Free(&displayOpPool, buf)

	return report
}

// PatchOp mutates one retained operation and dirties its old and new bounds.
func (p *Page) PatchOp(i int, fn func(o *DisplayOp)) bool {
	p.tickMutated.Store(true)
	r, ok := patch.OpDamage(p.draw.Display, i, fn)
	if !ok {
		return false
	}

	markRect(p, r)
	p.lastPatch = patch.WithProps(patch.DisplayPatch{Changed: 1, Regions: 1}, p.ChangedProperties())

	return true
}

func rentOps(p *Page, n int) *opBuffer {
	return patch.Rent(&displayOpPool, n)
}

func freeOps(p *Page, buf *opBuffer) {
	patch.Free(&displayOpPool, buf)
}

// PatchElement re-lays out one element of the retained page after editing it
// in the cached tree. The engine lays the element's subtree out again and
// hands the patched page back, so the page adopts it and marks the changed
// region dirty.
//
// The patch is atomic: a refused patch puts the tree back to the source's
// state.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) error {
	if p == nil {
		return ErrNilPage
	}
	defer notifyFrame(p)
	clear(p.hoverScopes)

	p.tickMutated.Store(true)
	var state render.State
	err := patch.ApplyElement(ctx, patch.ElementIn{
		Refused: ErrPatchRefused,
		ID:      id,
		Edits:   edits,
		Cache:   p.cache,
		Display: p.draw.Display,
		Retain:  p.retainState,
		Width:   p.width,
		Height:  p.height,
		State: func() render.State {
			state = renderState(p)
			return state
		},
		Adopt: func(d *Display) {
			commitPatchedDisplay(p, d, []string{id}, func() {
				refreshAccessibility(p, id, edits)
			})
		},
		Rollback: func() { rollbackPatch(p) },
	})
	if err != nil {
		return err
	}

	finishPatchedCommit(p, state)

	return nil
}

// commitPatchedDisplay adopts a patched display from PatchElement or
// PatchElements. The engine hands over the whole display because a patch that
// changes the operation count moves every later operation index. The page
// captures transitions against the previous display, rebuilds its box index,
// marks changed operations dirty, rebinds motion and transitions, releases the
// replaced display, and refreshes accessibility for the patched subtrees.
func commitPatchedDisplay(p *Page, d *Display, roots []string, refreshA11y func()) {
	p.commitRoots = append([]string{}, roots...)
	// RelayoutElements only reuses prev.Boxes in place on a paint-only batch
	// (no geometry, sticky, animation, or transition declaration changes).
	paintOnly := p.draw.Display != nil && d != nil && sameBoxList(p.draw.Display.Boxes, d.Boxes)
	if paintOnly {
		carryDisplayDeclarations(p, p.draw.Display, d)
	}
	captures := captureTransitions(p, p.draw.Display, d)
	prev := adoptSplicedPatch(p, d, paintOnly)
	bindMotion(p, prev, d)
	bindTransitions(p, captures)
	if prev != nil && prev != d {
		releaseDisplay(p, prev)
	}
	if refreshA11y != nil {
		refreshA11y()
	}
}

// finishPatchedCommit updates the cached pseudo-state, relay memo, and
// committed frame key after a subtree patch commits.
func finishPatchedCommit(p *Page, state render.State) {
	p.statePseudos = p.cache.StatePseudos()
	noteRelayed(p, state)
	// The patched display is now the committed frame for the patched subtrees.
	// Record its key while preserving previously committed pointer, focus, and
	// theme state so a click handler skips a redundant full layout without
	// suppressing a hover handler's trailing hover redraw.
	rememberPatched(p, state)
}

// adoptSpliced installs a display the engine spliced and returns the list the
// page drew before it. It is the shared adoption body for an element edit and
// for the layout driver's relayout: boxes and hits are rebuilt, the difference
// between the two displays is marked dirty, and the generation is bumped. A
// moved sibling is outside the patched element's own box, so the display diff
// is the damage source, not ElementPatch.Old and New alone.
func adoptSpliced(p *Page, d *Display) *Display {
	return adoptSplicedPatch(p, d, false)
}

func adoptSplicedPatch(p *Page, d *Display, paintOnly bool) *Display {
	p.tickMutated.Store(true)
	prev := p.draw.SwapList(d)
	if paintOnly && p.contentOK && p.contentW == d.Width && p.contentH == d.Height && !p.hasViewportSticky && len(p.hits) == 0 {
		// Paint-only splice reused prev.Boxes in place with identical geometry
		// and no sticky/fixed boxes: p.boxes, p.index, p.content, and
		// p.hasViewportSticky are already current.
		p.boxes = d.Boxes
	} else {
		setDisplayBoxes(p, d)
	}
	dirtyFromDisplay(p, prev, d)
	p.generation++
	p.layerGen++
	p.stats.layouts++
	p.stats.repaints++

	return prev
}

// rollbackPatch undoes the edits of a refused patch. The cached tree is
// dropped, so the next redraw parses the source again and the page shows the
// source's state. The render key is forgotten, so it redraws.
func rollbackPatch(p *Page) {
	invalidateCache(p)
	p.hasLast = false
}

// ChangedProperties returns a snapshot, falling back to the bound store when empty.
func (p *Page) ChangedProperties() []string {
	if p == nil {
		return nil
	}

	return p.stateBinding.ChangedProperties()
}

// TakeChangedProperties drains local metadata; the store fallback remains a snapshot.
func takeChangedProperties(p *Page) []string {
	if p == nil {
		return nil
	}

	return p.stateBinding.TakeChangedProperties()
}

// SetChangedProperties copies handler metadata into the binding.
func (p *Page) SetChangedProperties(props []string) {
	if p == nil {
		return
	}

	p.stateBinding.SetChangedProperties(props)
}

// AddChangedProperty appends a nonempty property path.
func (p *Page) AddChangedProperty(prop string) {
	if p == nil {
		return
	}

	p.stateBinding.AddChangedProperties(prop)
}

// AddChangedProperties appends nonempty property paths.
func (p *Page) AddChangedProperties(props []string) {
	if p == nil {
		return
	}

	p.stateBinding.AddChangedProperties(props...)
}

// ClearChangedProperties clears local metadata without clearing the store.
func (p *Page) ClearChangedProperties() {
	if p == nil {
		return
	}

	p.stateBinding.ClearChangedProperties()
}

// State returns the bound state store.
func (p *Page) State() *state.Store { return p.stateBinding.Store() }

// BindState replaces the subscription. Bind nil to detach.
func bindState(p *Page, s *state.Store) {
	if p != nil {
		p.stateBinding.Bind(s)
	}
}

// SyncState refreshes surfaced metadata from the current binding only.
func syncState(p *Page) {
	if p != nil {
		p.stateBinding.Sync()
	}
}

func ensureElements(p *Page) *element.Registry {
	if p == nil {
		return patch.Ensure(nil)
	}

	p.elements = patch.Ensure(p.elements)

	return p.elements
}

// Element returns a retained handle for id.
func (p *Page) Element(id string) *element.Node {
	return patch.Node(ensureElements(p), id)
}

// ElementRegistry exposes the underlying registry.
func (p *Page) ElementRegistry() *element.Registry {
	return ensureElements(p)
}

// ElementPatch is one element's share of a batched patch: the element id
// plus the tree edits for it. An edit with an empty ID targets the element
// the patch names, the same rule ElementEdit documents.
type ElementPatch struct {
	ID    string
	Edits []ElementEdit
}

// PatchElements edits several retained elements in one batch and re-lays
// them out together. An empty batch returns nil without touching the page.
//
// OPT-07-02: every patch's tree edits land first, then the page runs ONE
// shared style re-resolve (RelayoutFor) for the whole batch instead of one
// per element. Each touched element is re-laid out in order against that
// single styled document, chaining from one element's patched display to
// the next, and the final display is spliced in and drawn once, so a batch
// of N patches pays one re-resolve, N subtree relayouts, and one adoption.
//
// The batch is atomic: a refused element rolls the tree back to the
// source's state and the error wraps ErrPatchRefused, so callers fall back
// to a refresh. An entry whose edits match nothing rides along without
// failing the batch; the batch is refused only when no entry changes
// anything, the way one PatchElement with no change is refused.
func (p *Page) PatchElements(ctx context.Context, patches []ElementPatch) error {
	if p == nil {
		return ErrNilPage
	}
	if len(patches) == 0 {
		return nil
	}
	defer notifyFrame(p)
	clear(p.hoverScopes)

	p.tickMutated.Store(true)
	for i := range patches {
		if err := patch.Ready(ErrPatchRefused, patches[i].ID, p.cache != nil, p.draw.Display != nil, p.retainState); err != nil {
			return err
		}
	}

	touched := make([]bool, len(patches))
	changed := false
	for i := range patches {
		if patch.Apply(p.cache, patches[i].ID, patches[i].Edits) {
			touched[i] = true
			changed = true
		}
	}
	if !changed {
		return fmt.Errorf("no change: %w", ErrPatchRefused)
	}

	state := renderState(p)
	// OPT-07-02: one shared style re-resolve for the whole batch. The
	// per-element loop below re-places subtrees only; it never re-resolves.
	styled, err := p.cache.RelayoutFor(ctx, p.width, p.height, state)
	if err != nil {
		return failBatch(p, err)
	}

	ids := make([]string, 0, len(patches))
	for i := range patches {
		if touched[i] {
			ids = append(ids, patches[i].ID)
		}
	}

	current, err := render.RelayoutElements(
		ctx, styled, p.draw.Display, ids, layout.Options{Images: state.Images, RetainState: true},
	)
	if err != nil {
		return failBatch(p, err)
	}

	commitPatchedDisplay(p, current, ids, func() {
		for i := range patches {
			if touched[i] {
				refreshAccessibility(p, patches[i].ID, patches[i].Edits)
			}
		}
	})
	p.cache.SetStyled(styled)
	finishPatchedCommit(p, state)

	return nil
}

// failBatch rolls a refused batch back and wraps the cause in
// ErrPatchRefused, so callers fall back to a refresh.
func failBatch(p *Page, err error) error {
	rollbackPatch(p)

	return patch.Refuse(ErrPatchRefused, err)
}

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

var errHoverSharedPaint = errors.New("page: hover patch scope shares paint with a sibling")

const maxHoverFailureKeys = 256

// layoutStaged lays out the styled document for one redraw. A change that
// only moves the hover re-places the subtrees the new hover state can reach
// instead of laying the whole page out again. Anything else, and any patch
// the engine refuses, lays the page out exactly as before, so a page the
// patch cannot represent draws what a full layout would draw.
func layoutStaged(p *Page, ctx context.Context,
	styled *css.Document,
	state render.State) (*Display, error) {
	hoverChanged := hoverPatchEligible(p, state) && p.last.Hover != state.Hover
	var scope hoverScopeResult
	if hoverChanged {
		scope = hoverPatchRoots(p, styled, state)
	}
	var checkReflow bool
	var refusedGeoRoots []string
	if scope.ok && (!scope.reflow || !stableHoverGeometry(p)) {
		patched, err := listHoverPatched(p, ctx, styled, state, scope.roots)
		if err == nil {
			p.stats.hoverPatches++
			return patched, nil
		}
		p.stats.hoverPatchErrors++
		noteHoverPatchFailure(p, "selector", classifyHoverPatchError(err), hoverFailureTarget(err, scope.roots))
		checkReflow = stableHoverGeometry(p)
		// If the cascade roots failed (e.g. an element transitioned to display:none
		// and placed no box), fall back to geometric roots only when the sheets
		// contain no escaping selector at all. The geometric roots cannot prove
		// which scope a :has or sibling-after-:hover rule belongs to.
		if !hoverRiskAny(p) {
			geoRoots, geoOK := draw.HoverPatchRoots(p.boxes, p.last.Hover, state.Hover)
			if geoOK && !slices.Equal(geoRoots, scope.roots) {
				patched, retryErr := listHoverPatched(p, ctx, styled, state, geoRoots)
				if retryErr == nil {
					p.stats.hoverPatches++
					return patched, nil
				}
				refusedGeoRoots = geoRoots
				p.stats.hoverPatchErrors++
				noteHoverPatchFailure(p, "geometry", classifyHoverPatchError(retryErr), hoverFailureTarget(retryErr, geoRoots))
			}
		}
	}

	if hoverChanged {
		p.stats.hoverFallbacks++
		if !scope.ok {
			reason := "no-patch-roots"
			if hoverIsRisky(p, p.last.Hover, state.Hover) {
				reason = "escaping-hover-selector"
			}
			noteHoverPatchFailure(p, "selection", reason, []string{p.last.Hover, state.Hover})
		}
	}
	p.commitRoots = nil
	p.stats.fullLayouts++
	display, err := draw.List(ctx, styled, state)
	previous := p.draw.Display
	if err == nil && checkReflow && hoverRootReflowed(previous, display, scope.roots) &&
		(len(refusedGeoRoots) == 0 || hoverRootReflowed(previous, display, refusedGeoRoots)) {
		scope.reflow = true
		p.hoverScopes[hoverScopePair{p.last.Hover, state.Hover}] = scope
	}
	return display, err
}

// stableHoverGeometry excludes paint or geometry mutations driven by ticks.
func stableHoverGeometry(p *Page) bool {
	noFrameCallback := !p.tickActive && !p.hasFn.Load()

	return noFrameCallback && motionCommitIdle(p)
}

// hoverRootReflowed checks a refused root against the successful full layout.
// The root or its immediate parent must have moved or changed size before
// this hover pair can skip retrying the patch under the same layout key.
func hoverRootReflowed(previous, next *Display, roots []string) bool {
	if previous == nil || next == nil {
		return false
	}
	for _, root := range roots {
		var old, newBox *Box
		for i := range previous.Boxes {
			if previous.Boxes[i].ID == root {
				old = &previous.Boxes[i]
				break
			}
		}
		for i := range next.Boxes {
			if next.Boxes[i].ID == root {
				newBox = &next.Boxes[i]
				break
			}
		}
		if old == nil || newBox == nil {
			continue
		}
		if hoverBoxMoved(old, newBox) {
			return true
		}
		oldIndex, newIndex := old.ParentIndex, newBox.ParentIndex
		if oldIndex < 0 || oldIndex >= len(previous.Boxes) || newIndex < 0 || newIndex >= len(next.Boxes) {
			continue
		}
		oldParent, newParent := &previous.Boxes[oldIndex], &next.Boxes[newIndex]
		sameParent := oldParent.ID == newParent.ID && oldParent.Tag == newParent.Tag
		identifiedParent := oldParent.ID != "" || oldIndex == newIndex
		if sameParent && identifiedParent && hoverBoxMoved(oldParent, newParent) {
			return true
		}
	}
	return false
}

func hoverBoxMoved(old, next *Box) bool {
	return old.X != next.X || old.Y != next.Y || old.W != next.W || old.H != next.H
}

func classifyHoverPatchError(err error) string {
	switch {
	case errors.Is(err, errHoverPatchEmpty):
		return "empty-display"
	case errors.Is(err, errHoverSharedPaint):
		return "shared-sibling-paint"
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

func noteHoverPatchFailure(p *Page, stage, reason string, targets []string) {
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

func hoverPatchFailureSnapshot(p *Page) []host.HoverPatchFailureCount {
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

func hoverPatchRoots(p *Page, styled *css.Document, state render.State) hoverScopeResult {
	guardHit := p.sourceRev != 0 &&
		p.hoverGuardSrcRev == p.sourceRev &&
		p.hoverGuardSrcHash == p.sourceHash &&
		p.hoverGuardSrcLen == p.sourceLen &&
		p.hoverGuardTheme == p.themeSrc
	if !guardHit {
		if key := hoverGuardHash(p.source, p.themeSrc); key != p.hoverGuardHash {
			p.hoverGuardHash = key
			p.hoverRiskGlobal, p.hoverRiskScopes = draw.HoverPatchRiskyScopes(p.source, p.themeSrc)
		}
		if p.sourceRev != 0 {
			p.hoverGuardSrcRev = p.sourceRev
			p.hoverGuardSrcHash = p.sourceHash
			p.hoverGuardSrcLen = p.sourceLen
			p.hoverGuardTheme = p.themeSrc
		}
	}

	key := keyFor(p, state)
	key.Hover = ""
	if p.hoverScopes == nil || p.hoverScopeKey != key {
		p.hoverScopeKey = key
		p.hoverScopes = make(map[hoverScopePair]hoverScopeResult)
	}
	pair := hoverScopePair{p.last.Hover, state.Hover}
	if cached, found := p.hoverScopes[pair]; found {
		return cached
	}
	roots, ok := findHoverScopes(p, styled, state)
	if len(p.hoverScopes) >= maxHoverScopes {
		clear(p.hoverScopes)
	}
	result := hoverScopeResult{roots: roots, ok: ok}
	p.hoverScopes[pair] = result
	return result
}

type hoverScopePair struct{ previous, current string }
type hoverScopeResult struct {
	roots  []string
	ok     bool
	reflow bool
}

// reuseUnchangedHoverStyle advances the style cache over a known empty scope.
// HoverChangedScopes proved the prior and next hover match the same selectors,
// so both states can share the document. A new pair still recascades before
// finding its scopes, using a document with the new hover state.
func reuseUnchangedHoverStyle(p *Page, state render.State) (*css.Document, bool) {
	if len(p.hoverScopes) == 0 || p.last.Hover == state.Hover {
		return nil, false
	}
	pair := hoverScopePair{p.last.Hover, state.Hover}
	scope := p.hoverScopes[pair]
	if !scope.ok || len(scope.roots) != 0 || !hoverPatchEligible(p, state) {
		return nil, false
	}
	key := keyFor(p, state)
	key.Hover = ""
	if p.hoverScopeKey != key {
		return nil, false
	}

	previous := state
	previous.Hover = pair.previous
	doc, ok := relayedDoc(p, previous)
	if ok {
		p.relayedHover = state.Hover
	}

	return doc, ok
}

func findHoverScopes(p *Page, styled *css.Document, state render.State) ([]string, bool) {
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
		if hoverRiskAny(p) {
			return nil, false
		}
	} else if hoverRiskAny(p) {
		return nil, false
	}
	return draw.HoverPatchRoots(p.boxes, p.last.Hover, state.Hover)
}

// hoverRiskAny reports whether the sheets contain any selector that can style
// outside the hovered subtrees. The geometric fallback uses this page-wide
// answer because it cannot map a selector scope to the nodes it affects.
func hoverRiskAny(p *Page) bool {
	return p.hoverRiskGlobal || len(p.hoverRiskScopes) > 0
}

// hoverIsRisky reports whether an escaping rule (:has or sibling-after-:hover)
// can match when hover moves between oldID and newID.
func hoverIsRisky(p *Page, oldID, newID string) bool {
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

	return elementOrAncestorsMatchScopes(p, oldID, activeScopes) ||
		elementOrAncestorsMatchScopes(p, newID, activeScopes)
}

func elementOrAncestorsMatchScopes(p *Page, id string, scopes []string) bool {
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
		return sourceElementMatchesScopes(p, id, scopes)
	}

	for idx := start; idx >= 0 && idx < len(p.boxes); {
		boxID := p.boxes[idx].ID
		if boxID != "" {
			for _, scope := range scopes {
				if boxID == scope {
					return true
				}
			}
			if sourceElementMatchesScopes(p, boxID, scopes) {
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
func sourceElementMatchesScopes(p *Page, id string, scopes []string) bool {
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
// into, and no retained-state debt or independent caret update is pending.
func hoverPatchEligible(p *Page, state render.State) bool {
	if !p.hasLast || p.draw.Display == nil || p.cache == nil {
		return false
	}

	if !p.retainState || !p.lastRetain || p.restampRedraw {
		return false
	}
	if currentCaretPaint(p) != p.lastCaret {
		return false
	}

	prev, next := p.last, keyFor(p, state)
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
func listHoverPatched(p *Page, ctx context.Context,
	styled *css.Document,
	state render.State,
	roots []string) (*Display, error) {
	current := p.draw.Display
	if len(roots) == 0 {
		// The cascade proved that no selector changed, so there is no
		// subtree to place. An empty commit scope keeps this distinct
		// from a full layout when observing live transitions.
		p.commitRoots = []string{}

		return current, nil
	}
	options := layout.Options{Images: state.Images, RetainState: true}

	if root := sharedHoverPaintRoot(current.Boxes, roots); root != "" {
		return nil, &hoverPatchAttemptError{target: root, err: errHoverSharedPaint}
	}

	if p.hoverPatch == nil {
		patched, err := render.RelayoutElements(ctx, styled, current, roots, options)
		if err != nil {
			p.commitRoots = nil
			return nil, err
		}
		if patched == nil {
			p.commitRoots = nil
			return nil, errHoverPatchEmpty
		}
		p.commitRoots = append([]string{}, roots...)
		return patched, nil
	}

	for _, root := range roots {
		patch, err := p.hoverPatch(ctx, styled, current, root, options)
		if err != nil {
			if current != nil && current != p.draw.Display {
				render.ReleaseDisplay(current)
			}
			p.commitRoots = nil
			return nil, &hoverPatchAttemptError{target: root, err: err}
		}

		if patch == nil || patch.Display == nil {
			if current != nil && current != p.draw.Display {
				render.ReleaseDisplay(current)
			}
			p.commitRoots = nil
			return nil, &hoverPatchAttemptError{target: root, err: errHoverPatchEmpty}
		}

		if current != nil && current != p.draw.Display && current != patch.Display {
			render.ReleaseDisplay(current)
		}
		current = patch.Display
	}

	p.commitRoots = append([]string{}, roots...)

	return current, nil
}

// A subtree splice can replace every operation named by its root. When
// siblings share an operation ID, replacing one can erase the other's paint.
func sharedHoverPaintRoot(boxes []Box, roots []string) string {
	for _, id := range roots {
		index := hoverPaintBoxIndex(boxes, id)
		if index < 0 || len(boxes[index].OperationIDs) == 0 {
			continue
		}
		root := &boxes[index]
		owned := make(map[uint64]struct{}, len(root.OperationIDs))
		for _, operation := range root.OperationIDs {
			owned[operation] = struct{}{}
		}
		for i := range boxes {
			sibling := &boxes[i]
			if i == index || sibling.ParentIndex != root.ParentIndex {
				continue
			}
			for _, operation := range sibling.OperationIDs {
				if _, ok := owned[operation]; ok {
					return id
				}
			}
		}
	}

	return ""
}

func hoverPaintBoxIndex(boxes []Box, id string) int {
	for i := range boxes {
		if boxes[i].ID == id {
			return i
		}
	}

	return -1
}
