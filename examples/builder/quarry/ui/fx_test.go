package ui

import (
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/ui"
)

// TestShotEffects renders candidate CSS effects so unsupported ones can be
// dropped. It runs only with QUARRY_SHOTS=1.
func TestShotEffects(t *testing.T) {
	if os.Getenv("QUARRY_SHOTS") == "" {
		t.Skip("visual probe")
	}
	r := newRig(t, 700, 300, "")
	box := func(id, _ string) *ui.Element {
		return ui.Div().ID(id).Class("fx " + id).Text(id)
	}
	root := ui.Div().Class("fx-root").Children(
		box("shadow-soft", ""), box("shadow-inset", ""), box("gradient", ""), box("grad-shadow", ""), box("border-top", ""), box("radius-gradient", ""),
	)
	r.must(r.a.page.SetTheme(ThemeCSS(true) + `
.fx-root { display:flex; gap:20px; padding:30px; background-color:#1c1c1c; flex-wrap: wrap; }
.fx { width:100px; height:50px; color:#fff; font-size:11px; display:flex; align-items:center; justify-content:center; border-radius:6px; background-color:#1c212e; }
.shadow-soft { box-shadow: 0 4px 12px rgba(0,0,0,0.6); }
.shadow-inset { box-shadow: inset 0 1px 0 rgba(255,255,255,0.25); }
.gradient { background-image: linear-gradient(180deg, #4cc9fb, #0e8fd0); }
.grad-shadow { background-image: linear-gradient(180deg, #4cc9fb, #0e8fd0); box-shadow: 0 1px 2px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.35); }
.border-top { border-top: 2px solid #3ecf8e; border-left: 3px solid #3ecf8e; }
.radius-gradient { background-image: linear-gradient(135deg, rgba(56,189,248,0.25), rgba(56,189,248,0.02)); border: 1px solid #3ecf8e; }
`))
	r.must(r.a.view.Update(r.ctx, root))
	r.png("fx")
}
