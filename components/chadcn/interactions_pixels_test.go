//go:build !js && !android && !ios

package chadcn_test

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

// Each step runs page input on the native update loop, then captures its actual
// framebuffer. RunGame gets a dedicated process through check-graphics.sh.
func TestManagedInteractionsNative(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	f := newManagedControls(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	g := &managedNativeGame{controls: f, ctx: ctx, test: t, game: ownframe.NewGame(ctx, f.view.Page)}
	ebiten.SetWindowSize(640, 480)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if g.step != 5 {
		t.Fatal("native interactions did not finish")
	}
}

type managedNativeGame struct {
	controls *managedControls
	ctx      context.Context
	test     *testing.T
	game     ebiten.Game
	step     int
	prepared bool
	err      error
}

func (g *managedNativeGame) Layout(int, int) (int, int) { return 640, 480 }
func (g *managedNativeGame) Update() error {
	if g.err != nil || g.step == 5 {
		return ebiten.Termination
	}
	if g.prepared {
		return g.game.Update()
	}
	f, p := g.controls, g.controls.view.Page
	switch g.step {
	case 1:
		if err := p.Form().Focus(g.ctx, "a-tab"); err != nil {
			return err
		}
		if err := p.Edit().KeyDown(g.ctx, "arrowright"); err != nil {
			return err
		}
		if f.active != "b" || p.Form().FocusID() != "b-tab" {
			return fmt.Errorf("native tab selection/focus failed")
		}
	case 2:
		rail := managedBox(g.test, p, "volume-rail")
		x, y := rail.X+rail.W/2, rail.Y+rail.H/2
		if err := p.Edit().Press(g.ctx, x, y); err != nil {
			return err
		}
		claimed, err := p.Testing().BeginDrag(g.ctx, x, y)
		if err != nil {
			return err
		}
		if !claimed {
			return fmt.Errorf("native slider did not claim")
		}
		if err := p.Testing().MoveDrag(g.ctx, x+rail.W/4, y); err != nil {
			return err
		}
		if err := p.Testing().EndDrag(g.ctx); err != nil {
			return err
		}
		if f.value != 75 {
			return fmt.Errorf("native slider value %g", f.value)
		}
	case 3:
		trigger := managedBox(g.test, p, "open")
		if err := p.Click(g.ctx, trigger.X+trigger.W/2, trigger.Y+trigger.H/2); err != nil {
			return err
		}
		for range 6 {
			if err := p.Form().FocusNext(g.ctx); err != nil {
				return err
			}
		}
		if !f.open || p.Form().FocusID() == "outside" {
			return fmt.Errorf("native modal lost focus containment")
		}
	case 4:
		if err := p.Form().FocusPrev(g.ctx); err != nil {
			return err
		}
		if err := p.Edit().KeyDown(g.ctx, "escape"); err != nil {
			return err
		}
		if f.open || p.Form().FocusID() != "open" {
			return fmt.Errorf("native modal did not close/restore")
		}
	}
	g.prepared = true
	return g.game.Update()
}
func (g *managedNativeGame) Draw(screen *ebiten.Image) {
	g.game.Draw(screen)
	if !g.prepared || g.err != nil {
		return
	}
	if dir := os.Getenv("OWNFRAME_INTERACTION_CAPTURE_DIR"); dir != "" {
		pixels := make([]byte, 640*480*4)
		screen.ReadPixels(pixels)
		img := &image.RGBA{Pix: pixels, Stride: 640 * 4, Rect: image.Rect(0, 0, 640, 480)}
		//nolint:gosec // reason: the test runner selects this local framebuffer capture directory.
		if err := os.MkdirAll(dir, 0700); err != nil {
			g.err = err
			return
		}
		path := filepath.Join(dir, fmt.Sprintf("components-%d.png", g.step))
		//nolint:gosec // reason: the test runner explicitly selects the local framebuffer capture directory.
		file, err := os.Create(path)
		if err != nil {
			g.err = err
			return
		}
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
	g.test.Logf("native step %d: tab=%s slider=%g modal=%v focus=%s", g.step, g.controls.active, g.controls.value, g.controls.open, g.controls.view.Page.Form().FocusID())
	g.step++
	g.prepared = false
}
