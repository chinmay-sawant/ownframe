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

const (
	themeStepClickLight      = 1
	themeStepHoverLight      = 2
	themeStepCheckLight      = 10
	themeStepClickDark       = 11
	themeStepHoverDark       = 12
	themeStepCheckDark       = 20
	themeStepHoverPlatform   = 21
	themeStepHoverLive       = 22
	themeStepHoverTraces     = 23
	themeStepCheckUnhover    = 31
	themeStepDone            = 32
	themeFramePause          = 30 * time.Millisecond
	themeBrightChannelSumMin = 600
	themeDarkChannelSumMax   = 180
	themeLightBgSumMin       = 680
	themeMinInkPixels        = 10
	themeMinCornerSamples    = 1
	themeColorChannels       = 4
)

// TestThemeToggleMidTransitionPixels verifies on the real GPU window path that
// toggling Dark -> Light -> Dark while hovering controls during the 150ms CSS
// transition settles every button and sidebar label to crisp theme colors with
// readable text, and that hovering Platform -> Live -> Traces clears Platform's
// hover highlight cleanly.
func TestThemeToggleMidTransitionPixels(t *testing.T) {
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

	g := &loglensThemeGame{
		app:       a,
		ctx:       ctx,
		game:      ownframe.NewGame(ctx, a.page),
		actedStep: -1,
	}
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type loglensThemeGame struct {
	app       *app
	ctx       context.Context
	game      ebiten.Game
	step      int
	actedStep int
	err       error
}

func (g *loglensThemeGame) Layout(int, int) (int, int) { return windowWidth, windowHeight }

func (g *loglensThemeGame) findBox(label string) (ownframe.Box, error) {
	for _, b := range g.app.page.Boxes() {
		if b.Text == label {
			return b, nil
		}
	}
	return ownframe.Box{}, fmt.Errorf("missing box %q", label)
}

func (g *loglensThemeGame) clickLabel(label string) error {
	b, err := g.findBox(label)
	if err != nil {
		return err
	}
	cx, cy := b.X+b.W/2, b.Y+b.H/2
	g.app.page.Probe().BeginRedrawBatch()
	if err := g.app.page.Edit().Hover(g.ctx, cx, cy); err != nil {
		return err
	}
	if err := g.app.page.Edit().Press(g.ctx, cx, cy); err != nil {
		return err
	}
	if err := g.app.page.Click(g.ctx, cx, cy); err != nil {
		return err
	}
	if err := g.app.page.Edit().Release(g.ctx); err != nil {
		return err
	}
	if err := g.app.page.Probe().EndRedrawBatch(g.ctx); err != nil {
		return err
	}
	return g.app.page.Testing().Tick(g.ctx)
}

func (g *loglensThemeGame) hoverLabel(label string) error {
	b, err := g.findBox(label)
	if err != nil {
		return err
	}
	cx, cy := b.X+b.W/2, b.Y+b.H/2
	g.app.page.Probe().BeginRedrawBatch()
	if err := g.app.page.Edit().Hover(g.ctx, cx, cy); err != nil {
		return err
	}
	if err := g.app.page.Probe().EndRedrawBatch(g.ctx); err != nil {
		return err
	}
	return g.app.page.Testing().Tick(g.ctx)
}

func (g *loglensThemeGame) Update() error {
	if g.err != nil || g.step >= themeStepDone {
		return ebiten.Termination
	}
	if g.actedStep != g.step {
		g.actedStep = g.step
		switch g.step {
		case themeStepClickLight:
			if err := g.clickLabel("Light"); err != nil {
				g.err = err
				return ebiten.Termination
			}
		case themeStepHoverLight, themeStepHoverDark, themeStepHoverTraces:
			if err := g.hoverLabel("Traces"); err != nil {
				g.err = err
				return ebiten.Termination
			}
		case themeStepClickDark:
			if err := g.clickLabel("Dark"); err != nil {
				g.err = err
				return ebiten.Termination
			}
		case themeStepHoverPlatform:
			if err := g.hoverLabel("Platform"); err != nil {
				g.err = err
				return ebiten.Termination
			}
		case themeStepHoverLive:
			if err := g.hoverLabel("Live"); err != nil {
				g.err = err
				return ebiten.Termination
			}
		}
	}
	return g.game.Update()
}

func (g *loglensThemeGame) Draw(screen *ebiten.Image) {
	if g.err != nil || g.step >= themeStepDone {
		return
	}
	g.game.Draw(screen)
	switch g.step {
	case themeStepCheckLight:
		pixels := readScreenPixels(screen)
		if err := g.checkButtonContrast(pixels, "Refresh", false); err != nil {
			g.err = err
		} else if err := g.checkButtonContrast(pixels, "New alert", true); err != nil {
			g.err = err
		}
	case themeStepCheckDark:
		pixels := readScreenPixels(screen)
		for _, label := range []string{"Refresh", "Export", "⌘K"} {
			if err := g.checkButtonContrast(pixels, label, true); err != nil {
				g.err = err
				break
			}
		}
		if g.err == nil {
			if err := g.checkButtonContrast(pixels, "New alert", false); err != nil {
				g.err = err
			}
		}
	case themeStepCheckUnhover:
		pixels := readScreenPixels(screen)
		if err := g.checkButtonContrast(pixels, "Platform", true); err != nil {
			g.err = err
		}
	}
	g.step++
	time.Sleep(themeFramePause)
}

func readScreenPixels(screen *ebiten.Image) []byte {
	pixels := make([]byte, windowWidth*windowHeight*themeColorChannels)
	screen.ReadPixels(pixels)
	return pixels
}

// checkButtonContrast verifies that label's box has the expected background
// polarity (darkBg=true means dark background with bright ink; darkBg=false
// means light background with dark ink) and at least themeMinInkPixels
// contrasting glyph pixels.
func (g *loglensThemeGame) checkButtonContrast(pixels []byte, label string, darkBg bool) error {
	b, err := g.findBox(label)
	if err != nil {
		return err
	}
	x0, y0 := max(0, int(b.X)), max(0, int(b.Y))
	x1, y1 := min(windowWidth, int(b.X+b.W)), min(windowHeight, int(b.Y+b.H))
	bgSamples, inkPixels := 0, 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			i := (y*windowWidth + x) * themeColorChannels
			sum := int(pixels[i]) + int(pixels[i+1]) + int(pixels[i+2])
			if darkBg {
				if sum <= themeDarkChannelSumMax {
					bgSamples++
				}
				if sum >= themeBrightChannelSumMin {
					inkPixels++
				}
			} else {
				if sum >= themeLightBgSumMin {
					bgSamples++
				}
				if sum <= themeDarkChannelSumMax {
					inkPixels++
				}
			}
		}
	}
	if bgSamples < themeMinCornerSamples || inkPixels < themeMinInkPixels {
		return fmt.Errorf("%q contrast failed (darkBg=%v): bgSamples=%d inkPixels=%d", label, darkBg, bgSamples, inkPixels)
	}
	return nil
}
