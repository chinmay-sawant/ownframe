package page

import (
	"runtime"
	"testing"
	"weak"
)

func TestReplacedDisplayIsCollectible(t *testing.T) {
	p, err := New(Config{HTML: `<div id="a">one</div>`, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	p.SetAnimations(false)
	p.SetTransitions(false)
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	probe := func() weak.Pointer[Display] { return weak.Make(p.draw.Display) }()
	if err := ReplaceHTML(t.Context(), p, `<div id="b">two</div>`); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()
	for range 3 {
		runtime.GC()
	}
	if probe.Value() != nil {
		t.Fatal("replacing a page retained its old display")
	}
	runtime.KeepAlive(p)
}
