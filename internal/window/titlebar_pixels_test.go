//go:build !js && !android && !ios

package window

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestWindowSharedHeaderPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("graphics session required")
	}
	ctx := context.Background()
	p, err := page.New(page.Config{Width: 320, Height: 80, HTML: `<body style="margin:0;background:#ff0000"><div style="height:80px"></div></body>`})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	sh := newShell(ctx, p)
	defer sh.cancel()
	face, err := newHeaderFont()
	if err != nil {
		t.Fatal(err)
	}
	sh.titlebar = titlebarState{enabled: true, style: headerStyle{height: 48, font: face, background: color.NRGBA{B: 255, A: 255}}}
	defer func() {
		if sh.titlebar.canvas != nil {
			sh.titlebar.canvas.Deallocate()
		}
	}()
	game := &headerPixelGame{shell: sh}
	ebiten.SetWindowSize(320, 128)
	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}
	if game.err != nil {
		t.Fatal(game.err)
	}
}

type headerPixelGame struct {
	shell *shell
	done  bool
	err   error
}

func (g *headerPixelGame) Layout(w, h int) (int, int) { return g.shell.Layout(w, h) }
func (g *headerPixelGame) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return g.shell.Update()
}
func (g *headerPixelGame) Draw(screen *ebiten.Image) {
	g.shell.Draw(screen)
	hr, hg, hb, ha := screen.At(2, 2).RGBA()
	cr, cg, cb, ca := screen.At(2, 49).RGBA()
	if hr != 0 || hg != 0 || hb != 0xffff || ha != 0xffff || cr != 0xffff || cg != 0 || cb != 0 || ca != 0xffff {
		g.err = fmt.Errorf("header=(%d,%d,%d,%d) content=(%d,%d,%d,%d); expected blue header above red page", hr, hg, hb, ha, cr, cg, cb, ca)
	}
	g.done = true
}
