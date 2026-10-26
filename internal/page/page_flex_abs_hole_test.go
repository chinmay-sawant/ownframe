package page_test

import "testing"

func flexAbsSource(display string) string {
	return `<html><head><style>body{margin:0}` +
		`#row{` + display + `;width:200px;height:80px}` +
		`#abs{position:absolute;top:10px;left:40px;width:20px;height:20px}` +
		`#sib{width:50px;height:50px}` +
		`</style></head><body>` +
		`<div id="row"><div id="abs"></div><div id="sib"></div></div>` +
		`</body></html>`
}

// TestFlexChildAbsoluteOverlays pins that position:absolute takes a flex
// child out of flow: it paints at its static-plus-offset position and the
// sibling lays out as if it were absent. Block containers behave the same.
func TestFlexChildAbsoluteOverlays(t *testing.T) {
	t.Parallel()

	for _, display := range []string{"display:flex", "position:relative;display:block"} {
		p := newViewportPage(t, flexAbsSource(display), 200, 100)
		abs := viewportBox(t, p, "abs")
		sib := viewportBox(t, p, "sib")

		if abs.X != 40 || abs.Y != 10 {
			t.Fatalf("%s abs = %.0f,%.0f, want 40,10", display, abs.X, abs.Y)
		}

		if sib.X != 0 {
			t.Fatalf("%s sib X = %.0f, want 0 (abs out of flow)", display, sib.X)
		}
	}
}
