package page

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

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
			box := p.boxByID(id)
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
