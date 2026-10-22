package scene

import "github.com/chinmay-sawant/ownframe/examples/tetris/game"

// rgb is one 8-bit colour.
type rgb struct{ r, g, b uint8 }

func (c rgb) floats() (float64, float64, float64) {
	return float64(c.r) / channelMax, float64(c.g) / channelMax, float64(c.b) / channelMax
}

func (c rgb) hex() string {
	const digits = "0123456789abcdef"

	out := []byte{'#', digits[c.r>>4], digits[c.r&0xf], digits[c.g>>4],
		digits[c.g&0xf], digits[c.b>>4], digits[c.b&0xf]}

	return string(out)
}

// palette names every colour the scene paints or styles. The zero cell
// colour is the empty board cell; kinds index by game.Piece.
type palette struct {
	bg, panel, board, cell rgb
	ink, muted, btn, line  rgb
	over                   rgb
	overAlpha              float64
	ghostCell              rgb
	ghostAlpha             float64
	kinds                  [8]rgb
}

// paletteFor picks the palette for a theme.
func paletteFor(dark bool) *palette {
	if dark {
		return &palettes[1]
	}

	return &palettes[0]
}

// cellColor returns the fill colour and opacity for one board or preview
// cell. A ghost cell uses its piece colour at low opacity.
func cellColor(k game.Piece, ghost, dark bool) (float64, float64, float64, float64) {
	p := paletteFor(dark)

	if !k.Valid() {
		r, g, b := p.cell.floats()

		return r, g, b, 1
	}

	c, a := p.kinds[k], 1.0
	if ghost {
		a = p.ghostAlpha
	}

	r, g, b := c.floats()

	return r, g, b, a
}
