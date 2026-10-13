package window

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// chromeChanged reports chrome that differs from the previous set and
// records the new set. An unchanged overlay repaints the same pixels, so it
// does not need the vsync loop. The buffers swap, allocating nothing.
func (g *paceGame) chromeChanged(restore []image.Rectangle) bool {
	g.chromeNext = append(g.chromeNext[:0], restore...)
	changed := !slices.Equal(g.chromeNext, g.chrome)
	g.chrome, g.chromeNext = g.chromeNext, g.chrome

	return changed
}

// viewMoved reports a focus or screen size change, either of which makes
// the replay paint in full.
func (g *paceGame) viewMoved(screen *ebiten.Image) bool {
	focus := ebiten.IsFocused()
	size := screen.Bounds().Size()
	moved := !g.seen || focus != g.focus || size != g.size
	g.seen, g.focus, g.size = true, focus, size

	return moved
}
