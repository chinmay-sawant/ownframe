package page

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/hit"
)

// assertMotionFullyIdle requires every counter a no-declaration commit would
// zero to already read zero. A static commit that skips the animation,
// capture, and transition walks must leave exactly this state behind.
func assertMotionFullyIdle(t *testing.T, st MotionStats) {
	t.Helper()

	if st.Bound != 0 || st.Active != 0 || st.Sampled != 0 {
		t.Fatalf("animation players = bound %d active %d sampled %d, want all zero",
			st.Bound, st.Active, st.Sampled)
	}
	if st.Rebinds != 0 {
		t.Fatalf("rebinds = %d, want 0", st.Rebinds)
	}
	if st.NeedsRestamp != 0 || st.NeedsRetainedState || st.LastError != "" ||
		len(st.Gaps) != 0 || st.UnresolvedTimelines != 0 || st.InactiveTimelines != 0 {
		t.Fatalf("animation diagnostics = %+v, want zero", st)
	}
	if st.LayoutBound != 0 || st.LayoutQueued != 0 || st.SVGPaintRequests != 0 {
		t.Fatalf("layout/svg work = %+v, want zero", st)
	}
	if st.TransitionsBound != 0 || st.TransitionsStarted != 0 ||
		st.TransitionsCancelled != 0 || st.TransitionsSkipped != 0 ||
		st.TransitionsBlocked != 0 || st.TransitionSkipReason != "" ||
		st.TransitionsNoID != 0 || st.TransitionsNoOld != 0 ||
		st.TransitionsActive != 0 || st.TransitionsSampled != 0 ||
		st.TransitionError != "" {
		t.Fatalf("transition renderers = %+v, want zero", st)
	}
}

