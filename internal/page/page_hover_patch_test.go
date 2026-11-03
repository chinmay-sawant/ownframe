package page

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestHoverPatchLeavingUnstyledPanel(t *testing.T) {
	source := `<style>#panel{width:180px;height:100px}button:hover{background:red}</style>` +
		`<div id="panel">` + strings.Repeat(`<div>row</div>`, 80) + `</div>` +
		`<nav id="side" style="position:fixed;left:220px;top:0"><button id="button">Go</button></nav>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	calls := stubHoverPatch(p, true)
	if err := p.Hover(t.Context(), 10, 10); err != nil {
		t.Fatal(err)
	}
	hoverAt(t, p, "button")
	if len(*calls) != 1 || (*calls)[0][0] != "button" {
		t.Fatalf("leaving an unstyled panel patched %v; want only button", *calls)
	}
	full, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := full.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	hoverAt(t, full, "button")
	if !bytes.Equal(p.PNG(), full.PNG()) {
		t.Fatal("scoped hover differs from a full render")
	}
}

const hoverPatchFixture = `<html><head><style>` +
	`.row{width:200px;height:30px}` +
	`.row:hover{background-color:#0c1015}` +
	`.row:hover .arrow{transform:translateX(2px)}` +
	`.row:hover::after{content:"*";color:#ffffff}` +
	`</style></head><body><div id="list">` +
	`<div class="row" id="row-a"><span id="price-a">1</span><span class="arrow" id="arrow-a">x</span></div>` +
	`<div class="row" id="row-b"><span id="price-b">2</span><span class="arrow" id="arrow-b">x</span></div>` +
	`</div></body></html>`

func newHoverPatchPage(t *testing.T, retain bool) *Page {
	t.Helper()

	p, err := New(Config{HTML: hoverPatchFixture, Width: 320, Height: 200, RetainState: retain})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func hoverCenter(p *Page, id string) (float64, float64) {
	b := p.boxByID(id)

	return b.X + b.W/2, b.Y + b.H/2
}

func hoverAt(t *testing.T, p *Page, id string) {
	t.Helper()

	x, y := hoverCenter(p, id)
	if err := p.Hover(context.Background(), x, y); err != nil {
		t.Fatal(err)
	}
}

func stubHoverPatch(p *Page, delegate bool) (calls *[][]string) {
	calls = &[][]string{}
	p.hoverPatch = func(
		ctx context.Context,
		doc *css.Document,
		prev *layout.Display,
		elementID string,
		options layout.Options,
	) (*layout.ElementPatch, error) {
		*calls = append(*calls, []string{elementID})
		if !delegate {
			return nil, errHoverPatchEmpty
		}

		return render.RelayoutElement(ctx, doc, prev, elementID, options)
	}

	return calls
}

func assertSameDisplay(t *testing.T, a, b *Page) {
	t.Helper()

	if len(a.boxes) != len(b.boxes) {
		t.Fatalf("box counts patched=%d full=%d", len(a.boxes), len(b.boxes))
	}
	for i := range a.boxes {
		left, right := a.boxes[i], b.boxes[i]
		// A splice assigns fresh operation IDs; geometry and semantics must match.
		left.OperationIDs, right.OperationIDs = nil, nil
		if !reflect.DeepEqual(left, right) {
			t.Fatalf("box %d differs from full layout: patched=%+v full=%+v", i, left, right)
		}
	}
	if !bytes.Equal(a.PNG(), b.PNG()) {
		t.Fatal("patched pixels differ from full layout")
	}

	da, db := a.draw.Display, b.draw.Display
	if da == nil || db == nil {
		t.Fatal("missing display")
	}

	if len(da.Ops) != len(db.Ops) {
		t.Fatalf("ops = %d, want %d", len(da.Ops), len(db.Ops))
	}

	for i := range da.Ops {
		if !dirty.SameOp(&da.Ops[i], &db.Ops[i]) {
			t.Fatalf("op %d differs from the full layout", i)
		}
	}
}

func TestHoverPatchTakenAndMatchesFull(t *testing.T) {
	full := newHoverPatchPage(t, true)
	full.hoverPatch = func(context.Context, *css.Document, *layout.Display, string, layout.Options) (*layout.ElementPatch, error) {
		return nil, errHoverPatchEmpty
	}
	hoverAt(t, full, "row-a")
	hoverAt(t, full, "row-b")

	patched := newHoverPatchPage(t, true)
	calls := stubHoverPatch(patched, true)
	hoverAt(t, patched, "row-a")

	// The cascade identifies the row directly instead of relaying the list.
	if len(*calls) != 1 || (*calls)[0][0] != "row-a" {
		t.Fatalf("entering hover patched %v, want [[row-a]]", *calls)
	}

	hoverAt(t, patched, "row-b")

	if len(*calls) != 3 || (*calls)[1][0] != "row-a" || (*calls)[2][0] != "row-b" {
		t.Fatalf("patch calls = %v, want [[row-a] [row-a] [row-b]]", *calls)
	}

	assertSameDisplay(t, patched, full)
}

func TestHoverPatchGeometryFallsBack(t *testing.T) {
	source := `<html><head><style>` +
		`#e{width:100px;height:20px}` +
		`#e:hover{width:120px}` +
		`#f{width:100px;height:20px}` +
		`</style></head><body><div id="e">a</div><div id="f">b</div></body></html>`

	newGeomPage := func(t *testing.T) *Page {
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

	full := newGeomPage(t)
	hoverAt(t, full, "f")
	hoverAt(t, full, "e")

	patched := newGeomPage(t)
	calls := stubHoverPatch(patched, true)
	hoverAt(t, patched, "f")
	hoverAt(t, patched, "e")

	if len(*calls) != 1 {
		t.Fatalf("geometry hover attempted %d patches, want 1 with the width change refused", len(*calls))
	}

	assertSameDisplay(t, patched, full)

	stats := patched.Stats()
	if stats.Layouts != 3 {
		t.Fatalf("layouts = %d, want 3", stats.Layouts)
	}
}

func TestHoverPatchRiskySkips(t *testing.T) {
	source := `<html><head><style>` +
		`#e{width:100px;height:20px}` +
		`body:has(#e:hover){background-color:#111111}` +
		`</style></head><body><div id="e">a</div><div id="f">b</div></body></html>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	calls := stubHoverPatch(p, true)
	hoverAt(t, p, "e")
	hoverAt(t, p, "f")

	if len(*calls) != 0 {
		t.Fatalf(":has page patched %v, want full layouts", *calls)
	}

	// The risky page still redraws correctly through the full layout.
	if p.Stats().Layouts != 3 {
		t.Fatalf("layouts = %d, want 3", p.Stats().Layouts)
	}
}

func TestHoverScopeCacheFollowsElementEdits(t *testing.T) {
	source := `<style>#target{width:100px;height:30px}.enabled:hover{background:red}</style><div id="target">Target</div><div id="spacer"><p>a</p><p>b</p><p>c</p></div>`
	p, err := New(Config{HTML: source, Width: 320, Height: 240, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	hoverAt(t, p, "target")
	if err = p.Hover(ctx, 310, 230); err != nil {
		t.Fatal(err)
	}
	if err = p.PatchElement(ctx, "target", ElementEdit{Attribute: "class", Value: "enabled"}); err != nil {
		t.Fatal(err)
	}
	hoverAt(t, p, "target")
	full, err := New(Config{HTML: strings.Replace(source, `id="target"`, `id="target" class="enabled"`, 1), Width: 320, Height: 240})
	if err != nil {
		t.Fatal(err)
	}
	if err = full.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	hoverAt(t, full, "target")
	if !bytes.Equal(p.PNG(), full.PNG()) {
		t.Fatal("cached hover scopes ignored the new hover class")
	}
}
