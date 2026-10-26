package page

import (
	"context"
	"reflect"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

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

	if !reflect.DeepEqual(a.boxes, b.boxes) {
		t.Fatal("patched boxes differ from the full layout")
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
	hoverAt(t, full, "row-a")
	hoverAt(t, full, "row-b")

	patched := newHoverPatchPage(t, true)
	calls := stubHoverPatch(patched, true)
	hoverAt(t, patched, "row-a")

	if len(*calls) != 0 {
		t.Fatalf("entering hover patched %v, want full layout", *calls)
	}

	hoverAt(t, patched, "row-b")

	if len(*calls) != 2 || (*calls)[0][0] != "row-a" || (*calls)[1][0] != "row-b" {
		t.Fatalf("patch calls = %v, want [[row-a] [row-b]]", *calls)
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

	if len(*calls) != 2 {
		t.Fatalf("geometry hover attempted %d patches, want 2 with the width change refused", len(*calls))
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
