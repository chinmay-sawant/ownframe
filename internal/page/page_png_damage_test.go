package page

import (
	"bytes"
	"context"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestRetainedPaintInvalidatesEncodedPNG(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		edit func(*testing.T, *Page, int)
	}{
		{"operation patch", func(t *testing.T, p *Page, i int) {
			t.Helper()
			p.PatchOp(i, func(op *DisplayOp) { op.R, op.G = 0, 1 })
		}},
		{"display patch", func(t *testing.T, p *Page, i int) {
			t.Helper()
			p.PatchDisplay(func(d *Display) { d.Ops[i].R, d.Ops[i].G = 0, 1 })
		}},
		{"tick", func(t *testing.T, p *Page, i int) {
			t.Helper()
			p.SetTick(func(context.Context) error {
				p.Display().Ops[i].R, p.Display().Ops[i].G = 0, 1
				return nil
			})
			if err := p.Tick(context.Background()); err != nil {
				t.Fatal(err)
			}
		}},
		{"explicit invalidation", func(t *testing.T, p *Page, i int) {
			t.Helper()
			p.Display().Ops[i].R, p.Display().Ops[i].G = 0, 1
			p.Invalidate("swatch")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, i := pngSwatch(t)
			before := p.PNG()
			generation := p.Generation()
			tc.edit(t, p, i)
			after := p.PNG()
			if bytes.Equal(before, after) {
				t.Fatal("retained edit returned the old encoded picture")
			}
			img, err := png.Decode(bytes.NewReader(after))
			if err != nil {
				t.Fatal(err)
			}
			r, g, b, _ := img.At(20, 20).RGBA()
			if r != 0 || g != 0xffff || b != 0 {
				t.Fatalf("pixel = %x,%x,%x, want green", r, g, b)
			}
			if p.Generation() != generation {
				t.Fatal("retained edit unexpectedly needed a redraw")
			}
		})
	}
}

func TestNoopRetainedPaintKeepsEncodedPNG(t *testing.T) {
	t.Parallel()
	p, _ := pngSwatch(t)
	p.PNG()
	cached := &p.draw.PNG[0]
	p.PatchDisplay(func(*Display) {})
	p.SetTick(func(context.Context) error { return nil })
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(p.draw.PNG) == 0 || &p.draw.PNG[0] != cached {
		t.Fatal("no-op paint discarded the encoded cache")
	}
}

func pngSwatch(t *testing.T) (*Page, int) {
	t.Helper()
	p, err := New(Config{HTML: `<body style="margin:0"><div id="swatch" style="width:40px;height:40px;background:red"></div></body>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	for i, op := range p.Display().Ops {
		if op.Kind == layout.DisplayOpFillRect && op.R == 1 && op.G == 0 && op.B == 0 {
			return p, i
		}
	}
	t.Fatal("fixture has no red operation")
	return nil, 0
}
