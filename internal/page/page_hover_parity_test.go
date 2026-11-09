package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

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
