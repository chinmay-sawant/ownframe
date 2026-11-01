package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestResponsiveViewportBelowDesktopMinimum(t *testing.T) {
	ctx := context.Background()
	p, err := page.New(page.Config{
		HTML:  `<style>body{margin:0}header{display:flex;justify-content:flex-end;width:100%;height:40px}button{width:40px;height:40px}#compact{display:none}@media(max-width:700px){#compact{display:block}}</style><header><button id="close">X</button></header><div id="compact">Compact</div>`,
		Width: 1120, Height: 760, MinWidth: 900, MinHeight: 610,
	})
	if err != nil {
		t.Fatal(err)
	}
	s := newShell(ctx, p)
	defer s.cancel()
	for _, width := range []int{600, 1200, 600} {
		s.Layout(width, 700)
		s.moving = false
		if err := s.resize(); err != nil {
			t.Fatal(err)
		}
		if w, h := p.Size(); w != width || h != 700 {
			t.Fatalf("viewport=%dx%d, want %dx700", w, h, width)
		}
		compact := false
		closeFound := false
		for _, b := range p.Boxes() {
			if b.ID == "compact" && b.W > 0 && b.H > 0 {
				compact = true
			}
			if b.ID == "close" {
				closeFound = true
				if b.X+b.W > float64(width) {
					t.Fatalf("close extends past viewport: %+v", b)
				}
			}
		}
		if !closeFound {
			t.Fatal("missing close control")
		}
		if compact != (width <= 700) {
			t.Fatalf("width=%d compact media query active=%v", width, compact)
		}
	}
	if w, _ := p.MinSize(); w != 900 {
		t.Fatalf("native minimum changed to %d", w)
	}
}
