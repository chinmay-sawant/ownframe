//go:build !js && !android && !ios

package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestSidebarHoverPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for a graphics session")
	}
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	a.page.Window().SetDevTools(true)
	g := &loglensHoverGame{app: a, ctx: ctx, game: ownframe.NewGame(ctx, a.page)}
	a.page.ScrollTo(0, 13)
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type loglensHoverGame struct {
	app      *app
	ctx      context.Context
	game     ebiten.Game
	step     int
	err      error
	baseline map[string]int
}

func (g *loglensHoverGame) Layout(int, int) (int, int) { return windowWidth, windowHeight }

func (g *loglensHoverGame) Update() error {
	if g.err != nil || g.step == 13 {
		return ebiten.Termination
	}
	if g.step == 0 {
		return g.game.Update()
	}
	if g.step%4 == 1 {
		if err := g.app.page.Edit().Hover(g.ctx, 400, 20); err != nil {
			return err
		}
		return g.app.page.Testing().Tick(g.ctx)
	}
	if g.step%4 == 3 {
		if err := g.app.page.Edit().Hover(g.ctx, 500, 400); err != nil {
			return err
		}
		return g.app.page.Testing().Tick(g.ctx)
	}
	action := "nav:overview"
	if (g.step/3)%2 == 0 {
		action = "nav:logs"
	}
	for _, box := range g.app.page.Boxes() {
		if box.Action == action {
			x, y := box.X+box.W/2, box.Y+box.H/2
			if err := g.app.page.Edit().Hover(g.ctx, x, y); err != nil {
				return err
			}
			return g.app.page.Testing().Tick(g.ctx)
		}
	}
	return fmt.Errorf("sidebar control %s disappeared", action)
}

func (g *loglensHoverGame) Draw(screen *ebiten.Image) {
	if g.err != nil || g.step == 13 {
		return
	}
	g.game.Draw(screen)
	actual := make([]byte, windowWidth*windowHeight*4)
	screen.ReadPixels(actual)
	if g.baseline == nil {
		g.baseline = map[string]int{}
	}
	for _, box := range g.app.page.Boxes() {
		if box.Tag != "span" || box.X >= navLeft || (box.Text != "Overview" && box.Text != "Logs") {
			continue
		}
		ink := 0
		for y := int(box.Y); y < int(box.Y+box.H); y++ {
			for x := int(box.X); x < int(box.X+box.W); x++ {
				i := (y*windowWidth + x) * 4
				if int(actual[i])+int(actual[i+1])+int(actual[i+2]) > 600 {
					ink++
				}
			}
		}
		if g.step == 0 {
			g.baseline[box.Text] = ink
		}
		if ink < 10 || ink < g.baseline[box.Text]/2 {
			g.err = fmt.Errorf("%s disappears on hover frame %d: %d text pixels, baseline %d", box.Text, g.step, ink, g.baseline[box.Text])
			return
		}
	}
	if len(g.baseline) != 2 {
		g.err = fmt.Errorf("sidebar text controls missing from the fixture: %v", g.baseline)
		return
	}
	g.step++
	time.Sleep(50 * time.Millisecond)
}
