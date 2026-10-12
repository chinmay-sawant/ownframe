package window

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	menuFill     = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf2}
	menuEdge     = color.RGBA{R: 0xbd, G: 0xbd, B: 0xbd, A: 0xff}
	menuInk      = color.RGBA{R: 0x1c, G: 0x19, B: 0x15, A: 0xff}
	menuDisabled = color.RGBA{R: 0x9c, G: 0x9c, B: 0x9c, A: 0xff}
)

// menuTextWidth measures one row label with the badge face.
func menuTextWidth(label string) (float64, float64) {
	return text.Measure(label, badgeFace, 0)
}

// drawMenu paints the context menu over the frame. The menu is chrome: it
// never enters the page picture and no redraw sees it.
func (s *shell) drawMenu(screen *ebiten.Image) {
	if !s.menu.open || len(s.menu.items) == 0 {
		return
	}

	x, y, w, h := s.menuRect()
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), menuFill, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, menuEdge, false)

	for i, item := range s.menu.items {
		ink := menuInk
		if !item.Enabled {
			ink = menuDisabled
		}

		var op text.DrawOptions
		op.GeoM.Translate(x+menuPadX, y+float64(i)*menuRowH+menuPadY)
		op.ColorScale.ScaleWithColor(ink)
		text.Draw(screen, item.Label, badgeFace, &op)
	}

	s.frameNote(frameRect(x, y, w, h))
}
