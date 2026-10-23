package page_test

import "testing"

// TestStickyScrollsWithWindow records the blinkless hole: position:sticky
// stays in flow and marks no Fixed ops, so a window scroll translates the
// header with the page. CSS sticky would keep it at viewport 0.
func TestStickyScrollsWithWindow(t *testing.T) {
	t.Parallel()

	src := `<html><head><style>body{margin:0}` +
		`#head{position:sticky;top:0;height:40px;background:#c00}` +
		`#body{height:800px;background:#00c}` +
		`</style></head><body>` +
		`<div id="head"></div><div id="body"></div>` +
		`</body></html>`
	p := newViewportPage(t, src, 200, 100)
	head := viewportBox(t, p, "head")
	if head.Y != 0 || head.H != 40 {
		t.Fatalf("in-flow sticky head = %+v", head)
	}
	d := p.Display()
	if d == nil {
		t.Fatal("no display")
	}
	for _, op := range d.Ops {
		if op.Fixed {
			t.Fatal("sticky op marked Fixed; window would pin it")
		}
	}
	const scrollY = 80.0
	if painted := head.Y - scrollY; painted != -80 {
		t.Fatalf("window-translated Y = %.0f, want -80 not viewport 0", painted)
	}
}
