package ownframe

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"
)

// repaintCase is one page whose incremental path must match a full Redraw.
type repaintCase struct {
	name     string
	newPage  func(t *testing.T) *Page
	interact func(t *testing.T, p *Page, ctx context.Context)
}

// TestPagePictureMatchesRedraw drives one interaction on two pages and
// compares the incremental PNG with the PNG after a full Redraw. This is the
// software-picture consistency guard. Window tile pixels are checked separately
// by make graphics-check.
func TestPagePictureMatchesRedraw(t *testing.T) {
	t.Parallel()

	for _, c := range repaintCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			inc := c.newPage(t)
			full := c.newPage(t)

			before := inc.Probe().PNG()
			c.interact(t, inc, ctx)
			c.interact(t, full, ctx)
			if err := full.Redraw(ctx); err != nil {
				t.Fatal(err)
			}

			after := inc.Probe().PNG()
			if before == nil || after == nil {
				t.Fatal("no PNG to compare")
			}

			beforeImg, err := png.Decode(bytes.NewReader(before))
			if err != nil {
				t.Fatal(err)
			}

			afterImg, err := png.Decode(bytes.NewReader(after))
			if err != nil {
				t.Fatal(err)
			}

			if samePicture(beforeImg, afterImg) {
				t.Fatal("the interaction did not change the picture")
			}

			fullImg, err := png.Decode(bytes.NewReader(full.Probe().PNG()))
			if err != nil {
				t.Fatal(err)
			}

			if !samePicture(afterImg, fullImg) {
				t.Fatal("incremental picture differs from a full Redraw")
			}
		})
	}
}

// newPage builds a page, lets the case set its data and handlers, and lays
// it out once.
func newPage(t *testing.T, cfg Config, setup func(p *Page)) *Page {
	t.Helper()

	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if setup != nil {
		setup(p)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func clickID(t *testing.T, p *Page, ctx context.Context, id string) {
	t.Helper()

	x, y := boxCenter(t, p, id)
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}

func boxCenter(t *testing.T, p *Page, id string) (float64, float64) {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id {
			return b.X + b.W/2, b.Y + b.H/2
		}
	}

	t.Fatalf("no box %q", id)

	return 0, 0
}

// samePicture reports whether two decoded pictures share bounds and every
// pixel, so encoder byte drift never fails a paint-equality check.
func samePicture(a, b image.Image) bool {
	if !a.Bounds().Eq(b.Bounds()) {
		return false
	}

	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}

	return true
}

func TestResizePictureMatchesFreshPage(t *testing.T) {
	t.Parallel()
	const source = `<style>body{margin:0;background:#fff}#a{width:40px;height:30px;background:red}
@media(min-width:400px){#a{width:120px;background:blue}}</style><div id="a">{{.Label}}</div><input id="field" value="live">`
	data := map[string]string{"Label": "unchanged data"}
	newAt := func(width int) *Page {
		p, err := New(Config{HTML: source, Width: width, Height: 200})
		if err != nil {
			t.Fatal(err)
		}
		p.SetData(data)
		if err := p.Redraw(t.Context()); err != nil {
			t.Fatal(err)
		}
		p.Form().SetFormValue("field", "edited value")
		if err := p.Redraw(t.Context()); err != nil {
			t.Fatal(err)
		}
		if p.Form().FormValue("field") != "edited value" {
			t.Fatal("fixture did not apply the live field edit")
		}
		return p
	}
	decode := func(p *Page) image.Image {
		img, err := png.Decode(bytes.NewReader(p.Probe().PNG()))
		if err != nil {
			t.Fatal(err)
		}
		return img
	}
	p := newAt(320)
	for _, width := range []int{480, 320, 480} {
		p.SetSize(width, 200)
		if err := p.Redraw(t.Context()); err != nil {
			t.Fatal(err)
		}
		if !samePicture(decode(p), decode(newAt(width))) {
			t.Fatalf("resized picture differs from a fresh page at width %d", width)
		}
	}
}
