package remote

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestPhoneFontScalingReachesRenderedControls(t *testing.T) {
	a := newTest(t, WithPhone(true))
	p := a.Page()
	p.SetSize(240, 320)
	var base float64
	for _, font := range []int{16, 48} {
		a.view.FontSize = font
		a.setData()
		if err := p.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		d := p.Display()
		if d == nil {
			t.Fatal("missing display")
		}
		var width float64
		for _, b := range p.Boxes() {
			if b.ID == "ok" {
				width = b.W * d.PointsPerPixel
			}
		}
		found := false
		for _, op := range d.Ops {
			if op.Kind != ownframe.DisplayOpText || op.Text != "OK" {
				continue
			}
			found = true
			if font == 16 {
				base = op.Size
			} else if op.Size != base*3 {
				t.Fatalf("font size %.1f, want %.1f", op.Size, base*3)
			}
			if op.W > width+.1 {
				t.Fatalf("OK text %.1f exceeds button %.1f", op.W, width)
			}
		}
		if !found {
			t.Fatal("missing OK text")
		}
	}
}
