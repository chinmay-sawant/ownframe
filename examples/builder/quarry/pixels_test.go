package main

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
	quarryui "github.com/chinmay-sawant/ownframe/examples/builder/quarry/ui"
)

func newApp(ctx context.Context, t testing.TB, w, h int) *quarryui.App {
	t.Helper()
	ownframe.UseMemory(true)
	a, err := quarryui.NewApp(ctx, quarryui.LaunchOptions{DataDir: store.Memory, DemoDir: t.TempDir(), Width: w, Height: h})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	return a
}

func shot(ctx context.Context, t testing.TB, a *quarryui.App) *pngImage {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(a.Page().Probe().PNGWithContext(ctx)))
	if err != nil {
		t.Fatal(err)
	}
	return &pngImage{img}
}

type pngImage struct {
	img interface{ At(x, y int) color.Color }
}

func (p *pngImage) at(x, y int) color.RGBA {
	c, _ := color.RGBAModel.Convert(p.img.At(x, y)).(color.RGBA)
	return c
}

func hex(c color.RGBA) string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

// countColor counts pixels equal to want inside the rectangle.
func (p *pngImage) countColor(want color.RGBA, x0, y0, x1, y1 int) int {
	n := 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if p.at(x, y) == want {
				n++
			}
		}
	}
	return n
}

func rgb(s string) color.RGBA {
	var r, g, b uint8
	_, _ = fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return color.RGBA{r, g, b, 255}
}

func TestPixels(t *testing.T) {
	ctx := context.Background()
	type theme struct {
		name                           string
		dark                           bool
		canvas, title, rail, side, acc string
	}
	themes := []theme{
		{"dark", true, "#1c1c1c", "#171717", "#141414", "#171717", "#3ecf8e"},
		{"light", false, "#ffffff", "#fafafa", "#f4f4f4", "#fafafa", "#1f9d6b"},
	}
	views := []struct {
		name string
		open func(a *quarryui.App) error
	}{
		{"workbench", func(a *quarryui.App) error {
			return a.RunSQL(ctx, "SELECT id, full_name, email FROM users ORDER BY id")
		}},
		{"table", func(a *quarryui.App) error { return a.OpenTableNamed(ctx, "orders") }},
		{"monitor", func(a *quarryui.App) error { return a.OpenMonitor(ctx) }},
	}
	for _, sz := range [][2]int{{1440, 960}, {1280, 800}} {
		w, h := sz[0], sz[1]
		for _, th := range themes {
			for _, v := range views {
				t.Run(fmt.Sprintf("%dx%d/%s/%s", w, h, th.name, v.name), func(t *testing.T) {
					a := newApp(ctx, t, w, h)
					if !th.dark {
						if err := a.ToggleTheme(ctx); err != nil {
							t.Fatal(err)
						}
					}
					if err := v.open(a); err != nil {
						t.Fatal(err)
					}
					img := shot(ctx, t, a)
					// Theme chrome: title bar, rail, sidebar, status bar pixels.
					type probe struct {
						name  string
						color string
						x, y  int
					}
					for _, p := range []probe{
						{"titlebar", th.title, 400, 18},
						{"rail", th.rail, 24, 600},
						{"sidebar", th.side, 270, 600},
						{"status", th.title, 1000, h - 12},
					} {
						if got, want := img.at(p.x, p.y), rgb(p.color); got != want {
							t.Errorf("%s pixel (%d,%d) = %s, want %s", p.name, p.x, p.y, hex(got), hex(want))
						}
					}
					// Canvas and accent appear in the workspace.
					if img.countColor(rgb(th.canvas), 320, 120, w-20, h-40) == 0 && v.name != "monitor" {
						t.Errorf("no %s canvas pixels in workspace", th.canvas)
					}
					if img.countColor(rgb(th.acc), 300, 36, w, 72) == 0 {
						t.Errorf("accent %s missing from the tab bar underline", th.acc)
					}
					// Zero overflow: nothing paints past the window and the
					// right-most column is the workspace, not clipped chrome.
					for _, b := range a.Page().Boxes() {
						if b.X+b.W > float64(w)+1 && b.W > 0 && b.ID != "" && b.ID != "grid-inner" {
							if len(b.ID) > 4 && (b.ID[:4] == "grid" || b.ID[:4] == "cell" || b.ID[:4] == "mon" || b.ID[:4] == "chart") {
								continue
							}
							t.Errorf("box %q overflows horizontally: %+v", b.ID, b)
						}
					}
				})
			}
		}
	}
}
