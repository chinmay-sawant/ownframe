package replay

import (
	"bytes"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

type parityGame struct {
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
		if err := checkGPUOpacity(); err != nil {
			g.err = err
			return
		}
	}
	if g.before == nil {
		g.before = ebiten.NewImage(1908, 999)
		g.after = ebiten.NewImage(1908, 999)
	}
	g.before.Clear()
	g.after.Clear()
	dy := -[]float64{0, 500.25, 1000, 20000, 36000}[g.step]
	for _, i := range g.display.Order {
		op := &g.display.Ops[i]
		drawOp(g.before, op, 0, dy)
	}
	DrawVisible(g.after, g.display, 0, dy)
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
