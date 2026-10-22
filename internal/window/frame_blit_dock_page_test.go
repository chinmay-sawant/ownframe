package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
