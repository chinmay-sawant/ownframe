//go:build !js && !android && !ios

package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestCommandBackdropPixelsStableAcrossFrames verifies that opening the
// translucent Command Palette backdrop (.app-overlay) and drawing repeated
// partial hover and idle frames with SetScreenClearedEveryFrame(false) never
// double-composites the backdrop into progressive darkening or a 256px tile
// checkerboard, and that dismissing the palette restores un-dimmed pixels.
func TestCommandBackdropPixelsStableAcrossFrames(t *testing.T) {
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
	// Light mode gives maximum contrast for rgba(0,0,0,0.38) over white
	// (#ffffff -> rgb 158 on 1x blend vs rgb 98 on 2x over-blend).
	if err := a.toggleTheme(ctx); err != nil {
		t.Fatal(err)
	}

	g := &backdropStabilityGame{
		app:  a,
		ctx:  ctx,
		game: ownframe.NewGame(ctx, a.page),
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

type backdropStabilityGame struct {
	app        *app
	ctx        context.Context
	game       ebiten.Game
	step       int
	err        error
	undimmed   [4]byte
	baseline   [][4]byte
	samples    []image.Point
	closedDone bool
}

func (g *backdropStabilityGame) Layout(int, int) (int, int) { return windowWidth, windowHeight }

func (g *backdropStabilityGame) Update() error {
	if g.err != nil || g.step >= 13 {
		return ebiten.Termination
	}
	switch g.step {
	case 0:
		return g.game.Update()
	case 1:
		for _, box := range g.app.page.Boxes() {
			if box.Text == "⌘K" {
				if err := g.app.page.Click(g.ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
					return err
				}
				return g.app.page.Testing().Tick(g.ctx)
			}
		}
		g.app.commandOpen = true
		return g.app.refresh(g.ctx)
	case 12:
		g.app.commandOpen = false
		g.app.commandActive = ""
		g.closedDone = true
		if err := g.app.refresh(g.ctx); err != nil {
			return err
		}
		return g.app.page.Testing().Tick(g.ctx)
	default:
		// Alternate hovering command items (small partial dirty rect) and idle
		// frames so untouched 256px tiles would darken on any over-blend bug.
		items := g.commandItemCenters()
		if len(items) > 0 && g.step%2 == 0 {
			pt := items[(g.step/2)%len(items)]
			if err := g.app.page.Edit().Hover(g.ctx, float64(pt.X), float64(pt.Y)); err != nil {
				return err
			}
			return g.app.page.Testing().Tick(g.ctx)
		}
		return g.game.Update()
	}
}

func (g *backdropStabilityGame) commandItemCenters() []image.Point {
	var pts []image.Point
	for _, box := range g.app.page.Boxes() {
		if box.Action != "" && box.Y > 150 && box.Y < 500 && box.X > 400 && box.X < 1000 {
			pts = append(pts, image.Pt(int(box.X+box.W/2), int(box.Y+box.H/2)))
		}
	}
	return pts
}

func (g *backdropStabilityGame) Draw(screen *ebiten.Image) {
	if g.err != nil || g.step >= 13 {
		return
	}
	g.game.Draw(screen)
	actual := make([]byte, windowWidth*windowHeight*4)
	screen.ReadPixels(actual)

	const mainSampleX, mainSampleY = 370, 800
	switch {
	case g.step == 0:
		idx := (mainSampleY*windowWidth + mainSampleX) * 4
		g.undimmed = [4]byte{actual[idx], actual[idx+1], actual[idx+2], actual[idx+3]}
	case g.step == 1:
		// Sample points outside the command panel across topbar, sidebar, and
		// unfixed main content.
		g.samples = []image.Point{
			{X: 30, Y: 30},
			{X: mainSampleX, Y: mainSampleY},
			{X: 80, Y: 220},
		}
		g.baseline = make([][4]byte, len(g.samples))
		for i, pt := range g.samples {
			idx := (pt.Y*windowWidth + pt.X) * 4
			g.baseline[i] = [4]byte{actual[idx], actual[idx+1], actual[idx+2], actual[idx+3]}
		}
		if g.baseline[1][0] >= g.undimmed[0]-20 {
			g.err = fmt.Errorf("backdrop did not dim main content: got %v, undimmed %v", g.baseline[1], g.undimmed)
			return
		}
	case g.closedDone:
		idx := (mainSampleY*windowWidth + mainSampleX) * 4
		got := [4]byte{actual[idx], actual[idx+1], actual[idx+2], actual[idx+3]}
		if absDiffByte(got[0], g.undimmed[0]) > 2 || absDiffByte(got[1], g.undimmed[1]) > 2 || absDiffByte(got[2], g.undimmed[2]) > 2 {
			g.err = fmt.Errorf("closing command palette left dimmed backdrop at (%d,%d): got %v, want %v",
				mainSampleX, mainSampleY, got, g.undimmed)
			return
		}
	default:
		for i, pt := range g.samples {
			idx := (pt.Y*windowWidth + pt.X) * 4
			got := [4]byte{actual[idx], actual[idx+1], actual[idx+2], actual[idx+3]}
			want := g.baseline[i]
			if absDiffByte(got[0], want[0]) > 2 || absDiffByte(got[1], want[1]) > 2 || absDiffByte(got[2], want[2]) > 2 {
				g.err = fmt.Errorf("frame %d backdrop drifted at (%d,%d): got rgba(%d,%d,%d,%d), frame-1 baseline rgba(%d,%d,%d,%d)",
					g.step, pt.X, pt.Y, got[0], got[1], got[2], got[3], want[0], want[1], want[2], want[3])
				return
			}
		}
	}
	g.step++
}

func absDiffByte(a, b byte) int {
	if a >= b {
		return int(a - b)
	}
	return int(b - a)
}
