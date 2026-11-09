package page

import (
	"context"
	"strings"
	"testing"
)

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
