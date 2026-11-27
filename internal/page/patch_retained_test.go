package page

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func renderParityPNG(t *testing.T, htmlSrc string, w, h int) []byte {
	t.Helper()
	p, err := New(Config{
		HTML:        htmlSrc,
		Width:       w,
		Height:      h,
		RetainState: true,
	})
	if err != nil {
		t.Fatalf("New parity page: %v", err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw parity page: %v", err)
	}
	return p.PNG()
}

func TestPatchElementClassMutationsParity(t *testing.T) {
	ctx := context.Background()
	style := `<style>
		#box { box-sizing: border-box; width: 120px; height: 40px; background: #cccccc; color: #111111; }
		#box.active { background: #2255cc; color: #ffffff; }
		#box.bordered { border: 2px solid #ff0000; }
	</style>`
	initial := `<html><head>` + style + `</head><body><div id="box">hello</div></body></html>`

	p, err := New(Config{HTML: initial, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// AddClass: adds "active".
	if err := p.PatchElement(ctx, "box", ElementEdit{AddClass: "active"}); err != nil {
		t.Fatalf("AddClass active: %v", err)
	}
	got := p.PNG()
	want := renderParityPNG(t, `<html><head>`+style+`</head><body><div id="box" class="active">hello</div></body></html>`, 200, 100)
	if !bytes.Equal(got, want) {
		t.Fatal("AddClass PNG mismatch vs cold Redraw")
	}

	// No-op AddClass on already-present class must be refused with ErrPatchRefused.
	if err := p.PatchElement(ctx, "box", ElementEdit{AddClass: "active"}); !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("duplicate AddClass err = %v, want ErrPatchRefused", err)
	}

	// ToggleClass: adds "bordered", then removes "active".
	if err := p.PatchElement(ctx, "box", ElementEdit{ToggleClass: "bordered"}, ElementEdit{RemoveClass: "active"}); err != nil {
		t.Fatalf("ToggleClass + RemoveClass: %v", err)
	}
	got = p.PNG()
	want = renderParityPNG(t, `<html><head>`+style+`</head><body><div id="box" class="bordered">hello</div></body></html>`, 200, 100)
	if !bytes.Equal(got, want) {
		t.Fatal("ToggleClass + RemoveClass PNG mismatch vs cold Redraw")
	}

	// No-op RemoveClass on absent class must be refused.
	if err := p.PatchElement(ctx, "box", ElementEdit{RemoveClass: "active"}); !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("absent RemoveClass err = %v, want ErrPatchRefused", err)
	}
}

func TestPatchElementRemoveAttributeParity(t *testing.T) {
	ctx := context.Background()
	style := `<style>
		#box { width: 120px; height: 40px; background: #cccccc; }
		#box[data-on="1"] { background: #00aa44; }
	</style>`
	initial := `<html><head>` + style + `</head><body><div id="box" data-on="1">ok</div></body></html>`

	p, err := New(Config{HTML: initial, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.PatchElement(ctx, "box", ElementEdit{RemoveAttribute: "data-on"}); err != nil {
		t.Fatalf("RemoveAttribute: %v", err)
	}
	got := p.PNG()
	want := renderParityPNG(t, `<html><head>`+style+`</head><body><div id="box">ok</div></body></html>`, 200, 100)
	if !bytes.Equal(got, want) {
		t.Fatal("RemoveAttribute PNG mismatch vs cold Redraw")
	}

	// Removing an already-absent attribute must be refused.
	if err := p.PatchElement(ctx, "box", ElementEdit{RemoveAttribute: "data-on"}); !errors.Is(err, ErrPatchRefused) {
		t.Fatalf("second RemoveAttribute err = %v, want ErrPatchRefused", err)
	}
}

func TestPatchElementStructuralMutationsParityAndA11y(t *testing.T) {
	ctx := context.Background()
	style := `<style>
		#list { width: 180px; height: 120px; overflow: hidden; background: #f0f0f0; }
		.row { display: block; width: 160px; height: 24px; background: #ddeeff; }
	</style>`
	initial := `<html><head>` + style + `</head><body><div id="list"><button id="r1" class="row">one</button></div></body></html>`

	p, err := New(Config{HTML: initial, Width: 220, Height: 160, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// Attach accessibility tree before patching so refreshAccessibility mutates it in place.
	if _, err := p.AccessibilityTree(); err != nil {
		t.Fatalf("initial AccessibilityTree: %v", err)
	}

	// AppendHTML adds #r2 inside fixed-size #list.
	if err := p.PatchElement(ctx, "list", ElementEdit{AppendHTML: `<button id="r2" class="row">two</button>`}); err != nil {
		t.Fatalf("AppendHTML: %v", err)
	}
	got := p.PNG()
	want := renderParityPNG(t, `<html><head>`+style+`</head><body><div id="list"><button id="r1" class="row">one</button><button id="r2" class="row">two</button></div></body></html>`, 220, 160)
	if !bytes.Equal(got, want) {
		t.Fatal("AppendHTML PNG mismatch vs cold Redraw")
	}

	// RemoveElement detaches #r1 while #list stays the patch root.
	if err := p.PatchElement(ctx, "list", ElementEdit{ID: "r1", Remove: true}); err != nil {
		t.Fatalf("RemoveElement: %v", err)
	}
	got = p.PNG()
	want = renderParityPNG(t, `<html><head>`+style+`</head><body><div id="list"><button id="r2" class="row">two</button></div></body></html>`, 220, 160)
	if !bytes.Equal(got, want) {
		t.Fatal("RemoveElement PNG mismatch vs cold Redraw")
	}

	// SetInnerHTML replaces #list children with #r3.
	frag := `<button id="r3" class="row">three</button>`
	if err := p.PatchElement(ctx, "list", ElementEdit{InnerHTML: &frag}); err != nil {
		t.Fatalf("SetInnerHTML: %v", err)
	}
	got = p.PNG()
	want = renderParityPNG(t, `<html><head>`+style+`</head><body><div id="list"><button id="r3" class="row">three</button></div></body></html>`, 220, 160)
	if !bytes.Equal(got, want) {
		t.Fatal("SetInnerHTML PNG mismatch vs cold Redraw")
	}

	// Verify the accessibility tree reflects the final retained structure.
	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatalf("final AccessibilityTree: %v", err)
	}
	foundR3 := false
	for _, n := range tree.Nodes {
		if n.ElementID == "r1" || n.ElementID == "r2" {
			t.Fatalf("stale removed node %q still present in AccessibilityTree", n.ElementID)
		}
		if n.ElementID == "r3" {
			foundR3 = true
		}
	}
	if !foundR3 {
		t.Fatal("expected node r3 in AccessibilityTree after SetInnerHTML")
	}
}
