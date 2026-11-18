package window

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestStickyViewportPixelsAndHits(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fixture, err := os.ReadFile("testdata/sticky.html")
	if err != nil {
		t.Fatal(err)
	}
	p, err := page.New(page.Config{Width: 200, Height: 100, Scrollbar: host.ScrollbarHidden, HTML: string(fixture)})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData("red")
	clicked := ""
	p.Handle(page.Handlers{Click: func(_ context.Context, box page.Box) error { clicked = box.Action; return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	shell := newShell(ctx, p)
	defer shell.cancel()
	defer shell.closeA11y()
	game := &stickyPixelGame{shell: shell, page: p, ctx: ctx, test: t, clicked: &clicked}
	ebiten.SetWindowSize(200, 100)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}
	if game.err != nil {
		t.Fatal(game.err)
	}
	if game.step != 10 {
		t.Fatal("sticky native scenario did not finish")
	}
}

type stickyPixelGame struct {
	shell    *shell
	page     *page.Page
	ctx      context.Context
	test     *testing.T
	clicked  *string
	step     int
	prepared bool
	err      error
}

func (g *stickyPixelGame) Layout(int, int) (int, int) { return 200, 100 }
func (g *stickyPixelGame) Update() error {
	if g.err != nil || g.step == 10 {
		return ebiten.Termination
	}
	if g.prepared {
		return nil
	}
	offsets := []int{0, 100, 200, 360, 400, 200, 200, 300, 300, 0}
	g.shell.scrollY = offsets[g.step]
	if g.step == 6 {
		g.page.SetData("yellow")
		if err := g.page.Redraw(g.ctx); err != nil {
			return err
		}
	}
	g.shell.pageZoom = 1
	if g.step == 7 || g.step == 8 {
		g.shell.pageZoom = 1.5
	}
	if err := g.shell.syncImage(); err != nil {
		return err
	}
	if err := g.shell.syncScrollWindow(); err != nil {
		return err
	}
	if g.step == 2 || g.step == 8 {
		x, y := g.shell.contentAt(10, 10, 200, 100)
		if err := g.page.Click(g.ctx, x, y); err != nil {
			return err
		}
		if *g.clicked != "head" {
			return fmt.Errorf("sticky screen point hit %q at zoom %g", *g.clicked, g.shell.zoom())
		}
		if g.step == 2 {
			x, y = g.shell.contentAt(175, 10, 200, 100)
			if err := g.page.Click(g.ctx, x, y); err != nil {
				return err
			}
			if *g.clicked != "cover" {
				return fmt.Errorf("higher-z overlap hit %q", *g.clicked)
			}
		}
	}
	g.prepared = true
	return nil
}
func (g *stickyPixelGame) Draw(screen *ebiten.Image) {
	if !g.prepared || g.err != nil {
		return
	}
	g.shell.Draw(screen)
	pixels := make([]byte, 200*100*4)
	screen.ReadPixels(pixels)
	sample := func(x, y int) [3]byte { at := (y*200 + x) * 4; return [3]byte{pixels[at], pixels[at+1], pixels[at+2]} }
	red, blue, yellow := [3]byte{255, 0, 0}, [3]byte{0, 0, 255}, [3]byte{255, 255, 0}
	pointY := 10
	want := red
	switch g.step {
	case 0:
		pointY = 90
	case 3:
		pointY = 10
	case 4:
		want = blue
	case 6, 7, 8:
		want = yellow
	case 9:
		pointY = 90
		want = yellow
	}
	if got := sample(10, pointY); got != want {
		g.err = fmt.Errorf("sticky step %d pixel %v want %v", g.step, got, want)
		return
	}
	if g.step == 2 && sample(175, 10) != [3]byte{0, 255, 0} {
		g.err = fmt.Errorf("higher-z overlap not painted above sticky")
		return
	}
	if g.step == 3 && sample(10, 30) != blue {
		g.err = fmt.Errorf("sticky containing-block release left old paint")
		return
	}
	if (g.step == 7 || g.step == 8) && sample(10, 65) != blue {
		g.err = fmt.Errorf("zoomed sticky paint has wrong height")
		return
	}
	if dir := os.Getenv("OWNFRAME_STICKY_CAPTURE_DIR"); dir != "" {
		//nolint:gosec // reason: the test runner chooses this local capture directory.
		if err := os.MkdirAll(dir, 0700); err != nil {
			g.err = err
			return
		}
		//nolint:gosec // reason: the test runner chooses this local capture directory.
		file, err := os.Create(filepath.Join(dir, fmt.Sprintf("sticky-%d.png", g.step)))
		if err != nil {
			g.err = err
			return
		}
		img := &image.RGBA{Pix: pixels, Stride: 200 * 4, Rect: image.Rect(0, 0, 200, 100)}
		err = png.Encode(file, img)
		closeErr := file.Close()
		if err != nil {
			g.err = err
			return
		}
		if closeErr != nil {
			g.err = closeErr
			return
		}
	}
	g.test.Logf("sticky step %d: scroll=%d zoom=%g pixel=%v click=%s", g.step, g.shell.scrollY, g.shell.zoom(), sample(10, pointY), *g.clicked)
	g.step++
	g.prepared = false
}
