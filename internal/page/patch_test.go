package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/ownframe/internal/state"
	"slices"
	"sync"
	"testing"
)

// TestPatchElementRefreshesTheAccessibilityTree checks that the tree
// reports the text the patch set.
func TestPatchElementRefreshesTheAccessibilityTree(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<p id="p">old</p>`)

	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	if got := accessibilityTestNode(t, tree, "id:p").Name; got != "old" {
		t.Fatalf("name %q", got)
	}

	text := "new"
	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Text: &text}))

	tree, err = p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	node := accessibilityTestNode(t, tree, "id:p")
	if node.Name != "new" {
		t.Fatalf("name %q, want new", node.Name)
	}
	if tree.Document == 0 {
		t.Fatal("the tree has no document")
	}
}

// TestPatchElementClearsTheCachedPicture checks that the page does not serve
// the picture from before the patch.
func TestPatchElementClearsTheCachedPicture(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#p{width:40px;height:20px;background:#112233}</style><p id="p">x</p>`)

	if len(p.PNG()) == 0 {
		t.Fatal("the page has no picture")
	}

	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "style",
		Value: "width:40px;height:20px;background:#445566"}))
	if p.draw.PNG != nil {
		t.Fatal("the patch kept the cached picture")
	}
}

// TestPatchElementDirtiesTheChangedInk checks that the damage after a patch
// covers the operations that changed, not only the element's own box.
func TestPatchElementDirtiesTheChangedInk(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#d.on{font-size:20px}</style>`+
		`<div id="d" style="width:20px;height:20px;white-space:nowrap">hello world</div>`)

	p.TakeDirty()
	prev := p.draw.Display
	mustPatch(t, p.PatchElement(ctx, "d", ElementEdit{Attribute: "class", Value: "on"}))

	d, full, changed := diffDisplay(prev, p.draw.Display)
	if full || d.n == 0 || changed == 0 {
		t.Fatalf("the patch changed nothing: full %v n %d changed %d", full, d.n, changed)
	}

	rect, ok := p.TakeDirty()
	if !ok {
		t.Fatal("the patch left no dirty region")
	}
	for i := 0; i < d.n; i++ {
		if rect.Union(d.rects[i]) != rect {
			t.Fatalf("the dirty region %v misses the changed ink %v", rect, d.rects[i])
		}
	}

	box := p.boxByID("d")
	if rect.Dx() <= int(box.W) {
		t.Fatalf("the dirty region %v is the element's box %v", rect, box)
	}
}

// TestPatchElementRebuildsTheHitList checks that a patched element still
// takes the click.
func TestPatchElementRebuildsTheHitList(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#p{width:100px;height:40px}</style><p id="p">text</p>`)

	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}))

	box := p.boxByID("p")
	box, ok := p.hitBox(box.X+box.W/2, box.Y+box.H/2)
	if !ok || box.ID != "p" {
		t.Fatalf("the patched element does not take the click: %+v %v", box, ok)
	}
}

// TestPatchElementEditsOneBox checks that an element patch edits the cached
// tree, copies the element's operations into the retained display, and marks
// the element's region dirty.
func TestPatchElementEditsOneBox(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML: `<div id="box" class="a">hello</div><div id="other">world</div>`,
		Theme: `#box{width:80px;height:40px;background:#112233}` +
			`#box.on{background:#445566}`,
		Width:       300,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Repaints
	if p.boxByID("box").ID == "" {
		t.Fatal("no box")
	}
	if p.draw.Display == nil {
		t.Fatal("the page has no display list")
	}
	if fill := p.boxFill("box"); fill == nil {
		t.Fatal("the box owns no fill")
	} else if fill.R > 0.2 {
		t.Fatalf("the fill starts at %.2f, want the base colour", fill.R)
	}

	mustPatch(t, p.PatchElement(ctx, "box", ElementEdit{Attribute: "class", Value: "a on"}))
	if p.Stats().Repaints != before+1 {
		t.Fatalf("repaints = %d, want one more", p.Stats().Repaints)
	}
	fill := p.boxFill("box")
	if fill == nil {
		t.Fatal("the box owns no fill")
	}
	if fill.R < 0.2 {
		t.Fatalf("the fill is still %.2f, want the patched colour", fill.R)
	}
	if fill.R < 0.2 {
		t.Fatalf("the fill is still %.2f, want the patched colour", fill.R)
	}
}

