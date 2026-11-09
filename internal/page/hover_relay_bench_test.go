package page

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

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
		box := p.boxByID(ids[i%len(ids)])
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
		box := p.boxByID(ids[i%len(ids)])
		if err := p.Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			b.Fatal(err)
		}
	}
}
