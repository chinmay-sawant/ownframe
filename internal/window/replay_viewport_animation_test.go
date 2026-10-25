package window

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
)

// This isolated 100px fixture has no DevTools and never closes a dock.
func TestViewportAnimatedReturnPixels(t *testing.T) {
	runAnimatedReturnTest(t, true)
}

func TestViewportDirectAnimationReturnPixels(t *testing.T) {
	runAnimatedReturnTest(t, false)
}

func runAnimatedReturnTest(t *testing.T, promote bool) {
	t.Helper()
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	d := windowLayerDisplay()
	d.Height = 2000
	d.Ops[0].H = 1500
	app := &layerDirtyScreen{dirtyScreen: &dirtyScreen{
		fakeScreen: &fakeScreen{width: 100, height: 100}, display: d, gen: 1,
	}}
	s := newDirtyShell(app.dirtyScreen)
	s.app = app
	g := &animatedReturnGame{shell: s, app: app, promote: promote}
	ebiten.SetWindowSize(100, 100)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type animatedReturnGame struct {
	shell   *shell
	app     *layerDirtyScreen
	step    int
	err     error
	promote bool
}

func (g *animatedReturnGame) Layout(int, int) (int, int) { return 100, 100 }
func (g *animatedReturnGame) Update() error {
	if g.err != nil || g.step == 4 {
		return ebiten.Termination
	}
	return nil
}

func (g *animatedReturnGame) Draw(dst *ebiten.Image) {
	if g.err != nil || g.step == 4 {
		return
	}
	matrix := windowLayerMatrix(15)
	if g.step >= 2 {
		matrix = windowLayerMatrix(45)
	}
	for i := 1; i < len(g.app.display.Ops); i++ {
		g.app.display.Ops[i].SetXform(matrix)
		g.app.display.Ops[i].SetPaintOpacity(0.35)
	}
	g.app.specs = nil
	g.shell.scrollY = 130
	if g.step == 0 || g.step == 3 {
		g.shell.scrollY = 0
		if g.promote {
			base, _ := replay.InvertLayerMatrix(matrix)
			g.app.specs = []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: matrix, Base: base, Alpha: 0.35}}
		}
	}
	g.shell.drawViewport(dst, g.app.display, g.app)
	expected := ebiten.NewImage(100, 100)
	defer expected.Deallocate()
	expected.Fill(g.shell.pageBackground())
	g.shell.renderer.DrawVisibleUnfixedZBackground(expected, g.app.display, 0, -float64(g.shell.scrollY), 0)
	g.shell.renderer.DrawLayers(expected, g.app.display, 0, -float64(g.shell.scrollY))
	actualPixels, expectedPixels := make([]byte, 100*100*4), make([]byte, 100*100*4)
	dst.ReadPixels(actualPixels)
	expected.ReadPixels(expectedPixels)
	if !bytes.Equal(actualPixels, expectedPixels) {
		g.err = fmt.Errorf("animation scroll stage %d differs from fresh replay", g.step)
	}
	g.step++
}
