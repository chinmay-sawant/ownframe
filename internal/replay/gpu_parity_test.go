package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestGPUVisibleParity(t *testing.T) {
	if os.Getenv("OWNFRAME_REPLAY_GPU_TEST") != "1" && os.Getenv("GPUI_REPLAY_GPU_TEST") != "1" {
		t.Skip("requires an isolated graphics session")
	}
	source, err := os.ReadFile("testdata/layout.html")
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]map[string]any, 480)
	for i := range rows {
		rows[i] = map[string]any{"ID": i, "Region": i % 8, "Name": "service"}
	}
	p, err := page.New(page.Config{HTML: string(source), Width: 1908, Height: 999})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(map[string]any{"Revision": 0, "Top": 0, "Bottom": 0, "Nav": []string{"Overview", "Services", "Regions", "Deployments", "Incidents", "Capacity"}, "Rows": rows})
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Display()
	d.Order = append(d.Order, len(d.Ops))
	d.Ops = append(d.Ops, layout.DisplayOp{Kind: layout.DisplayOpFillRect, Fixed: true, X: 10, Y: 10, W: 20, H: 20, R: 1})
	g := &parityGame{display: d}
	defer g.renderer.Close()
	ebiten.SetWindowSize(1908, 999)
	if err = ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type parityGame struct {
	renderer      Renderer
	display       *layout.Display
	before, after *ebiten.Image
	step          int
	err           error
}

func (g *parityGame) Layout(int, int) (int, int) { return 1908, 999 }
func (g *parityGame) Update() error {
	if g.step >= 5 || g.err != nil {
		return ebiten.Termination
	}
	return nil
}
func (g *parityGame) Draw(dst *ebiten.Image) {
	if g.step >= 5 {
		return
	}
	if g.step == 0 {
		if err := checkSleepyEyeRenders(); err != nil {
			g.err = err
			return
		}
		if err := checkTranslucentPixels(); err != nil {
			g.err = err
			return
		}
		if err := checkGPUOpacity(); err != nil {
			g.err = err
			return
		}
		if err := checkRasterCache(); err != nil {
			g.err = err
			return
		}
		if err := checkRasterClip(); err != nil {
			g.err = err
			return
		}
		if err := checkTextRaster(); err != nil {
			g.err = err
			return
		}
		if err := checkTextClip(); err != nil {
			g.err = err
			return
		}
		if err := checkXformParity(); err != nil {
			g.err = err
			return
		}
		if err := checkProjectedText(); err != nil {
			g.err = err
			return
		}
		if err := checkGroupComposite(); err != nil {
			g.err = err
			return
		}
		if err := checkGroupClip(); err != nil {
			g.err = err
			return
		}
		if err := checkChunkParity(); err != nil {
			g.err = err
			return
		}
		if err := checkPathGolden(); err != nil {
			g.err = err
			return
		}
		if err := checkPathXform(); err != nil {
			g.err = err
			return
		}
		if err := checkPathGroup(); err != nil {
			g.err = err
			return
		}
	}
	if g.before == nil {
		g.before = newParityImage(1908, 999)
		g.after = newParityImage(1908, 999)
	}
	g.before.Clear()
	g.after.Clear()
	dy := -[]float64{0, 500.25, 1000, 20000, 36000}[g.step]
	for _, i := range g.display.Order {
		op := &g.display.Ops[i]
		drawOp(g.before, op, 0, dy)
	}
	g.renderer.DrawVisible(g.after, g.display, 0, dy)
	a, b := make([]byte, 1908*999*4), make([]byte, 1908*999*4)
	g.before.ReadPixels(a)
	g.after.ReadPixels(b)
	if !bytes.Equal(a, b) {
		count := 0
		for i := range a {
			if a[i] != b[i] {
				count++
			}
		}
		g.err = fmt.Errorf("scroll %.2f differs in %d channels", -dy, count)
	}
	dst.DrawImage(g.after, nil)
	g.step++
}

// Standalone targets keep the same pixel origin for both sides of a comparison.
// Atlas placement can change interpolation rounding by one color level.
func newParityImage(width, height int) *ebiten.Image {
	return ebiten.NewImageWithOptions(image.Rect(0, 0, width, height), &ebiten.NewImageOptions{Unmanaged: true})
}
