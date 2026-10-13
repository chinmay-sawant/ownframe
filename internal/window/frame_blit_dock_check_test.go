package window

import (
	"fmt"
	"image"
)

// checkDock guards frames 5 to 7: the dock hides the box under it, and
// closing it shows the box's current colour. The box's points scale to
// 80..100 x 40..60 pixels, so 90,50 is inside it.
func (g *partialFrameGame) checkDock(bounds image.Rectangle, pixels []byte, scale int) {
	if g.err != nil {
		return
	}

	red := [4]byte{255, 0, 0, 255}
	blue := [4]byte{0, 0, 255, 255}

	switch g.frame {
	case 5:
		if got := g.at(bounds, pixels, 15*scale, 15*scale); got != red {
			g.err = fmt.Errorf("dock frame 5 lost the page box: pixel = %v", got)
		}

		if got := g.at(bounds, pixels, 90*scale, 50*scale); got == red {
			g.err = fmt.Errorf("dock frame 5 left the hidden box showing: pixel = %v", got)
		}
	case 6:
		if got := g.at(bounds, pixels, 90*scale, 50*scale); got == blue {
			g.err = fmt.Errorf("dock frame 6 showed the hidden box: pixel = %v", got)
		}
	case 7:
		if got := g.at(bounds, pixels, 90*scale, 50*scale); got != blue {
			g.err = fmt.Errorf("dock frame 7 kept the stale strip: pixel = %v", got)
		}
	}
}