// TestCommitStaticMotionStaysIdle commits two static frames and requires the
// motion runtime to stay fully idle: no players bound, no rebinds, no
// transition renderers, and no errors. The generation and layout counters
// prove both commits really published, so the test is not vacuous.
func TestCommitStaticMotionStaysIdle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<div id="a">n={{.}}</div>`, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(1)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	layouts := p.Stats().Layouts
	gen := p.Generation()

	p.SetData(2)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.Generation() == gen {
		t.Fatal("the data change committed nothing")
	}
	if p.Stats().Layouts == layouts {
		t.Fatal("the data change laid nothing out")
	}
	if len(p.Boxes()) == 0 {
		t.Fatal("the commit published no boxes")
	}
	assertMotionFullyIdle(t, p.Motion())
	if _, ok := p.TakeDirty(); !ok {
		t.Fatal("the commit marked no damage")
	}
}

// TestCommitAnimatedMotionStillBinds draws a running animation and requires
// the commit to bind it, then requires the next commit to rebind it. The
// idle fast-out must not swallow a page that declares motion.
func TestCommitAnimatedMotionStillBinds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newMotionSource(t, motionRunning)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Bound; got != 1 {
		t.Fatalf("bound players = %d, want 1", got)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Bound; got != 1 {
		t.Fatalf("bound players after the second commit = %d, want 1", got)
	}
	if got := p.Motion().Rebinds; got == 0 {
		t.Fatal("the second commit did not rebind the players")
	}
}

// TestCommitPatchStartsTransitions drives a class change through the element
// patch path and requires all three transitions to bind. Capture and bind
// must delegate whenever the display declares transitions.
func TestCommitPatchStartsTransitions(t *testing.T) {
	t.Parallel()

	p := newTransitionWirePage(t)
	patchTargetClass(t, p, "on")
	if got := p.Motion().TransitionsBound; got != 3 {
		t.Fatalf("bound transitions = %d, want 3", got)
	}
}

// TestCommitTransitionNoIDCounts requires a capture that finds only an
// id-less transition declaration to record it. The bind must run to publish
// that diagnostic instead of skipping on its empty capture set.
func TestCommitTransitionNoIDCounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML:   `<div style="width:10px;height:10px;transition: opacity 1s">v{{.}}</div>`,
		Width:  200,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(1)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().TransitionsNoID; got != 0 {
		t.Fatalf("first commit TransitionsNoID = %d, want 0", got)
	}

	p.SetData(2)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().TransitionsNoID; got != 1 {
		t.Fatalf("second commit TransitionsNoID = %d, want 1", got)
	}
}

// TestHitIndexSameIDsRepoint recommit the same ids with new rectangles and
// requires the index to answer the new boxes with no stale entries. The
// direct SameIDs and Repoint checks pin the gate the page uses.
func TestHitIndexSameIDsRepoint(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<div id="x"></div>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	first := []Box{
		{ID: "a", X: 1, W: 1, H: 1},
		{ID: "dup", X: 2, W: 1, H: 1},
		{ID: "dup", X: 3, W: 1, H: 1},
		{ID: "", X: 4, W: 1, H: 1},
		{ID: "", X: 5, W: 1, H: 1},
	}
	moved := []Box{
		{ID: "a", X: 11, W: 1, H: 1},
		{ID: "dup", X: 12, W: 1, H: 1},
		{ID: "dup", X: 13, W: 1, H: 1},
		{ID: "", X: 14, W: 1, H: 1},
		{ID: "", X: 15, W: 1, H: 1},
	}
	p.setBoxes(first)
	p.setBoxes(moved)

	if got := p.boxByID("a"); got.X != 11 {
		t.Fatalf("boxByID(a) = %+v, want the new rect", got)
	}
	if got := p.boxByID("dup"); got.X != 13 {
		t.Fatalf("boxByID(dup) = %+v, want the last box wins", got)
	}
	if got := p.boxByID(""); got.X != 15 {
		t.Fatalf("anonymous lookup = %+v, want the last anonymous box", got)
	}
	if !p.hasBox("dup") || p.hasBox("missing") {
		t.Fatal("hasBox did not answer from the repointed index")
	}

	p.setBoxes([]Box{{ID: "fresh", W: 1, H: 1}})
	if p.hasBox("a") || p.hasBox("dup") || p.hasBox("") {
		t.Fatal("the index kept a stale entry after the id set changed")
	}

	if !hit.SameIDs(first, moved) {
		t.Fatal("SameIDs missed an unchanged id set")
	}
	if hit.SameIDs(first, first[:len(first)-1]) {
		t.Fatal("SameIDs missed a length change")
	}
	if hit.SameIDs([]Box{{ID: "a"}}, []Box{{ID: "b"}}) {
		t.Fatal("SameIDs missed an id change")
	}
	ix := hit.Index(nil, first)
	repointed := hit.Repoint(ix, moved)
	if repointed["a"] != &moved[0] {
		t.Fatal("Repoint does not point at the new boxes")
	}
	if repointed["dup"] != &moved[2] {
		t.Fatal("Repoint does not keep last-box-wins")
	}
	if _, ok := repointed["missing"]; ok {
		t.Fatal("Repoint invented an entry")
	}
}

// pinnedBoxesFixture builds one fixed header with a child and one content
// box, with the parent links the fixed walk climbs.
func pinnedBoxesFixture() []Box {
	return []Box{
		{ID: "bar", W: 200, H: 24, ParentIndex: -1},
		{ID: "in", X: 4, Y: 4, W: 40, H: 16, ParentIndex: 0},
		{ID: "body", Y: 100, W: 200, H: 400, ParentIndex: -1},
	}
}

// pinnedDisplay wraps boxes in a display whose first two operations are
// fixed, so the header subtree is the pinned layer.
func pinnedDisplay(boxes []Box) *Display {
	return &Display{
		Boxes:    boxes,
		Ops:      []DisplayOp{{Fixed: true}, {Fixed: true}, {}},
		OpOwners: []int{0, 1, 2},
		Width:    200,
		Height:   600,
	}
}

// TestHitListPinnedAndReuse requires fixed marking to cover a pinned root
// and its child but not the content, and requires an unchanged recommit to
// keep the hit slice instead of recopying it.
func TestHitListPinnedAndReuse(t *testing.T) {
	t.Parallel()

	boxes := pinnedBoxesFixture()
	fixed := []Box{boxes[0]}
	out := hit.List(nil, boxes, fixed)
	if len(out) != 3 {
		t.Fatalf("hit list length = %d, want 3", len(out))
	}
	if !out[0].Fixed || !out[1].Fixed || out[2].Fixed {
		t.Fatalf("pinned flags = %v %v %v, want true true false",
			out[0].Fixed, out[1].Fixed, out[2].Fixed)
	}

	again := hit.List(out, pinnedBoxesFixture(), fixed)
	if len(again) != 3 || &again[0] != &out[0] {
		t.Fatal("the list did not reuse the caller storage")
	}
	if !again[0].Fixed || !again[1].Fixed || again[2].Fixed {
		t.Fatal("the reused list lost its pinned flags")
	}

	p, err := New(Config{HTML: `<div id="bar"></div>`, Width: 200, Height: 600})
	if err != nil {
		t.Fatal(err)
	}
	p.setDisplayBoxes(pinnedDisplay(boxes))
	if len(p.hits) != 3 {
		t.Fatalf("page hit list length = %d, want 3", len(p.hits))
	}
	first := &p.hits[0]

	p.setDisplayBoxes(pinnedDisplay(pinnedBoxesFixture()))
	if len(p.hits) != 3 || &p.hits[0] != first {
		t.Fatal("an unchanged recommit recopied the hit list")
	}
	if !p.hits[0].Fixed || !p.hits[1].Fixed || p.hits[2].Fixed {
		t.Fatal("the reused hit list lost its pinned flags")
	}

	moved := pinnedBoxesFixture()
	moved[1].X = 40
	p.setDisplayBoxes(pinnedDisplay(moved))
	if got := p.boxByID("in"); got.X != 40 {
		t.Fatalf("boxByID(in) = %+v, want the moved rect", got)
	}
	if !p.hits[0].Fixed || !p.hits[1].Fixed || p.hits[2].Fixed {
		t.Fatal("the rebuilt hit list lost its pinned flags")
	}

	p.setDisplayBoxes(&Display{Boxes: moved})
	if len(p.hits) != 0 {
		t.Fatal("a display with no pins kept a hit list")
	}
}

// TestContentRectCacheHit requires the cached rect to cover the canvas and
// every overflowing box, an unchanged recommit to keep the cache, and a
// resize or a moved box to recompute it.
func TestContentRectCacheHit(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<div></div>`, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	p.setBoxes([]Box{
		{ID: "a", W: 50, H: 50},
		{ID: "b", X: 150, Y: 80, W: 100, H: 100},
	})
	want := p.contentRect()
	if want.Min.X != 0 || want.Min.Y != 0 || want.Max.X != 250 || want.Max.Y != 180 {
		t.Fatalf("contentRect = %v, want (0,0)-(250,180)", want)
	}
	if got := p.contentRect(); got != want || !p.contentOK {
		t.Fatalf("second contentRect = %v ok %v, want the cached %v", got, p.contentOK, want)
	}

	p.setBoxes([]Box{
		{ID: "a", W: 50, H: 50},
		{ID: "b", X: 150, Y: 80, W: 100, H: 100},
	})
	if !p.contentOK {
		t.Fatal("an unchanged recommit dropped the cached rect")
	}
	if got := p.contentRect(); got != want {
		t.Fatalf("recomputed rect = %v, want the cached %v", got, want)
	}

	p.setBoxes([]Box{
		{ID: "a", W: 50, H: 50},
		{ID: "b", X: 300, Y: 80, W: 100, H: 100},
	})
	if got := p.contentRect(); got.Max.X != 400 || got.Max.Y != 180 {
		t.Fatalf("moved contentRect = %v, want max (400,180)", got)
	}

	p.width = 500
	if got := p.contentRect(); got.Max.X != 500 || got.Max.Y != 180 {
		t.Fatalf("resized contentRect = %v, want max (500,180)", got)
	}
}

// TestCommitSourceCopiesStayConsistent pins the invariant a future shared
// source handle must preserve: after a commit the stored source, the last
// sync body, and the accessibility source compare equal, and an unchanged
// redraw keeps them.
func TestCommitSourceCopiesStayConsistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<div>hi {{.}}</div>`, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData("a")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.source == "" || p.source != p.lastSyncBody {
		t.Fatal("the stored source and the last sync body disagree")
	}
	if !strings.Contains(p.source, "hi a") {
		t.Fatalf("source = %q, want the rendered data", p.source)
	}
	if got := p.a11yState().Source(); got != p.source {
		t.Fatal("the accessibility source disagrees with the stored source")
	}
	before := p.source

	p.SetData("a")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.source != before {
		t.Fatal("an unchanged redraw replaced the stored source")
	}

	p.SetData("b")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p.source, "hi b") || p.source != p.lastSyncBody {
		t.Fatalf("source = %q, want the new data in every copy", p.source)
	}
}
