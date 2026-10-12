package window

import (
	"fmt"
	"image"
)

func (g *partialFrameGame) check(bounds image.Rectangle, pixels []byte) {
	green := [4]byte{0, 255, 0, 255}
	red := [4]byte{255, 0, 0, 255}
	white := [4]byte{255, 255, 255, 255}
	corner := bounds.Max.Sub(image.Pt(10, 10))
	outside := g.at(bounds, pixels, corner.X, corner.Y)
	inside := bounds.Min.Add(image.Pt(bounds.Dx()/8, bounds.Dy()/8))
	scale := bounds.Dx() / 100

	switch g.frame {
	case 0, 3:
		if outside == green {
			g.err = fmt.Errorf("full frame %d left the sentinel painted", g.frame)
		}
	case 1, 2:
		if outside != green {
			g.err = fmt.Errorf("partial frame filled the screen: pixel = %v", outside)
		}

		if g.at(bounds, pixels, inside.X, inside.Y) == green {
			g.err = fmt.Errorf("partial frame did not blit the damage")
		}
	case 4:
		// The ink sits at 13..40 px before the scroll. After a 20 px scroll
		// up, row 5 shows the box and row 35 shows the page under it.
		if got := g.at(bounds, pixels, 15*scale, 5*scale); got != red {
			g.err = fmt.Errorf("scroll copy lost the scrolled ink: pixel = %v", got)
		}

		if got := g.at(bounds, pixels, 15*scale, 35*scale); got != white {
			g.err = fmt.Errorf("scroll copy left the old ink: pixel = %v", got)
		}
	}
}

func (g *partialFrameGame) at(bounds image.Rectangle, pixels []byte, x, y int) [4]byte {
	i := (y*bounds.Dx() + x) * 4

	return [4]byte{pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]}
}
