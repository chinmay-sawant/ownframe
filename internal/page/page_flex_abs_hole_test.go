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

// TestFlexChildAbsoluteStaysInFlow records the blinkless hole:
// position:absolute on a flex child is ignored. The child stays in flow
// and top/left do nothing. A block container is the app workaround.
func TestFlexChildAbsoluteStaysInFlow(t *testing.T) {
	t.Parallel()

	p := newViewportPage(t, flexAbsSource("display:flex"), 200, 100)
	abs := viewportBox(t, p, "abs")
	sib := viewportBox(t, p, "sib")
	if abs.X != 0 || abs.Y != 0 {
		t.Fatalf("flex abs = %.0f,%.0f, want in-flow 0,0", abs.X, abs.Y)
	}
	if sib.X != 20 {
		t.Fatalf("flex sib X = %.0f, want 20 (abs kept in flow)", sib.X)
	}

	b := newViewportPage(t, flexAbsSource("position:relative;display:block"), 200, 100)
	abs = viewportBox(t, b, "abs")
	sib = viewportBox(t, b, "sib")
	if abs.X != 40 || abs.Y != 10 {
		t.Fatalf("block abs = %.0f,%.0f, want 40,10", abs.X, abs.Y)
	}
	if sib.X != 0 {
		t.Fatalf("block sib X = %.0f, want 0 (abs out of flow)", sib.X)
	}
}
