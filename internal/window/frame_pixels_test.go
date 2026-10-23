package window

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"image/png"
	"os"
	"slices"
	"testing"
)

func (g *partialFrameGame) check(bounds image.Rectangle, pixels []byte) {
	if g.err != nil {
		return
	}

	green := [4]byte{0, 255, 0, 255}
	red := [4]byte{255, 0, 0, 255}
	white := [4]byte{255, 255, 255, 255}
	corner := bounds.Max.Sub(image.Pt(10, 10))
	outside := g.at(bounds, pixels, corner.X, corner.Y)
	inside := bounds.Min.Add(image.Pt(bounds.Dx()/5, bounds.Dy()/5))
	scale := bounds.Dx() / 100

	if g.frame == 0 {
		file, err := os.Open("testdata/partial-frame.png")
		if err != nil {
			g.err = err
			return
		}
		golden, err := png.Decode(file)
		closeErr := file.Close()
		if err != nil {
			g.err = err
			return
		}
		if closeErr != nil {
			g.err = closeErr
			return
		}
		if bounds.Dx() != 100 || bounds.Dy() != 100 {
			g.err = fmt.Errorf("golden size differs: %v", bounds)
			return
		}
		for y := range 100 {
			for x := range 100 {
				r, gv, b, a := golden.At(x, y).RGBA()
				got := g.at(bounds, pixels, x, y)
				const channelScale = 257
				differs := r != uint32(got[0])*channelScale || gv != uint32(got[1])*channelScale ||
					b != uint32(got[2])*channelScale || a != uint32(got[3])*channelScale
				if differs {
					g.err = fmt.Errorf("golden differs at %d,%d: %v != %v", x, y, got, golden.At(x, y))
					return
				}
			}
		}
	}
	switch g.frame {
	case 0, 3:
		if outside == green {
			g.err = fmt.Errorf("full frame %d left the sentinel painted", g.frame)
		}
	case 1, 2:
		if outside != green {
			g.err = fmt.Errorf("partial frame filled the screen: pixel = %v", outside)
		}

		if g.at(bounds, pixels, inside.X, inside.Y) != red {
			g.err = fmt.Errorf("partial frame did not blit the red ink")
		}
		for y := 10 * scale; y < 30*scale; y++ {
			for x := 10 * scale; x < 30*scale; x++ {
				want := white
				if x >= 12*scale && y >= 12*scale {
					want = red
				}
				if got := g.at(bounds, pixels, x, y); got != want {
					g.err = fmt.Errorf("partial damage differs at %d,%d: %v != %v", x, y, got, want)
					return
				}
			}
		}
	case 4:
		// The ink sits at 12..32 px before the scroll. After a 20 px scroll
		// up, row 5 shows the box and row 35 shows the page under it.
		if got := g.at(bounds, pixels, 15*scale, 5*scale); got != red {
			g.err = fmt.Errorf("scroll copy lost the scrolled ink: pixel = %v", got)
		}

		if got := g.at(bounds, pixels, 15*scale, 35*scale); got != white {
			g.err = fmt.Errorf("scroll copy left the old ink: pixel = %v", got)
		}
	default:
		g.checkDock(bounds, pixels, scale)
	}
}

func (g *partialFrameGame) at(bounds image.Rectangle, pixels []byte, x, y int) [4]byte {
	i := (y*bounds.Dx() + x) * 4

	return [4]byte{pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]}
}

// checkDock guards frames 5 to 7: the dock hides the box under it, and
// closing it shows the box's current colour. The box's points scale to
// 80..100 x 40..60 pixels, so 90,50 is inside it.
func (g *partialFrameGame) checkDock(bounds image.Rectangle, pixels []byte, scale int) {
	if g.err != nil {
		return
	}

	red := [4]byte{255, 0, 0, 255}
	blue := [4]byte{0, 0, 255, 255}

	switch g.frame {
	case 5:
		if got := g.at(bounds, pixels, 15*scale, 15*scale); got != red {
			g.err = fmt.Errorf("dock frame 5 lost the page box: pixel = %v", got)
		}

		if got := g.at(bounds, pixels, 90*scale, 50*scale); got == red {
			g.err = fmt.Errorf("dock frame 5 left the hidden box showing: pixel = %v", got)
		}
	case 6:
		if got := g.at(bounds, pixels, 90*scale, 50*scale); got == blue {
			g.err = fmt.Errorf("dock frame 6 showed the hidden box: pixel = %v", got)
		}
	case 7:
		if got := g.at(bounds, pixels, 90*scale, 50*scale); got != blue {
			g.err = fmt.Errorf("dock frame 7 kept the stale strip: pixel = %v", got)
		}
	}
}

func (g *partialFrameGame) prepare() {
	switch g.frame {
	case 0, 1:
		g.app.click(image.Rect(10, 10, 30, 30))
	case 2:
		if g.shell.partial.buf != nil {
			g.shell.partial.buf.Deallocate()
			g.shell.partial.buf = nil
		}
	case 3:
		g.app.ok = false
		g.app.rects = nil
		g.shell.scrollY = 20
	case 4:
		g.shell.scrollY = 0
		g.shell.dev.on = true
		g.shell.dev.dockW = 40
		g.setDockPage(1, 0, 0)
	case 5:
		g.setDockPage(0, 0, 1)
	case 6:
		g.shell.dev.on = false
		g.app.rects, g.app.ok = nil, false
	}
}

// setDockPage swaps in the page whose box sits under the dock and reports
// its damage.
func (g *partialFrameGame) setDockPage(r, gv, b float64) {
	d := dockDisplay(r, gv, b)
	g.app.display, g.shell.display = d, d
	g.app.click(image.Rect(60, 30, 75, 45))
}

