package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/hit"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"image"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"weak"
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

	// An identical redraw keeps the animated frame (see
	// TestIdenticalRedrawKeepsAnimatedFrame); a real input change commits
	// again and must rebind.
	p.SetSize(220, 100)
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
	setBoxes(p, first)
	setBoxes(p, moved)

	if got := boxByID(p, "a"); got.X != 11 {
		t.Fatalf("boxByID(a) = %+v, want the new rect", got)
	}
	if got := boxByID(p, "dup"); got.X != 13 {
		t.Fatalf("boxByID(dup) = %+v, want the last box wins", got)
	}
	if got := boxByID(p, ""); got.X != 15 {
		t.Fatalf("anonymous lookup = %+v, want the last anonymous box", got)
	}
	if !hasBox(p, "dup") || hasBox(p, "missing") {
		t.Fatal("hasBox did not answer from the repointed index")
	}

	setBoxes(p, []Box{{ID: "fresh", W: 1, H: 1}})
	if hasBox(p, "a") || hasBox(p, "dup") || hasBox(p, "") {
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
	setDisplayBoxes(p, pinnedDisplay(boxes))
	if len(p.hits) != 3 {
		t.Fatalf("page hit list length = %d, want 3", len(p.hits))
	}
	first := &p.hits[0]

	setDisplayBoxes(p, pinnedDisplay(pinnedBoxesFixture()))
	if len(p.hits) != 3 || &p.hits[0] != first {
		t.Fatal("an unchanged recommit recopied the hit list")
	}
	if !p.hits[0].Fixed || !p.hits[1].Fixed || p.hits[2].Fixed {
		t.Fatal("the reused hit list lost its pinned flags")
	}

	moved := pinnedBoxesFixture()
	moved[1].X = 40
	setDisplayBoxes(p, pinnedDisplay(moved))
	if got := boxByID(p, "in"); got.X != 40 {
		t.Fatalf("boxByID(in) = %+v, want the moved rect", got)
	}
	if !p.hits[0].Fixed || !p.hits[1].Fixed || p.hits[2].Fixed {
		t.Fatal("the rebuilt hit list lost its pinned flags")
	}

	setDisplayBoxes(p, &Display{Boxes: moved})
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
	setBoxes(p, []Box{
		{ID: "a", W: 50, H: 50},
		{ID: "b", X: 150, Y: 80, W: 100, H: 100},
	})
	want := contentRect(p)
	if want.Min.X != 0 || want.Min.Y != 0 || want.Max.X != 250 || want.Max.Y != 180 {
		t.Fatalf("contentRect = %v, want (0,0)-(250,180)", want)
	}
	if got := contentRect(p); got != want || !p.contentOK {
		t.Fatalf("second contentRect = %v ok %v, want the cached %v", got, p.contentOK, want)
	}

	setBoxes(p, []Box{
		{ID: "a", W: 50, H: 50},
		{ID: "b", X: 150, Y: 80, W: 100, H: 100},
	})
	if !p.contentOK {
		t.Fatal("an unchanged recommit dropped the cached rect")
	}
	if got := contentRect(p); got != want {
		t.Fatalf("recomputed rect = %v, want the cached %v", got, want)
	}

	setBoxes(p, []Box{
		{ID: "a", W: 50, H: 50},
		{ID: "b", X: 300, Y: 80, W: 100, H: 100},
	})
	if got := contentRect(p); got.Max.X != 400 || got.Max.Y != 180 {
		t.Fatalf("moved contentRect = %v, want max (400,180)", got)
	}

	p.width = 500
	if got := contentRect(p); got.Max.X != 500 || got.Max.Y != 180 {
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
	if got := a11yState(p).Source(); got != p.source {
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

func TestAdoptPatchedSourceKeepsHTMLInStep(t *testing.T) {
	t.Parallel()

	const oldSource = `<div id="x">old</div>`
	const newSource = `<div id="x">new</div>`

	p, err := New(Config{HTML: oldSource, StaticHTML: true, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	doc := p.nav.Doc()

	text := "new"
	if err := p.PatchElement(ctx, "x", ElementEdit{Text: &text}); err != nil {
		t.Fatal(err)
	}
	if err := AdoptPatchedSource(p, newSource); err != nil {
		t.Fatal(err)
	}
	// The adopt path hashes the source once and hands the same value to
	// history and to the source memo, so every recorded hash agrees with
	// a fresh fold of the adopted bytes.
	wantHash := draw.HashString(newSource)
	if got := p.HTML(); got != newSource {
		t.Fatalf("HTML() = %q, want the adopted source", got)
	}
	if got := p.source; got != newSource {
		t.Fatalf("source = %q, want the adopted source", got)
	}
	if got := p.lastSyncBody; got != newSource {
		t.Fatalf("lastSyncBody = %q, want the adopted source", got)
	}
	if got := p.nav.Current().HTML; got != newSource {
		t.Fatalf("nav current = %q, want the adopted source", got)
	}
	if got := p.nav.Current().Hash; got != nav.Hash(newSource) || got != wantHash {
		t.Fatalf("nav hash = %d, want %d", got, wantHash)
	}
	if entry := p.nav.Current(); !entry.Static || entry.Execute {
		t.Fatalf("adopt stored the wrong mode: %+v", entry)
	}
	if got := p.sourceHash; got != wantHash {
		t.Fatalf("source hash = %d, want %d", got, wantHash)
	}
	if got := p.lastBody; got.source != newSource || got.hash != wantHash {
		t.Fatalf("body memo = (%q, %d), want (%q, %d)", got.source, got.hash, newSource, wantHash)
	}
	if p.nav.Doc() != doc {
		t.Fatalf("adopt moved history: doc %d, want %d", p.nav.Doc(), doc)
	}

	// The remembered key covers the adopted source, so the next redraw is
	// the unchanged fast path and does not lay the page out again.
	if got := p.last.SourceHash; got != wantHash || p.last.SourceLen != len(newSource) ||
		p.last.SourceRev != p.sourceRev {
		t.Fatalf("render key source = (%d, %d, %d), want (%d, %d, %d)",
			got, p.last.SourceLen, p.last.SourceRev, wantHash, len(newSource), p.sourceRev)
	}
	if got := p.last; got != keyFor(p, renderState(p)) {
		t.Fatal("render key does not match the adopted source")
	}
	layouts := p.Stats().Layouts
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != layouts {
		t.Fatalf("post-adopt redraw relaid out: %d -> %d", layouts, got)
	}
	if got := p.HTML(); got != newSource {
		t.Fatalf("HTML() after redraw = %q, want the adopted source", got)
	}

	// Adopting the same bytes again keeps the source revision and the
	// render key: the shared hash proves the source did not move.
	rev, key := p.sourceRev, p.last
	if err := AdoptPatchedSource(p, newSource); err != nil {
		t.Fatal(err)
	}
	if p.sourceRev != rev {
		t.Fatalf("re-adopt bumped source revision %d -> %d", rev, p.sourceRev)
	}
	if p.last != key {
		t.Fatal("re-adopt changed the render key for identical bytes")
	}
}

// afterRedrawWideHTML is a box that follows the viewport, so a stale box
// shows a stale width.
const afterRedrawWideHTML = `<style>body { margin: 0; } div { width: 100%; height: 20px; background: #ff0000; }</style><div id="a"></div>`

func TestAfterRedrawSeesTheNewBoxes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawWideHTML)

	var width int
	var box Box
	p.SetAfterRedraw(func() {
		if d := p.Display(); d != nil {
			width = d.Width
		}
		if b, ok := boxWithID(p.Boxes(), "a"); ok {
			box = b
		}
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if width != 200 {
		t.Fatalf("display width = %d, want 200", width)
	}
	if box.W != 200 {
		t.Fatalf("box = %+v, want the 200px box", box)
	}

	p.SetSize(320, 200)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if width != 320 {
		t.Fatalf("display width = %d, want 320", width)
	}
	if box.W != 320 {
		t.Fatalf("box = %+v, want the 320px box", box)
	}
}

func TestAfterRedrawRefusesAReentrantRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<div id="a">{{.}}</div>`, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(0)

	// The callback changes the data, so the queued follow-up still has a
	// frame to commit. A callback that changed nothing would skip it.
	n := 0
	calls := 0
	p.SetAfterRedraw(func() {
		calls++
		n++
		p.SetData(n)
		if err := p.Redraw(ctx); err != nil {
			t.Errorf("Redraw in the callback: %v", err)
		}
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// The callback asks for a draw, the draw runs after the callback
	// returned, and the callback runs again in it. The second request is
	// dropped, so the callback cannot loop.
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestAfterRedrawNilRemovesTheCallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawHTML)

	calls := 0
	p.SetAfterRedraw(func() { calls++ })
	p.SetAfterRedraw(nil)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("calls = %d, want 0", calls)
	}
}

// afterRedrawHTML is one filled box, so the page keeps a display list and
// the commit path runs.
const afterRedrawHTML = `<style>div { width: 40px; height: 20px; background: #ff0000; }</style><div id="a"></div>`

// afterRedrawPage builds a page for the after-redraw tests.
func afterRedrawPage(t *testing.T, html string) *Page {
	t.Helper()

	p, err := New(Config{HTML: html, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

// boxWithID returns the box the id names.
func boxWithID(boxes []Box, id string) (Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return Box{}, false
}

func TestAfterRedrawRunsAfterTheCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawHTML)

	calls := 0
	var display *Display
	var boxes []Box
	p.SetAfterRedraw(func() {
		calls++
		display = p.Display()
		boxes = p.Boxes()
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if display == nil {
		t.Fatal("the callback did not see the new display")
	}
	if _, ok := boxWithID(boxes, "a"); !ok {
		t.Fatalf("boxes = %+v, want the box of a", boxes)
	}
}

func TestAfterRedrawSkipsAnUnchangedRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawHTML)

	calls := 0
	p.SetAfterRedraw(func() { calls++ })

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d after the first redraw, want 1", calls)
	}

	// The same frame again: nothing changed, so the page skips the layout
	// and the callback stays away.
	if err := p.RedrawIfChanged(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d after an unchanged redraw, want 1", calls)
	}
}

func TestAllocDueGap(t *testing.T) {
	t.Parallel()

	p := &Page{}
	start := time.Unix(0, 0)
	if !allocDue(p, start) {
		t.Fatal("first sample")
	}
	if allocDue(p, start.Add(499*time.Millisecond)) {
		t.Fatal("inside the gap")
	}
	if !allocDue(p, start.Add(allocSampleGap)) {
		t.Fatal("at the gap")
	}
}

func TestSanitizeBidiStrips(t *testing.T) {
	t.Parallel()
	got := SanitizeBidi("a\u202eb\u2066c")
	if got != "abc" {
		t.Fatalf("got %q", got)
	}
	if err := CheckIdentifier("alice_01"); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if err := CheckIdentifier("a\u202eb"); err == nil {
		t.Fatal("bidi identifier allowed")
	}
	if !strings.Contains(SanitizeBidi("x"), "x") {
		t.Fatal("plain text changed")
	}
}

func overlapFixture() []Box {
	return []Box{
		{ID: "outer", X: 0, Y: 0, W: 100, H: 100},
		{ID: "inner", X: 10, Y: 10, W: 20, H: 20},
		{ID: "", X: 12, Y: 12, W: 2, H: 2},
	}
}

func overlapPage(t *testing.T) *Page {
	t.Helper()
	p, err := New(Config{HTML: `<div id="outer"><div id="inner"></div></div>`, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestBoxAtUnifiesPointerWinners pins one overlapping fixture where hit,
// boxIDAt, and boxAt agree, then drives Click, Press, and Hover at the same
// point and asserts all three land on the innermost id.
func TestBoxAtUnifiesPointerWinners(t *testing.T) {
	ctx := context.Background()
	p := overlapPage(t)
	setBoxes(p, overlapFixture())

	box, ok := hit.Hit(p.Boxes(), 13, 13)
	if !ok || box.ID != "" {
		t.Fatalf("hit = %+v, %v, want the innermost anonymous box", box, ok)
	}
	if id := boxIDAt(p, 13, 13); id != "inner" {
		t.Fatalf("boxIDAt = %q, want inner", id)
	}
	if spot := boxAt(p, 13, 13); !spot.ok || spot.id != "inner" {
		t.Fatalf("boxAt = %+v, want inner", spot)
	}

	setBoxes(p, overlapFixture())
	if err := p.Press(ctx, 13, 13); err != nil {
		t.Fatal(err)
	}
	if activeID(p) != "inner" {
		t.Fatalf("press = %q, want inner", activeID(p))
	}

	setBoxes(p, overlapFixture())
	if err := p.Hover(ctx, 13, 13); err != nil {
		t.Fatal(err)
	}
	if p.HoverID() != "inner" {
		t.Fatalf("hover = %q, want inner", p.HoverID())
	}

	setBoxes(p, overlapFixture())
	var clicked Box
	p.Handle(Handlers{Click: func(_ context.Context, b Box) error {
		clicked = b
		return nil
	}})
	if err := p.Click(ctx, 13, 13); err != nil {
		t.Fatal(err)
	}
	if clicked.ID != "inner" {
		t.Fatalf("clicked = %+v, want inner", clicked)
	}
}

func TestBoxesSliceAndBoxAt(t *testing.T) {
	p := overlapPage(t)
	fixture := overlapFixture()
	setBoxes(p, fixture)

	slice := p.BoxesSlice()
	if len(slice) != len(fixture) {
		t.Fatalf("len(BoxesSlice) = %d, want %d", len(slice), len(fixture))
	}
	if len(slice) > 0 && &slice[0] != &p.boxes[0] {
		t.Fatal("BoxesSlice did not return direct slice")
	}

	copied := p.Boxes()
	if len(copied) > 0 && &copied[0] == &p.boxes[0] {
		t.Fatal("Boxes() must return a copy, not the direct slice")
	}

	box, ok := p.BoxAt(13, 13)
	if !ok || box.ID != "" {
		t.Fatalf("BoxAt(13, 13) = %+v, %v, want the innermost anonymous box", box, ok)
	}

	box, ok = p.BoxAt(5, 5)
	if !ok || box.ID != "outer" {
		t.Fatalf("BoxAt(5, 5) = %+v, %v, want outer", box, ok)
	}

	_, ok = p.BoxAt(500, 500)
	if ok {
		t.Fatal("BoxAt(500, 500) want false")
	}

	allocs := testing.AllocsPerRun(10, func() {
		_ = p.BoxesSlice()
		_, _ = p.BoxAt(13, 13)
	})
	if allocs != 0 {
		t.Fatalf("BoxesSlice + BoxAt allocated %f times, want 0", allocs)
	}
}

// TestBoxIndexLeavesNoScan puts thousands of boxes in the page, then clears
// the slice a per-lookup scan would walk. Only the id index can answer. It
// must keep the last-box-wins rule and the old answers for anonymous boxes
// and unknown ids, and a later setBoxes must drop stale entries.
func TestBoxIndexLeavesNoScan(t *testing.T) {
	p, err := New(Config{HTML: `<div id="x"></div>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	boxes := make([]Box, 0, 4000)
	for i := 0; i < 4000; i++ {
		boxes = append(boxes, Box{ID: "b" + strconv.Itoa(i), W: 1, H: 1})
	}

	boxes = append(boxes, Box{ID: "dup", X: 1, W: 1, H: 1})
	boxes = append(boxes, Box{ID: "dup", X: 2, W: 1, H: 1})
	setBoxes(p, boxes)
	p.boxes = nil

	if !hasBox(p, "dup") || hasBox(p, "missing") {
		t.Fatal("hasBox did not answer from the index")
	}

	if got := boxByID(p, "dup"); got.X != 2 {
		t.Fatalf("boxByID = %+v, want the last box", got)
	}

	if got := boxByID(p, "b3999"); got.ID != "b3999" {
		t.Fatalf("boxByID = %+v, want b3999", got)
	}

	if got := boxByID(p, "missing"); got.ID != "missing" || got.X != 0 {
		t.Fatalf("boxByID = %+v, want Box{ID: missing}", got)
	}

	setBoxes(p, []Box{
		{ID: "", X: 7, W: 1, H: 1},
		{ID: "k", W: 1, H: 1},
		{ID: "", X: 8, W: 1, H: 1},
	})

	if got := boxByID(p, ""); got.X != 8 {
		t.Fatalf("anonymous lookup = %+v, want the last anonymous box", got)
	}

	if hasBox(p, "dup") {
		t.Fatal("the index kept a stale entry")
	}
}

func TestEmojiPassPaintsImages(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<html><body><p id="m">Good 😂 and 🎉!</p></body></html>`, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(p.source, `src="emoji/1f602"`) || !strings.Contains(p.source, `src="emoji/1f389"`) {
		t.Fatalf("no emoji images: %s", p.source)
	}

	if !strings.Contains(p.source, "data-ownframe-emoji") {
		t.Fatal("no emoji sizing rule")
	}

	imgs := 0
	for i := range p.Display().Ops {
		if p.Display().Ops[i].Kind == layout.DisplayOpImage {
			imgs++
		}
	}

	if imgs != 2 {
		t.Fatalf("image ops = %d, want 2", imgs)
	}
}

func TestEmojiInFieldValue(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="e" type="text" value="hi">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}

	if err := p.Type(ctx, "😂"); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("e"); got != "hi😂" {
		t.Fatalf("value = %q", got)
	}

	if !strings.Contains(p.source, `src="emoji/1f602"`) {
		t.Fatalf("field paints no emoji image: %s", p.source)
	}
}

func TestEmojiImageResolves(t *testing.T) {
	p, err := New(Config{HTML: `<p>x</p>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := imageBytes(p, "emoji/1f602"); err != nil {
		t.Fatalf("bundled emoji: %v", err)
	}

	if _, err := imageBytes(p, "emoji/1fae0"); err == nil {
		t.Fatal("unsupported emoji resolved")
	}
}

func frontPage(t *testing.T, html string) *Page {
	t.Helper()
	p, err := New(Config{HTML: html, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestSelectFrontThreeStates pins the preference all sinks call: display
// wins, image only reports its bounds and fallback, and an undrawn page
// falls back to the frame size.
func TestSelectFrontThreeStates(t *testing.T) {
	front := SelectFront(frontPage(t, `<p>x</p>`))
	if front.Display != nil || front.Image != nil || front.Fallback {
		t.Fatalf("idle = %+v, want neither", front)
	}
	if front.Width != 120 || front.Height != 80 {
		t.Fatalf("idle size = %d x %d, want 120 x 80", front.Width, front.Height)
	}

	drawn := frontPage(t, `<p>hi</p>`)
	if err := drawn.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	front = SelectFront(drawn)
	if front.Display == nil || front.Display != drawn.Display() {
		t.Fatalf("display = %+v, want the page display", front.Display)
	}
	if front.Fallback || front.Image != nil || front.Generation != drawn.Generation() {
		t.Fatalf("display page = %+v, want no fallback", front)
	}
	if d := drawn.Display(); front.Width != d.Width || front.Height != d.Height {
		t.Fatalf("display size = %d x %d", front.Width, front.Height)
	}

	fb := frontPage(t, `<style>div{width:60px;height:40px;background:#123456;`+
		`border-radius:50% / 20%;}</style><div>x</div>`)
	if err := fb.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fb.Display() != nil || fb.Image() == nil {
		t.Fatal("elliptical border did not fall back to a bitmap")
	}
	front = SelectFront(fb)
	if front.Display != nil || !front.Fallback || front.Image == nil {
		t.Fatalf("fallback = %+v, want image only", front)
	}
	b := fb.Image().Bounds()
	if front.Width != b.Dx() || front.Height != b.Dy() {
		t.Fatalf("image size = %d x %d, want %d x %d", front.Width, front.Height, b.Dx(), b.Dy())
	}
}

// kidCacheFlags snapshots the lazy kid index state under the cache mutex.
func kidCacheFlags(t *testing.T, p *Page) (ready, staleRects bool) {
	t.Helper()
	e := lockedBoxes(p, false)
	defer p.boxCache.mu.Unlock()
	if e == nil {
		t.Fatal("page has no box cache")
	}

	return e.kid.ready, e.kid.staleRects
}

// TestKidIndexBuildsOnFirstUse pins the lazy build: a commit only records the
// box list, and the first subtree query builds the parent-to-children index.
func TestKidIndexBuildsOnFirstUse(t *testing.T) {
	p := &Page{}
	setBoxes(p, []Box{{ID: "root", ParentIndex: -1, W: 10, H: 10}})

	if ready, _ := kidCacheFlags(t, p); ready {
		t.Fatal("setBoxes must not build the kid index")
	}

	r, ok := kidRect(p, "root")
	if !ok {
		t.Fatal("kidRect missed the committed box")
	}
	if want := image.Rect(0, 0, 10, 10); r != want {
		t.Fatalf("kidRect = %v, want %v", r, want)
	}

	if ready, _ := kidCacheFlags(t, p); !ready {
		t.Fatal("the first kidRect must build the index")
	}
}

// TestKidIndexUpdatesRectsInPlace keeps a moved subtree on the built index:
// the id sequence and parent links are unchanged, so only the rects snapshot
// is rewritten and a second query answers with the new geometry.
func TestKidIndexUpdatesRectsInPlace(t *testing.T) {
	p := &Page{}
	setBoxes(p, []Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: 0, X: 90, Y: 90, W: 40, H: 40},
	})

	if _, ok := kidRect(p, "card"); !ok {
		t.Fatal("kidRect missed the card")
	}

	setBoxes(p, []Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: 0, X: 200, Y: 200, W: 40, H: 40},
	})

	ready, stale := kidCacheFlags(t, p)
	if !ready || !stale {
		t.Fatalf("same-shape commit kept ready=%v stale=%v, want the index kept and the rects stale", ready, stale)
	}

	r, ok := kidRect(p, "card")
	if !ok {
		t.Fatal("kidRect missed the moved card")
	}
	if want := image.Rect(0, 0, 240, 240); r != want {
		t.Fatalf("kidRect = %v, want the moved subtree %v", r, want)
	}

	if ready, stale := kidCacheFlags(t, p); !ready || stale {
		t.Fatalf("after the rect rewrite ready=%v stale=%v, want a current index", ready, stale)
	}
}

// TestKidIndexRebuildsOnStructuralChange drops the index when a parent link or
// the id sequence changes, so the subtree query follows the new tree.
func TestKidIndexRebuildsOnStructuralChange(t *testing.T) {
	p := &Page{}
	setBoxes(p, []Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: 0, X: 200, Y: 200, W: 40, H: 40},
	})

	if r, ok := kidRect(p, "card"); !ok || r != image.Rect(0, 0, 240, 240) {
		t.Fatalf("built card rect = %v, %v, want the badge inside", r, ok)
	}

	setBoxes(p, []Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: -1, X: 200, Y: 200, W: 40, H: 40},
	})

	if ready, _ := kidCacheFlags(t, p); ready {
		t.Fatal("a parent change must drop the kid index")
	}

	r, ok := kidRect(p, "card")
	if !ok {
		t.Fatal("kidRect missed the recut card")
	}
	if want := image.Rect(0, 0, 100, 100); r != want {
		t.Fatalf("rebuilt card rect = %v, want the badge detached %v", r, want)
	}

	setBoxes(p, []Box{{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100}})

	if ready, _ := kidCacheFlags(t, p); ready {
		t.Fatal("a removed box must drop the kid index")
	}
}

func TestReplacedDisplayIsCollectible(t *testing.T) {
	p, err := New(Config{HTML: `<div id="a">one</div>`, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	p.SetAnimations(false)
	p.SetTransitions(false)
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	probe := func() weak.Pointer[Display] { return weak.Make(p.draw.Display) }()
	if err := ReplaceHTML(t.Context(), p, `<div id="b">two</div>`, false); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()
	for range 3 {
		runtime.GC()
	}
	if probe.Value() != nil {
		t.Fatal("replacing a page retained its old display")
	}
	runtime.KeepAlive(p)
}

// TestDiscardedPageIsCollectible pins the page-owned box cache: a page that
// committed boxes, answered a subtree query, and bound motion must be
// collectible once the caller drops it. A process-wide cache keyed by page
// pointer kept every discarded page reachable.
func TestDiscardedPageIsCollectible(t *testing.T) {
	const pages = 8
	probes := make([]weak.Pointer[Page], 0, pages)
	for range pages {
		p, err := New(Config{
			HTML:        `<style>#a{transition:opacity 1s}</style><div id="a"><span>x</span></div>`,
			Width:       200,
			Height:      100,
			RetainState: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Redraw(t.Context()); err != nil {
			t.Fatal(err)
		}
		if _, ok := kidRect(p, "a"); !ok {
			t.Fatal("kidRect missed a committed box")
		}
		probes = append(probes, weak.Make(p))
	}
	for range 3 {
		runtime.GC()
	}
	for i, probe := range probes {
		if probe.Value() != nil {
			t.Fatalf("page %d stayed reachable after it was discarded", i)
		}
	}
}

// TestBoxCachesArePerPage checks that two live pages keep independent kid
// indexes: a commit on one never answers a subtree query on the other.
func TestBoxCachesArePerPage(t *testing.T) {
	a, b := &Page{}, &Page{}
	setBoxes(a, []Box{{ID: "root", ParentIndex: -1, W: 10, H: 10}})
	setBoxes(b, []Box{{ID: "other", ParentIndex: -1, W: 30, H: 30}})

	if r, ok := kidRect(a, "root"); !ok || r != image.Rect(0, 0, 10, 10) {
		t.Fatalf("a root = %v, %v", r, ok)
	}
	if _, ok := kidRect(a, "other"); ok {
		t.Fatal("page a answered a box committed on page b")
	}
	forgetBoxes(a)
	if _, ok := kidRect(a, "root"); ok {
		t.Fatal("forgetBoxes kept page a's entry")
	}
	if r, ok := kidRect(b, "other"); !ok || r != image.Rect(0, 0, 30, 30) {
		t.Fatalf("forgetting page a dropped page b: %v, %v", r, ok)
	}
}

func TestDisplayReleaseOnCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<div id="box">first</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}

	// Host takes the dirty rects, setting p.painted and allowing maybeReleasePainted to clear it.
	p.TakeDirtyRects()

	// Update content to force a fresh layout.
	if err := p.Load(ctx, `<div id="box">second</div>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	second := p.Display()
	if second == nil {
		t.Fatal("no display after second redraw")
	}
	if second == first {
		t.Fatal("expected second display to differ from first")
	}

	// first should have been released during commit; calling Release again is safe and idempotent.
	render.ReleaseDisplay(first)
}

func TestFallbackDisplayReleaseOnCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML: `<style>div{width:60px;height:40px;background:#123456;border-radius:50% / 20%;mix-blend-mode:hue}</style>` +
			`<div id="box">{{.V}}</div>`,
		Width:       320,
		Height:      400,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(struct{ V string }{V: "a"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.FallbackDisplay()
	if first == nil {
		t.Fatal("no fallback display after the first redraw")
	}
	if patch, err := render.RelayoutElement(ctx, p.cache.Styled(), first, "box", layout.Options{RetainState: true}); err != nil {
		t.Fatalf("first fallback is not patchable before release: %v", err)
	} else if patch != nil && patch.Display != nil && patch.Display != first {
		detachSharedDisplaySlices(patch.Display, first)
		render.ReleaseDisplay(patch.Display)
	}

	p.TakeDirty()
	p.SetData(struct{ V string }{V: "b"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	second := p.FallbackDisplay()
	if second == nil || second == first {
		t.Fatal("expected a new fallback display after the second redraw")
	}
	if _, err := render.RelayoutElement(ctx, p.cache.Styled(), first, "box", layout.Options{RetainState: true}); !errors.Is(err, layout.ErrNoRetainedState) {
		t.Fatalf("replaced fallback is still patchable: %v", err)
	}
}

func TestDisplayReleaseOnCommitWithRetainedState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML:        `<div id="box">first</div>`,
		Width:       320,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}

	// Host takes dirty rects (p.painted now holds first, and maybeReleasePainted returns early because RetainState=true).
	p.TakeDirtyRects()

	// Second commit: previous (first) must still be released even when previous == p.painted.
	if err := p.Load(ctx, `<div id="box">second</div>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	second := p.Display()
	if second == nil || second == first {
		t.Fatal("expected second display to differ from first")
	}

	// Verify first has been released: its result is nil, so RelayoutElement must return ErrNoRetainedState.
	_, relayoutErr := render.RelayoutElement(ctx, p.cache.Styled(), first, "box", layout.Options{})
	if !errors.Is(relayoutErr, layout.ErrNoRetainedState) {
		t.Fatalf("first display not released on commit: got %v, want %v", relayoutErr, layout.ErrNoRetainedState)
	}
}

func TestDisplayReleaseOnPatchElement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML:        `<div id="box" class="a">hello</div>`,
		Width:       320,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after redraw")
	}

	if err := p.PatchElement(ctx, "box", ElementEdit{Attribute: "class", Value: "b"}); err != nil {
		t.Fatal(err)
	}
	second := p.Display()
	if second == nil || second == first {
		t.Fatal("expected second display after patch")
	}

	// Verify first has been released during PatchElement.
	_, relayoutErr := render.RelayoutElement(ctx, p.cache.Styled(), first, "box", layout.Options{})
	if !errors.Is(relayoutErr, layout.ErrNoRetainedState) {
		t.Fatalf("first display not released on PatchElement: got %v, want %v", relayoutErr, layout.ErrNoRetainedState)
	}
}

func TestDisplayReleaseOnPatchElements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML:        `<div id="box1" class="a">one</div><div id="box2" class="a">two</div>`,
		Width:       320,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after redraw")
	}

	err = p.PatchElements(ctx, []ElementPatch{
		{ID: "box1", Edits: []ElementEdit{{Attribute: "class", Value: "b"}}},
		{ID: "box2", Edits: []ElementEdit{{Attribute: "class", Value: "b"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	second := p.Display()
	if second == nil || second == first {
		t.Fatal("expected second display after batch patch")
	}

	// Verify first has been released during PatchElements.
	_, relayoutErr := render.RelayoutElement(ctx, p.cache.Styled(), first, "box1", layout.Options{})
	if !errors.Is(relayoutErr, layout.ErrNoRetainedState) {
		t.Fatalf("first display not released on PatchElements: got %v, want %v", relayoutErr, layout.ErrNoRetainedState)
	}
}
