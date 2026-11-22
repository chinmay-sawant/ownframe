package page

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"reflect"
	"strings"
	"testing"
)

func TestHoverHandlerFallbackAndError(t *testing.T) {
	p, err := New(Config{HTML: `<style>#x:hover{width:120px}</style><div id="x" style="height:20px"></div>`, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("paint failed")
	p.Handle(Handlers{Hover: func(context.Context, Box, Box) (bool, error) { return false, boom }})
	before := p.Generation()
	if err = p.Hover(ctx, 10, 10); !errors.Is(err, boom) {
		t.Fatalf("error = %v", err)
	}
	if p.HoverID() != "" || p.Generation() != before {
		t.Fatal("failed hover changed state")
	}
	p.Handle(Handlers{Hover: func(context.Context, Box, Box) (bool, error) { return false, nil }})
	if err = p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if p.HoverID() != "x" || p.Generation() == before {
		t.Fatal("CSS fallback did not redraw")
	}
}

// BenchmarkHoverParseCycle mirrors the inventory app's real session mix: a
// :has theme plus an id-less transitioning control elsewhere on the page,
// hover sweeps across id-bearing cells whose handler patches a detail line,
// and full structural commits that toggle a panel open/closed to force a
// reparse+relayout like search/dialog open/close. One op is a 12-hover sweep
// plus one panel toggle, a 12:1 hover-to-commit mix. The harness reports
// ns/op, B/op, and allocs/op; the timer starts after setup.
func BenchmarkHoverParseCycle(b *testing.B) {
	var bins strings.Builder
	ids := make([]string, 0, 12)
	bins.WriteString(`<div id="grid">`)
	for _, aisle := range []string{"A", "B", "C", "D"} {
		for _, level := range []string{"L1", "L2", "L3"} {
			id := "bin-" + aisle + "-" + level
			ids = append(ids, id)
			fmt.Fprintf(&bins, `<div id="%s" class="inv-bin">1</div>`, id)
		}
	}
	bins.WriteString(`</div>`)
	source := `<html><head><style>` +
		`.inv-bin{width:40px;height:20px}` +
		`.inv-bin:hover{outline:2px solid red}` +
		`.of-button{transition: background-color 0.15s, opacity 0.15s}` +
		`.of-dropdown-menu-row:has(:hover) .of-dropdown-menu-submenu{display:block}` +
		`</style></head><body><button class="of-button">Export</button>` + bins.String() +
		`<div id="bin-detail">summary</div>` +
		`{{if .Open}}<div id="search-panel"><div class="result">result 1</div><div class="result">result 2</div></div>{{end}}` +
		`</body></html>`

	type panelData struct {
		Open bool
	}

	p, err := New(Config{HTML: source, Width: 800, Height: 600, RetainState: true})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	p.SetData(panelData{Open: false})
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	// Same handler shape as the inventory app: patch the detail line, keep
	// the hover redraw.
	p.Handle(Handlers{
		Hover: func(ctx context.Context, _, next Box) (bool, error) {
			text := "hover " + next.ID
			if err := p.PatchElement(ctx, "bin-detail", ElementEdit{Text: &text}); err != nil {
				return false, err
			}
			return false, nil
		},
	})
	b.ReportAllocs()
	b.ResetTimer()
	open := false
	for i := 0; i < b.N; i++ {
		for _, id := range ids {
			box := boxByID(p, id)
			if err := p.Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				b.Fatal(err)
			}
		}
		open = !open
		p.SetData(panelData{Open: open})
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHoverBinCascade measures one inventory-like hover move: a :has
// theme elsewhere on the page, a text patch on the detail line, and the
// hover's own scoped patch. It reports the per-hover cost after the
// single-cascade fix; before it, every hover paid a second css.Relayout.
func BenchmarkHoverBinCascade(b *testing.B) {
	var bins strings.Builder
	ids := make([]string, 0, 12)
	bins.WriteString(`<div id="grid">`)
	for _, aisle := range []string{"A", "B", "C", "D"} {
		for _, level := range []string{"L1", "L2", "L3"} {
			id := "bin-" + aisle + "-" + level
			ids = append(ids, id)
			fmt.Fprintf(&bins, `<div id="%s" class="inv-bin">1</div>`, id)
		}
	}
	bins.WriteString(`</div>`)
	source := `<html><head><style>` +
		`.inv-bin{width:40px;height:20px}` +
		`.inv-bin:hover{outline:2px solid red}` +
		`.of-dropdown-menu-row:has(:hover) .of-dropdown-menu-submenu{display:block}` +
		`</style></head><body>` + bins.String() +
		`<div id="bin-detail">summary</div></body></html>`

	p, err := New(Config{HTML: source, Width: 800, Height: 600, RetainState: true})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	// Same handler shape as the inventory app: patch the detail line, keep
	// the hover redraw.
	p.Handle(Handlers{
		Hover: func(ctx context.Context, _, next Box) (bool, error) {
			text := "hover " + next.ID
			if err := p.PatchElement(ctx, "bin-detail", ElementEdit{Text: &text}); err != nil {
				return false, err
			}
			return false, nil
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box := boxByID(p, ids[i%len(ids)])
		if err := p.Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHoverBinCascadeWithTransition mirrors real inventory: the theme
// carries id-less transitioning buttons (chadcn .of-button), outside every
// bin hover scope. Each hover commit must skip the transition capture walk.
func BenchmarkHoverBinCascadeWithTransition(b *testing.B) {
	var bins strings.Builder
	ids := make([]string, 0, 12)
	bins.WriteString(`<div id="grid">`)
	for _, aisle := range []string{"A", "B", "C", "D"} {
		for _, level := range []string{"L1", "L2", "L3"} {
			id := "bin-" + aisle + "-" + level
			ids = append(ids, id)
			fmt.Fprintf(&bins, `<div id="%s" class="inv-bin">1</div>`, id)
		}
	}
	bins.WriteString(`</div>`)
	source := `<html><head><style>` +
		`.inv-bin{width:40px;height:20px}` +
		`.inv-bin:hover{outline:2px solid red}` +
		`.of-button{transition: background-color 0.15s, opacity 0.15s}` +
		`.of-dropdown-menu-row:has(:hover) .of-dropdown-menu-submenu{display:block}` +
		`</style></head><body><button class="of-button">Export</button>` + bins.String() +
		`<div id="bin-detail">summary</div></body></html>`

	p, err := New(Config{HTML: source, Width: 800, Height: 600, RetainState: true})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	p.Handle(Handlers{
		Hover: func(ctx context.Context, _, next Box) (bool, error) {
			text := "hover " + next.ID
			if err := p.PatchElement(ctx, "bin-detail", ElementEdit{Text: &text}); err != nil {
				return false, err
			}
			return false, nil
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box := boxByID(p, ids[i%len(ids)])
		if err := p.Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			b.Fatal(err)
		}
	}
}

// TestHoverFlowSingleCascadeNoCapture is the agnostic end-to-end pin for the
// hover readout pattern every builder app shares (inventory bins, loglens
// rows): a handler text patch plus the hover's own scoped patch, on a page
// whose theme also carries :has rules and id-less transitioning controls.
//
// It requires, per hover: one cascade total (the handler patch relays it,
// the Redraw reuses it, so StyleTime does not move), no full reparse, no
// transition capture (TransitionsNoID stays 0), no bound run, and exactly
// two scoped commits (patch plus hover patch).
func TestHoverFlowSingleCascadeNoCapture(t *testing.T) {
	source := `<style>.cell{width:40px;height:20px}` +
		`.cell:hover{outline:2px solid red}` +
		`.btn{transition: opacity 0.15s}` +
		`.m:has(:hover) .s{display:block}` +
		`#detail{width:200px;height:20px}</style>` +
		`<div class="m"><div class="s">s</div></div>` +
		`<button class="btn">go</button>` +
		`<div id="a" class="cell">a</div><div id="b" class="cell">b</div>` +
		`<div id="detail">sum</div>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	p.SetPerf(true)
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.Handle(Handlers{
		Hover: func(ctx context.Context, _, next Box) (bool, error) {
			text := "hover " + next.ID
			if err := p.PatchElement(ctx, "detail", ElementEdit{Text: &text}); err != nil {
				return false, err
			}
			return false, nil
		},
	})

	base := p.Stats()
	hoverAt(t, p, "a")
	hoverAt(t, p, "b")
	got := p.Stats()

	if got.Parses != base.Parses {
		t.Fatalf("parses = %d, want %d (a full reparse ran on hover)", got.Parses, base.Parses)
	}
	if got.Layouts != base.Layouts+4 {
		t.Fatalf("layouts = %d, want %d (want one patch plus one hover patch per hover)", got.Layouts, base.Layouts+4)
	}
	// StyleTime is per-Redraw: BeginSample zeroes it, and only a cascade
	// adds to it. The last hover Redraw must have reused the handler
	// patch's document, so it stays exactly zero.
	if got.StyleTime != 0 {
		t.Fatalf("style time = %v on the last hover redraw, want 0 (a second cascade ran)", got.StyleTime)
	}
	if got := p.Motion().TransitionsNoID; got != 0 {
		t.Fatalf("TransitionsNoID = %d, want 0 (capture ran outside every transition)", got)
	}
	if got := p.Motion().TransitionsBound; got != 0 {
		t.Fatalf("TransitionsBound = %d, want 0 (no run should start)", got)
	}
}

func TestHoverSelectorsMatchFullLayout(t *testing.T) {
	for _, test := range []struct{ name, rule, markup string }{
		{"sibling", `#trigger:hover + #outside{background:red}`, `<div id="trigger">hover</div><div id="outside">outside</div>`},
		{"has", `.scope:has(#trigger:hover) + #outside{background:red}`, `<div class="scope"><div id="trigger">hover</div></div><div id="outside">outside</div>`},
		{"hidden-popup", `#popup{display:none}#trigger:hover #popup{display:block}`, `<div id="trigger">hover<div id="popup">popup</div></div>`},
	} {
		for _, refuseFirst := range []bool{false, true} {
			name := test.name
			if refuseFirst {
				name += "-refused"
			}
			t.Run(name, func(t *testing.T) {
				source := `<style>body{margin:0}#trigger,#outside{width:100px;height:40px}` + test.rule + `</style>` + test.markup
				makePage := func() *Page {
					p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
					if err != nil {
						t.Fatal(err)
					}
					p.SetAnimations(false)
					p.SetTransitions(false)
					if err := p.Redraw(t.Context()); err != nil {
						t.Fatal(err)
					}
					return p
				}
				patched, full := makePage(), makePage()
				if refuseFirst {
					attempts := 0
					patched.hoverPatch = func(ctx context.Context, doc *css.Document, prev *Display, id string, opts layout.Options) (*layout.ElementPatch, error) {
						attempts++
						if attempts == 1 {
							return nil, errHoverPatchEmpty
						}
						return render.RelayoutElement(ctx, doc, prev, id, opts)
					}
				}
				// Force the reference to use full placement, including its geometric retry.
				full.hoverPatch = func(context.Context, *css.Document, *Display, string, layout.Options) (*layout.ElementPatch, error) {
					return nil, errHoverPatchEmpty
				}
				for _, target := range []string{"trigger", "", "trigger", ""} {
					for _, p := range []*Page{patched, full} {
						x, y := 250.0, 150.0
						if target != "" {
							x, y = hoverCenter(p, target)
						}
						if err := p.Hover(t.Context(), x, y); err != nil {
							t.Fatal(err)
						}
					}
					assertSameDisplay(t, patched, full)
				}
			})
		}
	}
}

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

func TestHoverWithoutSelectorChangeKeepsSingleCaret(t *testing.T) {
	source := `<style>#styled:hover{color:red}</style>` +
		`<div id="styled">styled</div><div id="plain">plain</div><input id="field" value="text">`
	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	idle := p.PNG()
	if err := p.Focus(t.Context(), "field"); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()
	display := p.Display()
	ops := len(display.Ops)
	commit := p.Generation()
	layer := p.LayerGeneration()
	redraws := 0
	p.SetAfterRedraw(func() { redraws++ })

	hoverAt(t, p, "plain")

	if p.Display() != display || len(display.Ops) != ops {
		t.Fatalf("unchanged hover rebuilt display or changed its %d ops to %d", ops, len(display.Ops))
	}
	if p.LayerGeneration() != layer {
		t.Fatal("unchanged hover invalidated retained layers")
	}
	if p.Generation() != commit+1 || redraws != 1 {
		t.Fatalf("hover must commit and invoke the redraw hook; generation=%d hooks=%d", p.Generation(), redraws)
	}
	if _, ok := p.TakeDirtyRects(); !ok {
		t.Fatal("hover did not retain pending box damage")
	}

	// A batched blur may have the same CSS focus key. That must still clear
	// the caret even when the new hover changes no selector.
	p.BeginRedrawBatch()
	if err := p.Focus(t.Context(), ""); err != nil {
		t.Fatal(err)
	}
	if err := p.Hover(t.Context(), 300, 180); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p.PNG(), idle) {
		t.Fatal("blur left the focused paint on screen")
	}
}

func TestHoverReusesCascadeOnlyForKnownUnchangedScopes(t *testing.T) {
	source := `<style>#styled:hover{background:red}</style>` +
		`<div id="plain">plain</div><div id="styled">styled</div>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	idle := p.PNG()
	leave := func() {
		t.Helper()
		if err := p.Hover(t.Context(), 300, 180); err != nil {
			t.Fatal(err)
		}
	}
	hoverAt(t, p, "plain")
	leave()
	known := p.cache.Styled()
	hoverAt(t, p, "plain")
	leave()
	if p.cache.Styled() != known {
		t.Fatal("a repeated hover with no changed selectors recascaded the document")
	}

	hoverAt(t, p, "styled")
	if p.cache.Styled() == known || bytes.Equal(p.PNG(), idle) {
		t.Fatal("hover with a changed selector reused stale styles")
	}
	leave()
	if !bytes.Equal(p.PNG(), idle) {
		t.Fatal("leaving styled hover did not restore idle paint")
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
	b := boxByID(p, id)

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

func TestHoverSkipsRepeatedGeometryRefusal(t *testing.T) {
	for _, property := range []string{"width:110px", "margin:20px"} {
		t.Run(property, func(t *testing.T) {
			source := `<style>body{margin:0}
				.grid{display:flex;flex-wrap:wrap;width:170px}
				.grid>div{flex:none;width:70px;height:24px;background:red}
				#a:hover{` + property + `}</style>
				<div class="grid"><div id="a">A</div><div id="b">B</div></div>
				<p id="after">After the grid</p>`
			newPage := func(retain bool) *Page {
				p, err := New(Config{HTML: source, Width: 340, Height: 240, RetainState: retain})
				if err != nil {
					t.Fatal(err)
				}
				if err := p.Redraw(t.Context()); err != nil {
					t.Fatal(err)
				}
				return p
			}
			retained, full := newPage(true), newPage(false)
			beforeY := boxByID(full, "after").Y
			cycle := func() {
				hoverAt(t, retained, "a")
				hoverAt(t, full, "a")
				if boxByID(full, "after").Y <= beforeY {
					t.Fatal("fixture did not move the content after the wrapping grid")
				}
				assertSameDisplay(t, retained, full)
				for _, page := range []*Page{retained, full} {
					if err := page.Hover(t.Context(), -1, -1); err != nil {
						t.Fatal(err)
					}
				}
				assertSameDisplay(t, retained, full)
			}
			cycle()
			before := retained.Stats()
			if before.HoverPatchErrors != 2 || before.HoverFallbacks != 2 {
				t.Fatalf("first cycle errors=%d fallbacks=%d, want 2 of each",
					before.HoverPatchErrors, before.HoverFallbacks)
			}
			cycle()
			after := retained.Stats()
			if after.HoverPatchErrors != before.HoverPatchErrors {
				t.Fatalf("repeat attempted %d more patch calls", after.HoverPatchErrors-before.HoverPatchErrors)
			}
			if after.HoverFallbacks != before.HoverFallbacks+2 || after.FullLayouts != before.FullLayouts+2 {
				t.Fatalf("repeat did not preserve the full layout fallback: prior %+v, after %+v", before, after)
			}

			retained.SetTick(func(context.Context) error { return nil })
			cycle()
			withTick := retained.Stats()
			if withTick.HoverPatchErrors != after.HoverPatchErrors+2 {
				t.Fatalf("registered frame callback did not restore patch attempts: prior %d, after %d",
					after.HoverPatchErrors, withTick.HoverPatchErrors)
			}
			retained.SetTick(nil)

			for _, page := range []*Page{retained, full} {
				page.SetSize(380, 240)
				if err := page.Redraw(t.Context()); err != nil {
					t.Fatal(err)
				}
			}
			cycle()
			if got := retained.Stats().HoverPatchErrors; got != withTick.HoverPatchErrors+2 {
				t.Fatalf("viewport change did not invalidate the refusal: prior %d, after %d",
					withTick.HoverPatchErrors, got)
			}
		})
	}
}

func TestHoverSkipsRepeatedParentGeometryRefusal(t *testing.T) {
	source := `<style>body{margin:0}
		#grid{display:flex;flex-wrap:wrap;gap:4px;padding:4px}
		#grid>div{width:84px;height:22px;padding:2px 4px;background:#eee;
			border:1px solid #bbb;font-size:11px;box-sizing:border-box}
		#a:hover{margin-bottom:8px}</style>
		<div id="grid"><div id="a">A</div>` + strings.Repeat("<div>B</div>", 8) +
		`</div><p id="after">After the grid</p>` + strings.Repeat("<p>Outside</p>", 16)
	newPage := func(retain bool) *Page {
		p, err := New(Config{HTML: source, Width: 760, Height: 520, RetainState: retain})
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Redraw(t.Context()); err != nil {
			t.Fatal(err)
		}
		return p
	}
	retained, full := newPage(true), newPage(false)
	initial := boxByID(full, "a")
	initialAfterY := boxByID(full, "after").Y
	cycle := func() {
		hoverAt(t, retained, "a")
		hoverAt(t, full, "a")
		hovered := boxByID(full, "a")
		if initial.X != hovered.X || initial.Y != hovered.Y ||
			initial.W != hovered.W || initial.H != hovered.H || boxByID(full, "after").Y <= initialAfterY {
			t.Fatal("fixture must move content after the grid while keeping the hovered box in place")
		}
		assertSameDisplay(t, retained, full)
		for _, page := range []*Page{retained, full} {
			if err := page.Hover(t.Context(), -1, -1); err != nil {
				t.Fatal(err)
			}
		}
		assertSameDisplay(t, retained, full)
	}
	cycle()
	first := retained.Stats()
	if first.HoverPatchErrors != 4 || first.HoverFallbacks != 2 {
		t.Fatalf("first cycle errors=%d fallbacks=%d, want 4 errors and 2 fallbacks",
			first.HoverPatchErrors, first.HoverFallbacks)
	}
	cycle()
	second := retained.Stats()
	if second.HoverPatchErrors != first.HoverPatchErrors || second.HoverFallbacks != first.HoverFallbacks+2 {
		t.Fatalf("repeat cycle errors=%d fallbacks=%d, want 4 errors and 4 fallbacks",
			second.HoverPatchErrors, second.HoverFallbacks)
	}
}

func TestHoverPatchFailureStatsIncludeStageReasonAndTarget(t *testing.T) {
	p, err := New(Config{
		HTML:  `<style>#e{width:100px;height:20px}#e:hover{color:red}#f{width:100px;height:20px}</style><div id="e">a</div><div id="f">b</div>`,
		Width: 320, Height: 200, RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	p.SetPerf(true)
	hoverAt(t, p, "f")
	p.hoverPatch = func(context.Context, *css.Document, *layout.Display, string, layout.Options) (*layout.ElementPatch, error) {
		return nil, layout.ErrNoElement
	}
	hoverAt(t, p, "e")

	got := p.Stats().HoverPatchFailures
	want := map[string]bool{
		"selector/target-missing/e": false,
		"geometry/target-missing/f": false,
	}
	for _, item := range got {
		key := item.Stage + "/" + item.Reason + "/" + item.Target
		if _, ok := want[key]; ok && item.Count == 1 {
			want[key] = true
		}
	}
	for key, seen := range want {
		if !seen {
			t.Errorf("missing failure count %q in %+v", key, got)
		}
	}
}

func TestHoverPatchFailureCountsStayBounded(t *testing.T) {
	p, err := New(Config{HTML: `<div id="e">a</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	p.SetPerf(true)
	for i := range maxHoverFailureKeys + 16 {
		noteHoverPatchFailure(p, "selector", "target-missing", []string{fmt.Sprintf("target-%03d", i)})
	}

	items := hoverPatchFailureSnapshot(p)
	if len(items) > maxHoverFailureKeys {
		t.Fatalf("failure keys = %d, want at most %d", len(items), maxHoverFailureKeys)
	}
	other := false
	for _, item := range items {
		if item.Target == "<other>" {
			other = true
		}
	}
	if !other {
		t.Fatal("overflow failures were not collapsed into <other>")
	}
}

func TestHoverPatchFunctionalPseudoScopeFallsBack(t *testing.T) {
	source := `<style>` +
		`.card{width:100px;height:20px}` +
		`.details{display:none;width:100px;height:20px}` +
		`.card:not(.disabled):hover + .details{display:block}` +
		`</style><div id="card" class="card">card</div><div class="details">details</div>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	before := p.Stats()
	hoverAt(t, p, "card")
	after := p.Stats()
	if after.HoverPatches != before.HoverPatches {
		t.Fatalf("functional pseudo hover patched: before=%+v after=%+v", before, after)
	}
	if after.FullLayouts == before.FullLayouts {
		t.Fatal("functional pseudo hover did not take a full layout")
	}
}

func TestHoverPatchIDlessHasScopeFallsBack(t *testing.T) {
	source := `<style>` +
		`.row{width:120px;height:40px}` +
		`.row .item{width:100px;height:20px}` +
		`.row .submenu{display:none;width:100px;height:20px}` +
		`.row:has(:hover) .submenu{display:block}` +
		`</style><div class="row"><div id="item" class="item">item</div><div class="submenu">sub</div></div>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	before := p.Stats()
	hoverAt(t, p, "item")
	after := p.Stats()
	if after.HoverPatches != before.HoverPatches {
		t.Fatalf("id-less :has hover patched: before=%+v after=%+v", before, after)
	}
	if after.FullLayouts == before.FullLayouts {
		t.Fatal("id-less :has hover did not take a full layout")
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

// TestHoverPendingFullRecoversChangeSet pins the full-mark recovery: a full
// frame was committed after the host painted, but the host has not taken that
// damage when the hover patch lands. The page diffs against the painted
// display, so the report and the damage describe the hover instead of the
// stale full frame.
func TestHoverPendingFullRecoversChangeSet(t *testing.T) {
	p := newHoverPatchPage(t, true)
	p.TakeDirtyRects()

	markFull(p)
	p.generation++ // a display commit landed after the paint, before this hover

	hoverAt(t, p, "row-b")

	patch := p.LastPatch()
	if patch.Full || patch.Changed > 4 {
		t.Fatalf("patch = %+v, want the hover change only", patch)
	}

	rect, ok := p.TakeDirty()
	if !ok {
		t.Fatal("hover did not dirty")
	}

	if rect.Dx()*rect.Dy() >= p.width*p.height {
		t.Fatalf("damage %v covers the frame", rect)
	}
}

// TestHoverPendingFullWithoutPaintStaysFull pins the other side: before the
// first paint there is no baseline to recover from, so the frame stays full.
func TestHoverPendingFullWithoutPaintStaysFull(t *testing.T) {
	p := newHoverPatchPage(t, true)
	markFull(p)

	hoverAt(t, p, "row-b")

	if patch := p.LastPatch(); !patch.Full {
		t.Fatalf("patch = %+v, want the frame kept full before the first paint", patch)
	}
}

// TestHoverFullDamageKeepsChangedCount pins the perf counters: a full frame
// still reports every operation as dirty, but the changed count is the diff's
// own result rather than the whole display.
func TestHoverFullDamageKeepsChangedCount(t *testing.T) {
	p, err := New(Config{
		HTML:        hoverPatchFixture,
		Width:       320,
		Height:      200,
		RetainState: true,
		Perf:        true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	hoverAt(t, p, "row-a")

	stats := p.Stats()
	if stats.ChangedOps != p.LastPatch().Changed {
		t.Fatalf("changed = %d, patch = %d", stats.ChangedOps, p.LastPatch().Changed)
	}

	if stats.ChangedOps >= stats.DirtyOps {
		t.Fatalf("changed %d >= dirty %d, want the diff count", stats.ChangedOps, stats.DirtyOps)
	}
}

// TestHoverRedrawReusesPatchedCascade pins the single-cascade hover path: a
// hover handler patch relays the cache for the new hover state, so the
// Redraw's styleStaged must reuse that document instead of cascading again.
// Pointer equality proves no second css.Relayout ran.
func TestHoverRedrawReusesPatchedCascade(t *testing.T) {
	source := `<style>.cell{width:40px;height:20px}.cell:hover{outline:2px solid red}#detail{width:200px;height:20px}</style><div id="a" class="cell">a</div><div id="detail">sum</div>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// Mirror the reordered setHover: the id moves first, then the handler
	// patch cascades with the new hover state.
	setHoverID(p, "a")
	text := "Aisle A"
	if err := p.PatchElement(ctx, "detail", ElementEdit{Text: &text}); err != nil {
		t.Fatal(err)
	}
	relayed := p.cache.Styled()
	doc, err := styleStaged(p, ctx, false, p.source, renderState(p))
	if err != nil {
		t.Fatal(err)
	}
	if doc != relayed {
		t.Fatal("Redraw re-cascaded after a handler patch for the same state; want the relayed document")
	}
}

// TestScopedPatchSkipsUnrelatedCapture pins the scope-gated transition
// capture: a patch outside every transitioning element must not run the
// capture walk. The id-less transitioning box would be counted by a
// whole-display capture, so the skipped commit must leave TransitionsNoID
// at 0. The layouts delta proves the patch still committed.
func TestScopedPatchSkipsUnrelatedCapture(t *testing.T) {
	source := `<div id="a" style="width:10px;height:10px">x</div><div style="width:10px;height:10px;transition: opacity 1s">v</div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts
	text := "y"
	if err := p.PatchElement(ctx, "a", ElementEdit{Text: &text}); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want %d (patch did not commit)", got, before+1)
	}
	if got := p.Motion().TransitionsNoID; got != 0 {
		t.Fatalf("TransitionsNoID = %d, want 0 (capture ran outside the scope)", got)
	}
}

// TestScopedPatchBindsInScopeTransition is the control: patching the
// transitioning element itself must run the full capture, bind the run, and
// keep the id-less diagnostic. The scope gate must not over-skip.
func TestScopedPatchBindsInScopeTransition(t *testing.T) {
	source := `<style>.on{opacity:0.5}</style><div id="t" class="off" style="width:10px;height:10px;transition: opacity 1s">v</div><div style="width:10px;height:10px;transition: opacity 1s">u</div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.PatchElement(ctx, "t", ElementEdit{Attribute: "class", Value: "on"}); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().TransitionsBound; got != 1 {
		t.Fatalf("bound transitions = %d, want 1", got)
	}
	if got := p.Motion().TransitionsNoID; got != 1 {
		t.Fatalf("TransitionsNoID = %d, want 1 (diagnostic lost)", got)
	}
}

// TestScopeDeclaresTransitions pins the scope predicate: transitioning boxes
// match by id, by id-bearing ancestor, or not at all. Id-less boxes match
// only through an ancestor in scope, because the cascade anchors hover
// scopes at id-bearing boxes.
func TestScopeDeclaresTransitions(t *testing.T) {
	trans := []layout.Transition{{Property: "opacity"}}
	mk := func(id string, parent int, withTrans bool) Box {
		var tr []layout.Transition
		if withTrans {
			tr = trans
		}
		return Box{ID: id, ParentIndex: parent, Transitions: tr}
	}
	cases := []struct {
		name  string
		boxes []Box
		roots []string
		want  bool
	}{
		{"no transitions", []Box{mk("a", -1, false)}, []string{"a"}, false},
		{"root declares", []Box{mk("a", -1, true)}, []string{"a"}, true},
		{"outside scope", []Box{mk("a", -1, false), mk("b", -1, true)}, []string{"a"}, false},
		{"child of root", []Box{mk("a", -1, false), mk("b", 0, true)}, []string{"a"}, true},
		{"id-less under root", []Box{mk("a", -1, false), mk("", 0, true)}, []string{"a"}, true},
		{"id-less outside scope", []Box{mk("a", -1, false), mk("", -1, true)}, []string{"a"}, false},
		{"nil roots", []Box{mk("a", -1, true)}, nil, false},
		{"empty roots", []Box{mk("a", -1, true)}, []string{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &Display{Boxes: tc.boxes}
			if got := scopeDeclaresTransitions(d, tc.roots); got != tc.want {
				t.Fatalf("scopeDeclaresTransitions = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestHoverPatchWithUnrelatedHas verifies the inventory case: a page whose
// theme uses :has elsewhere still takes the scoped hover patch when the
// hovered element is unrelated to those rules. Before the fix the page-wide
// risky guard forced a full layout on every hover; the cascade scopes are
// precise, so only the geometric fallback needs the guard.
func TestHoverPatchWithUnrelatedHas(t *testing.T) {
	var bins strings.Builder
	bins.WriteString(`<div id="grid">`)
	for _, aisle := range []string{"A", "B", "C", "D"} {
		for _, level := range []string{"L1", "L2", "L3"} {
			id := "bin-" + aisle + "-" + level
			bins.WriteString(`<div id="` + id + `" class="inv-bin">1</div>`)
		}
	}
	bins.WriteString(`</div>`)
	source := `<html><head><style>` +
		`.inv-bin{width:40px;height:20px}` +
		`.inv-bin:hover{outline:2px solid red}` +
		// Unrelated :has rules, like chadcn's dropdown/menubar/tooltip.
		`.of-dropdown-menu-row:has(:hover) .of-dropdown-menu-submenu{display:block}` +
		`.of-tooltip--css-hover:has(:hover)>.x{display:block}` +
		`</style></head><body>` +
		`<div id="menu" class="of-dropdown-menu-row"><div class="of-dropdown-menu-submenu">sub</div></div>` +
		bins.String() +
		`</body></html>`

	p, err := New(Config{HTML: source, Width: 400, Height: 300, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	base := p.Stats().Layouts

	calls := stubHoverPatch(p, true)
	hoverAt(t, p, "bin-A-L1")
	hoverAt(t, p, "bin-B-L2")
	hoverAt(t, p, "bin-C-L3")
	hoverAt(t, p, "bin-D-L1")

	if len(*calls) == 0 {
		t.Fatalf("unrelated :has forced full layouts; want scoped patches")
	}

	// Scoped patches bump layouts once per hover (the splice counts as a
	// layout) but must not parse/cascade again. A full layout per hover
	// would show up as extra parses; scoped patches keep parses at 1.
	st := p.Stats()
	if st.Parses != 1 {
		t.Fatalf("parses = %d, want 1 (full layouts ran on hover)", st.Parses)
	}
	if got := st.Layouts - base; got != 4 {
		t.Fatalf("layouts delta = %d, want 4 scoped patches", got)
	}
}

func TestHoverPatchGlobalHasFallsBack(t *testing.T) {
	source := `<html><head><style>` +
		`body:has(:hover) { background: #fff }` +
		`.btn { width: 40px; height: 20px }` +
		`.btn:hover { background: red }` +
		`</style></head><body>` +
		`<div id="btn-1" class="btn">Click</div>` +
		`</body></html>`

	p, err := New(Config{HTML: source, Width: 400, Height: 300, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	calls := stubHoverPatch(p, true)
	hoverAt(t, p, "btn-1")

	// Global :has rules must prevent partial subtree patching because hovering
	// can affect any part of the body. Full layout must run.
	if len(*calls) != 0 {
		t.Fatalf("global :has allowed scoped patches: %v; want full layout fallback", *calls)
	}
}

func TestHoverPatchScopedHasInsideScopeFallsBack(t *testing.T) {
	source := `<html><head><style>` +
		`.of-dropdown:has(:hover) .submenu { display: block }` +
		`.menu-item { width: 60px; height: 20px }` +
		`.btn { width: 40px; height: 20px }` +
		`.btn:hover { color: red }` +
		`</style></head><body>` +
		`<div id="dd" class="of-dropdown"><div id="item-1" class="menu-item">Item</div></div>` +
		`<div id="btn-1" class="btn">Btn</div>` +
		`</body></html>`

	p, err := New(Config{HTML: source, Width: 400, Height: 300, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// Hovering outside the scope (on btn-1) should be eligible for scoped patch.
	calls := stubHoverPatch(p, true)
	hoverAt(t, p, "btn-1")
	if len(*calls) == 0 {
		t.Fatalf("hover on btn-1 outside risky scope was rejected; want scoped patch")
	}

	// Hovering inside the risky scope (on item-1 inside .of-dropdown) should be flagged as risky.
	// If blinkless HoverChangedScopes doesn't fit or is display:none, it must fall back to full layout
	// instead of using geometric roots that assume no escaping.
	if !hoverIsRisky(p, "", "item-1") {
		t.Fatalf("item-1 inside .of-dropdown should be flagged as hoverIsRisky")
	}
	if hoverIsRisky(p, "", "btn-1") {
		t.Fatalf("btn-1 outside .of-dropdown should NOT be hoverIsRisky")
	}
}

// TestHoverEngineRefusedPatchesMatchFullLayout covers hover rules the engine
// refuses to splice: a rule that resizes the hovered box (width) and a leave
// from a subtree that now holds a float. Each refusal must land on the full
// layout, so every enter and leave draws what a page that never patches draws.
func TestHoverEngineRefusedPatchesMatchFullLayout(t *testing.T) {
	for _, rule := range []string{"width:60px", "float:right"} {
		t.Run(rule, func(t *testing.T) { assertHoverCyclesMatchFull(t, rule) })
	}
}

// TestHoverRootMarginMatchesFullLayout records an engine defect, not an
// ownframe one: Blinkless accepts the patch for `#a:hover{margin-top:8px}`
// on a flex item, keeps the item at its old y, and leaves the container 8px
// short, where a full layout moves the item down and grows the container.
// ownframe cannot see the change (css.Document exposes no computed style and
// the patched root box does not move), so the fix belongs in the engine.
// Remove the skip once the pinned Blinkless splices a root's own margin.
func TestHoverRootMarginMatchesFullLayout(t *testing.T) {
	t.Skip("blocked on Blinkless: RelayoutElement ignores the hovered root's own margin change")
	assertHoverCyclesMatchFull(t, "margin-top:8px")
}

// assertHoverCyclesMatchFull hovers #a and leaves it twice, on a page that
// patches and on one whose patch always fails, and requires equal displays.
func assertHoverCyclesMatchFull(t *testing.T, rule string) {
	t.Helper()

	source := `<style>.g{display:flex;flex-wrap:wrap;gap:4px;padding:4px}` +
		`.i{width:84px;height:22px;box-sizing:border-box;background:#eee}` +
		`#a:hover{` + rule + `}</style><div class="g">` +
		`<div id="a" class="i">a</div><div id="b" class="i">b</div><div id="c" class="i">c</div></div>`
	newPage := func() *Page {
		p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}

		return p
	}
	full := newPage()
	full.hoverPatch = func(context.Context, *css.Document, *layout.Display, string, layout.Options) (*layout.ElementPatch, error) {
		return nil, errHoverPatchEmpty
	}
	patched := newPage()

	for step := range 4 {
		for _, p := range []*Page{patched, full} {
			if step%2 == 0 {
				hoverAt(t, p, "a")
			} else if err := p.Hover(context.Background(), 1, 1); err != nil {
				t.Fatal(err)
			}
		}
		assertSameDisplay(t, patched, full)
	}
}