// dockDisplay is the 100 x 100 page plus one box under the dock's left edge.
// The box's points scale to 80..100 x 40..60 pixels, right of the 40 px dock.
func dockDisplay(r, g, b float64) *layout.Display {
	d := partialDisplay(100, 100)
	d.Ops = append(d.Ops, layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: 60, Y: 30, W: 15, H: 15,
		R: r, G: g, B: b, Alpha: 1,
	})
	d.Order = []int{0, 1, 2}

	return d
}

// TestPartialFrameLeavesTheScreen is the pixel guard for the swap: a partial
// frame blits only its planned rects, so a sentinel painted outside them
// survives, a full frame fills the whole screen and wipes it, a scroll frame
// copies the whole buffer at the new offset, and a dock frame hides the page
// under the panel without repainting it until the dock closes.
func TestPartialFrameLeavesTheScreen(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("requires a graphics session; set OWNFRAME_FRAME_GPU_TEST=1")
	}

	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	app.display = partialDisplay(100, 100)
	app.gen = 1

	shell := newDirtyShell(app)
	shell.ctx = context.Background()
	game := &partialFrameGame{shell: shell, app: app}

	ebiten.SetWindowSize(100, 100)
	ebiten.SetScreenClearedEveryFrame(false)

	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}

	if game.err != nil {
		t.Fatal(game.err)
	}
}

// partialFrameGame paints a sentinel under the shell, then checks the pixels
// the frame left. Frame 0 is the first frame, which is full. Frames 1 and 2
// report damage, so they are partial. Frame 3 drops the buffer, which forces
// another full frame. Frame 4 scrolls without damage, so the frame copies the
// whole buffer. Frames 5 to 7 open a 40 px dock, damage the box under it, and
// close the dock. Each Draw prepares the next frame, because Ebiten may skip
// an Update between draws.
type partialFrameGame struct {
	shell *shell
	app   *dirtyScreen
	frame int
	err   error
}

func (g *partialFrameGame) Layout(int, int) (int, int) { return 100, 100 }

func (g *partialFrameGame) Update() error {
	// The focus-regain guard would turn a focus change into a full frame.
	// Keep the recorded focus current so the damage frames stay partial.
	g.shell.partial.focus = ebiten.IsFocused()

	if g.frame >= 8 {
		return ebiten.Termination
	}

	return nil
}

func (g *partialFrameGame) Draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{G: 255, A: 255})
	g.shell.Draw(dst)

	bounds := dst.Bounds()
	pixels := make([]byte, 4*bounds.Dx()*bounds.Dy())
	dst.ReadPixels(pixels)
	g.check(bounds, pixels)
	g.prepare()
	g.frame++
}

// stripes is n disjoint 5x5 rects in a row, for the cap cases.
func stripes(n int) []image.Rectangle {
	rects := make([]image.Rectangle, n)
	for i := range rects {
		rects[i] = image.Rect(i*10, 0, i*10+5, 5)
	}

	return rects
}

// planRects turns rects into repaintRect plans in order.
func planRects(rects []image.Rectangle) []repaintPlan {
	plans := make([]repaintPlan, len(rects))
	for i, r := range rects {
		plans[i] = repaintPlan{mode: repaintRect, rect: r}
	}

	return plans
}

func TestBlitPlanDecidesTheFrame(t *testing.T) {
	t.Parallel()

	full := image.Rect(0, 0, 100, 100)
	one := image.Rect(10, 10, 30, 30)
	two := image.Rect(60, 60, 80, 80)
	near := image.Rect(20, 20, 40, 40)
	wide := image.Rect(35, 35, 55, 55)
	chrome := image.Rect(0, 0, 8, 100)
	blit := []repaintPlan{{mode: repaintBlit, rect: full}}
	buffer := []repaintPlan{{mode: repaintBuffer, rect: full}}

	cases := []struct {
		name   string
		plans  []repaintPlan
		chrome []image.Rectangle
		full   bool
		rects  []image.Rectangle
	}{
		{"no plans", nil, nil, false, nil},
		{"blit only", blit, nil, false, nil},
		{"blit keeps chrome", blit, []image.Rectangle{chrome}, false,
			[]image.Rectangle{chrome}},
		{"one rect", planRects([]image.Rectangle{one}), nil, false,
			[]image.Rectangle{one}},
		{"two rects", planRects([]image.Rectangle{one, two}), nil, false,
			[]image.Rectangle{one, two}},
		{"overlapping rects", planRects([]image.Rectangle{one, near}), nil, false,
			[]image.Rectangle{one.Union(near)}},
		{"chrome unions a rect", planRects([]image.Rectangle{one}), []image.Rectangle{near}, false,
			[]image.Rectangle{one.Union(near)}},
		{"overlap through a third", planRects([]image.Rectangle{one, near, wide}), nil, false,
			[]image.Rectangle{one.Union(near).Union(wide)}},
		{"empty rects drop", planRects([]image.Rectangle{image.Rectangle{}, one}),
			[]image.Rectangle{image.Rectangle{}}, false, []image.Rectangle{one}},
		{"buffer plan", buffer, []image.Rectangle{chrome}, true, nil},
		{"at the cap", planRects(stripes(8)), nil, false, stripes(8)},
		{"past the cap", planRects(stripes(9)), nil, false,
			[]image.Rectangle{image.Rect(0, 0, 85, 5)}},
	}

	for _, c := range cases {
		full, rects := blitPlan(c.plans, c.chrome)
		if full != c.full {
			t.Fatalf("%s: full = %v, want %v", c.name, full, c.full)
		}

		if !slices.Equal(rects, c.rects) {
			t.Fatalf("%s: rects = %v, want %v", c.name, rects, c.rects)
		}
	}
}
