package page

// Tests for the batch patch entry point in patch_batch.go:
// ElementPatch{ID, Edits} and PatchElements(ctx, patches) apply N tree
// edits with one shared style re-resolve and a single splice+draw, and a
// refusal wraps ErrPatchRefused.

import (
	"context"
	"errors"
	"testing"
)

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