// boxFill returns the first fill operation the box owns.
func (p *Page) boxFill(id string) *DisplayOp {
	for _, b := range p.draw.Display.Boxes {
		if b.ID != id {
			continue
		}
		for i := range p.draw.Display.Ops {
			for _, own := range b.OperationIDs {
				if p.draw.Display.Ops[i].ID == own && p.draw.Display.Ops[i].Kind == DisplayOpFillRect {
					return &p.draw.Display.Ops[i]
				}
			}
		}
	}
	return nil
}

// TestPatchElementRefusesWithoutRetainedState checks that a page without
// the retained layout state refuses a patch.
func TestPatchElementRefusesWithoutRetainedState(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<p id="p">x</p>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	mustRefuse(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}))
}

// TestPatchElementRefusesAnEditThatChangesNothing checks that a patch with
// no edit, one without a name, and one on an element the tree does not hold
// are refused.
func TestPatchElementRefusesAnEditThatChangesNothing(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<p id="p">x</p>`)

	mustRefuse(t, p.PatchElement(ctx, "p"))
	mustRefuse(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "", Value: "on"}))
	mustRefuse(t, p.PatchElement(ctx, "missing", ElementEdit{Attribute: "class", Value: "on"}))
}

// TestPatchElementFollowsTheCacheStateRules checks that the page reads the
// state rules of the cached sheets.
func TestPatchElementFollowsTheCacheStateRules(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#p{width:10px;height:10px}</style><p id="p">x</p>`)

	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}))
	if p.statePseudos != p.cache.StatePseudos() {
		t.Fatalf("the page reports %+v and the cache %+v", p.statePseudos, p.cache.StatePseudos())
	}
}

