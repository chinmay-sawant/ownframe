package window

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestViewportScrollPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	const width, height = 128, 96
	d := partialDisplay(300, 2000)
	for y := 0; y < 2000; y += 80 {
		d.Ops = append(d.Ops, layout.DisplayOp{
			Kind: layout.DisplayOpFillRect, X: 15, Y: float64(y) * d.PointsPerPixel,
			W: 150, H: 24, R: 1, G: float64(y%3) / 2, Alpha: 1,
		})
		d.Order = append(d.Order, len(d.Ops)-1)
	}
	app := &dirtyScreen{fakeScreen: &fakeScreen{width: width, height: height}, display: d, gen: 1}
	g := &viewportPixelGame{shell: newDirtyShell(app), app: app}
	ebiten.SetWindowSize(width, height)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if !g.reused {
		t.Fatal("the scroll never reused the viewport overlap")
	}
}

type viewportPixelGame struct {
	shell  *shell
	app    *dirtyScreen
	step   int
	reused bool
	err    error
}

func (g *viewportPixelGame) Layout(int, int) (int, int) { return 128, 96 }
func (g *viewportPixelGame) Update() error {
	if g.err != nil || g.step == 7 {
		return ebiten.Termination
	}
	return nil
}

func (g *viewportPixelGame) Draw(dst *ebiten.Image) {
	if g.err != nil || g.step == 7 {
		return
	}
	offsets := [...]image.Point{{}, {Y: 120}, {Y: 240}, {X: 40, Y: 300}, {Y: 200}, {Y: 1400}, {}}
	at := offsets[g.step]
	g.shell.scrollX, g.shell.scrollY = at.X, at.Y
	if g.step == 2 {
		g.app.display.Ops[5].B = 1
		g.app.click(image.Rect(0, 240, 300, 280))
		g.shell.seq = g.app.gen
	} else {
		g.app.ok = false
		g.app.rects = nil
	}
	g.shell.partial.paintPx = 0
	g.shell.drawViewport(dst, g.app.display, g.app)
	if g.shell.viewport.spare != nil {
		g.reused = true
	}
	if g.step == 3 && g.shell.partial.paintPx >= 128*288 {
		g.err = fmt.Errorf("overlapping scroll repainted the full viewport: %d pixels", g.shell.partial.paintPx)
		return
	}
	expected := ebiten.NewImage(128, 96)
	defer expected.Deallocate()
	expected.Fill(g.shell.pageBackground())
	g.shell.renderer.DrawVisibleUnfixedZ(expected, g.app.display, -float64(at.X), -float64(at.Y), 0)
	actualPixels, expectedPixels := make([]byte, 128*96*4), make([]byte, 128*96*4)
	dst.ReadPixels(actualPixels)
	expected.ReadPixels(expectedPixels)
	if !bytes.Equal(actualPixels, expectedPixels) {
		g.err = fmt.Errorf("scroll stage %d at %v differs from fresh replay", g.step, at)
	}
	g.step++
}