// patchPage builds a page that keeps the layout state an element patch needs.
func patchPage(t *testing.T, source string) *Page {
	t.Helper()

	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func mustPatch(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func mustRefuse(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("err = %v, want ErrPatchRefused", err)
	}
}

// TestPageChangedPropsConcurrent covers RACE-1: store listeners fire on
// arbitrary goroutines while the page loop drains changedProps metadata.
// Run with -race: the binding must serialize every metadata access.
func TestPageChangedPropsConcurrent(t *testing.T) {
	p, err := New(Config{HTML: `<div>hi</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	st := state.NewStore(map[string]any{"count": 0})
	p.BindState(st)
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				st.TryPatchChange("count", g*1000+i)
				_ = p.ChangedProperties()
				_ = p.TakeChangedProperties()
				p.SyncState()
			}
		}(g)
	}
	wg.Wait()
	if got := p.State().ChangedProperties(); len(got) == 0 {
		t.Fatal("store metadata lost under concurrency")
	}
}

// TestPatchElementRefusedLeavesTheTreeAlone checks that a refused patch
// puts the tree back: the page shows the state it had before the edit.
func TestPatchElementRefusedLeavesTheTreeAlone(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#s{font-family:sans-serif}</style>`+
		`<div id="d" style="width:200px"><span id="s">abc</span></div>`)

	text := "a much longer text that resizes the span"
	mustRefuse(t, p.PatchElement(ctx, "s", ElementEdit{Text: &text}))
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	for _, op := range p.draw.Display.Ops {
		if op.Kind == DisplayOpText && op.W > 60 {
			t.Fatalf("the refused patch changed the text: %+v", op)
		}
	}
}

// TestPatchElementRefusesWithoutACache checks that a page that never drew
// refuses a patch.
func TestPatchElementRefusesWithoutACache(t *testing.T) {
	p, err := New(Config{HTML: `<p id="p">x</p>`, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	mustRefuse(t, p.PatchElement(context.Background(), "p", ElementEdit{Attribute: "class", Value: "on"}))
}

// TestPatchElementRefusesAnEmptyID checks that a patch without an element
// id is refused.
func TestPatchElementRefusesAnEmptyID(t *testing.T) {
	p := patchPage(t, `<p id="p">x</p>`)
	mustRefuse(t, p.PatchElement(context.Background(), "", ElementEdit{Attribute: "class", Value: "on"}))
}

func TestPatchElementRefusesANilPage(t *testing.T) {
	var p *Page
	err := p.PatchElement(context.Background(), "p", ElementEdit{Attribute: "class", Value: "on"})
	if !errors.Is(err, ErrNilPage) {
		t.Fatalf("err = %v, want ErrNilPage", err)
	}
}

func TestPageStateMetadataSurface(t *testing.T) {
	p, err := New(Config{HTML: `<div id=count>1</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	st := state.NewStore(map[string]any{"count": 1})
	p.BindState(st)
	if _, ok := st.TryPatchChange("count", 2); !ok {
		t.Fatal("setPath failed for count")
	}

	if got := p.ChangedProperties(); len(got) == 0 || got[0] != "count" {
		t.Fatalf("page changed = %v", got)
	}
	if got := p.State().ChangedProperties(); len(got) == 0 || got[0] != "count" {
		t.Fatalf("store changed = %v", got)
	}

	p.SetChangedProperties([]string{"score"})
	p.ClearChangedProperties()
	if got := p.ChangedProperties(); !slices.Equal(got, st.ChangedProperties()) {
		t.Fatalf("local clear did not restore the store fallback: %v", got)
	}
	// state store still has count
	if got := p.State().ChangedProperties(); len(got) == 0 {
		t.Fatal("store metadata lost")
	}

	// DisplayPatch surfaces properties.
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()
	p.SetChangedProperties([]string{"count"})
	patch := p.PatchDisplay(func(d *Display) {
		if len(d.Ops) == 0 {
			t.Fatal("no ops")
		}
		d.Ops[0].X += 1
	})
	if len(patch.ChangedProperties) == 0 || patch.ChangedProperties[0] != "count" {
		t.Fatalf("patch props = %+v", patch)
	}

	p.BindState(nil)
	if p.State() != nil {
		t.Fatal("detached page still exposes a store")
	}
	st.TryPatchChange("count", 3)
	if got := p.ChangedProperties(); len(got) != 0 {
		t.Fatalf("detached page receives metadata: %v", got)
	}
}

func TestTakeDrainsLocalProps(t *testing.T) {
	p, err := New(Config{HTML: `<div id="a">A</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	p.SetChangedProperties([]string{"a"})
	first := p.TakeChangedProperties()
	if len(first) != 1 || first[0] != "a" {
		t.Fatalf("first = %v", first)
	}
	if second := p.TakeChangedProperties(); len(second) != 0 {
		t.Fatalf("second = %v, want empty", second)
	}
	p.SetChangedProperties([]string{"b"})
	if got := p.TakeChangedProperties(); len(got) != 1 || got[0] != "b" {
		t.Fatalf("second set = %v", got)
	}
	if got := p.TakeChangedProperties(); len(got) != 0 {
		t.Fatalf("second drain = %v", got)
	}
}

func TestPatchDisplayRecordsDirty(t *testing.T) {
	p, err := New(Config{HTML: `<div id=x>hello</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()

	patch := p.PatchDisplay(func(d *Display) {
		if len(d.Ops) == 0 {
			t.Fatal("no ops")
		}
		d.Ops[0].X += 10
	})

	if patch.Changed != 1 || patch.Full {
		t.Fatalf("patch = %+v", patch)
	}
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("no dirty rects: %v %v", rects, ok)
	}
}

func TestPatchOpRecordsDirty(t *testing.T) {
	p, err := New(Config{HTML: `<div id=x>hello</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !p.PatchOp(0, func(o *DisplayOp) { o.Alpha = 0.5 }) {
		t.Fatal("patch failed")
	}
	if p.LastPatch().Changed != 1 {
		t.Fatalf("patch = %+v", p.LastPatch())
	}
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("no dirty rects: %v %v", rects, ok)
	}
}

func TestPatchDisplayIdentityNoDirty(t *testing.T) {
	p, err := New(Config{HTML: `<div id=x>hello</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()

	patch := p.PatchDisplay(func(*Display) {})
	if patch.Changed != 0 || patch.Full {
		t.Fatalf("patch = %+v", patch)
	}
	rects, ok := p.TakeDirtyRects()
	if ok || len(rects) != 0 {
		t.Fatalf("unexpected dirty: %v %v", rects, ok)
	}
}

// Tests for the batch patch entry point in patch_batch.go:
// ElementPatch{ID, Edits} and PatchElements(ctx, patches) apply N tree
// edits with one shared style re-resolve and a single splice+draw, and a
// refusal wraps ErrPatchRefused.

// batchBoxFill returns the first fill operation the box owns, or nil when
// the box owns none.
func batchBoxFill(p *Page, id string) *DisplayOp {
	for _, b := range p.draw.Display.Boxes {
		if b.ID != id {
			continue
		}
		for i := range p.draw.Display.Ops {
			for _, own := range b.OperationIDs {
				if p.draw.Display.Ops[i].ID == own && p.draw.Display.Ops[i].Kind == DisplayOpFillRect {
					return &p.draw.Display.Ops[i]
				}
			}
		}
	}
	return nil
}

const batchPatchTheme = `#a{width:80px;height:30px;background:#112233}` +
	`#a.on{background:#445566}` +
	`#b{width:80px;height:30px;background:#112233}` +
	`#b.on{background:#445566}` +
	`#c{width:80px;height:30px;background:#112233}` +
	`#c.on{background:#445566}`

const batchPatchHTML = `<div id="a" class="x">one</div>` +
	`<div id="b" class="x">two</div>` +
	`<div id="c" class="x">three</div>`

func batchPatchPage(t *testing.T) *Page {
	t.Helper()

	p, err := New(Config{
		HTML:        batchPatchHTML,
		Theme:       batchPatchTheme,
		Width:       320,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

// batchOnEdit flips one box to its .on colour.
func batchOnEdit() ElementEdit {
	return ElementEdit{Attribute: "class", Value: "x on"}
}

// TestPatchBatchEmptyReturnsNil checks that an empty batch is a no-op: it
// returns nil and draws nothing.
func TestPatchBatchEmptyReturnsNil(t *testing.T) {
	ctx := context.Background()
	p := batchPatchPage(t)
	before := p.Stats()

	if err := p.PatchElements(ctx, nil); err != nil {
		t.Fatalf("PatchElements(nil) = %v, want nil", err)
	}
	if err := p.PatchElements(ctx, []ElementPatch{}); err != nil {
		t.Fatalf("PatchElements(empty) = %v, want nil", err)
	}

	after := p.Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("parses = %d, cascades = %d, want %d and %d",
			after.Parses, after.Cascades, before.Parses, before.Cascades)
	}
	if after.Layouts != before.Layouts || after.Repaints != before.Repaints {
		t.Fatalf("layouts = %d, repaints = %d, want %d and %d",
			after.Layouts, after.Repaints, before.Layouts, before.Repaints)
	}
}

// TestPatchBatchAppliesAllEdits checks that one batch commits every
// element's edits without a reparse.
func TestPatchBatchAppliesAllEdits(t *testing.T) {
	ctx := context.Background()
	p := batchPatchPage(t)

	for _, id := range []string{"a", "b", "c"} {
		fill := batchBoxFill(p, id)
		if fill == nil {
			t.Fatalf("the box %q owns no fill", id)
		}
		if fill.R >= 0.2 {
			t.Fatalf("the fill for %q starts at %.2f, want the base colour", id, fill.R)
		}
	}
	before := p.Stats()

	patches := []ElementPatch{
		{ID: "a", Edits: []ElementEdit{batchOnEdit()}},
		{ID: "b", Edits: []ElementEdit{batchOnEdit()}},
		{ID: "c", Edits: []ElementEdit{batchOnEdit()}},
	}
	if err := p.PatchElements(ctx, patches); err != nil {
		t.Fatalf("PatchElements = %v, want nil", err)
	}

	for _, id := range []string{"a", "b", "c"} {
		fill := batchBoxFill(p, id)
		if fill == nil {
			t.Fatalf("the box %q owns no fill after the batch", id)
		}
		if fill.R < 0.2 {
			t.Fatalf("the fill for %q is still %.2f, want the patched colour", id, fill.R)
		}
	}

	after := p.Stats()
	if after.Parses != before.Parses {
		t.Fatalf("parses = %d, want %d: the batch reparsed", after.Parses, before.Parses)
	}
	if after.Cascades > before.Cascades+1 {
		t.Fatalf("cascades = %d, want at most %d: the batch re-resolved per element",
			after.Cascades, before.Cascades+1)
	}
}

// TestPatchBatchCostsOneResolve checks that a batch shares one style
// re-resolve across its elements: cascades grow by at most one no matter
// how many elements the batch touches, while each element relayout still
// records its own layout.
func TestPatchBatchCostsOneResolve(t *testing.T) {
	ctx := context.Background()
	p := batchPatchPage(t)
	before := p.Stats()

	patches := []ElementPatch{
		{ID: "a", Edits: []ElementEdit{batchOnEdit()}},
		{ID: "b", Edits: []ElementEdit{batchOnEdit()}},
		{ID: "c", Edits: []ElementEdit{batchOnEdit()}},
	}
	if err := p.PatchElements(ctx, patches); err != nil {
		t.Fatalf("PatchElements = %v, want nil", err)
	}

	after := p.Stats()
	if got := after.Cascades - before.Cascades; got > 1 {
		t.Fatalf("cascades grew by %d for %d elements, want at most one shared re-resolve",
			got, len(patches))
	}
	if got := after.Layouts - before.Layouts; got < 1 || got > uint64(len(patches)) {
		t.Fatalf("layouts grew by %d for %d elements, want one draw and at most one layout each",
			got, len(patches))
	}
	if after.Repaints <= before.Repaints {
		t.Fatalf("repaints = %d, want more than %d: the batch drew nothing",
			after.Repaints, before.Repaints)
	}
}

// TestPatchBatchRefusalWrapsErrPatchRefused checks that a batch that cannot
// apply reports ErrPatchRefused: when no element applies, and when one
// patch names no element.
func TestPatchBatchRefusalWrapsErrPatchRefused(t *testing.T) {
	ctx := context.Background()

	p := batchPatchPage(t)
	err := p.PatchElements(ctx, []ElementPatch{
		{ID: "missing", Edits: []ElementEdit{batchOnEdit()}},
	})
	if !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("err = %v, want ErrPatchRefused", err)
	}

	p = batchPatchPage(t)
	err = p.PatchElements(ctx, []ElementPatch{
		{ID: "a", Edits: []ElementEdit{batchOnEdit()}},
		{ID: "", Edits: []ElementEdit{batchOnEdit()}},
	})
	if !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("err = %v, want ErrPatchRefused", err)
	}
}

// TestPatchBatchSkipsAMiss checks that a patch with no change rides along
// without failing the batch: the touched element still commits.
func TestPatchBatchSkipsAMiss(t *testing.T) {
	ctx := context.Background()
	p := batchPatchPage(t)

	err := p.PatchElements(ctx, []ElementPatch{
		{ID: "a", Edits: []ElementEdit{batchOnEdit()}},
		{ID: "missing", Edits: []ElementEdit{batchOnEdit()}},
	})
	if err != nil {
		t.Fatalf("PatchElements = %v, want nil", err)
	}
	if fill := batchBoxFill(p, "a"); fill == nil || fill.R < 0.2 {
		t.Fatalf("the touched box did not commit: %+v", fill)
	}
	if fill := batchBoxFill(p, "b"); fill == nil || fill.R >= 0.2 {
		t.Fatalf("the untouched box changed: %+v", fill)
	}
}

// TestPatchBatchRefusedElementRollsBackTheBatch checks that the batch is
// atomic: when one element's relayout is refused, the whole batch rolls
// back to the source's state and the error wraps ErrPatchRefused. The span
// setup mirrors TestPatchElementRefusedLeavesTheTreeAlone, whose long text
// the engine refuses.
func TestPatchBatchRefusedElementRollsBackTheBatch(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML: `<div id="a" class="x">one</div>` +
			`<div id="d" style="width:200px"><span id="s">abc</span></div>`,
		Theme:       `#a{width:80px;height:30px;background:#112233}#a.on{background:#445566}`,
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

	long := "a much longer text that resizes the span"
	err = p.PatchElements(ctx, []ElementPatch{
		{ID: "a", Edits: []ElementEdit{batchOnEdit()}},
		{ID: "s", Edits: []ElementEdit{{Text: &long}}},
	})
	if !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("err = %v, want ErrPatchRefused", err)
	}
	if fill := batchBoxFill(p, "a"); fill == nil || fill.R >= 0.2 {
		t.Fatalf("the refused batch committed: %+v", fill)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatalf("Redraw after a refused batch = %v", err)
	}
}
