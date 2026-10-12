package window

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// partialFrameGame paints a sentinel under the shell, then checks the pixels
// the frame left. Frame 0 is the first frame, which is full. Frames 1 and 2
// report damage, so they are partial. Frame 3 drops the buffer, which forces
// another full frame. Frame 4 scrolls without damage, so the frame copies the
// whole buffer. Each Draw prepares the next frame, because Ebiten may skip an
// Update between two draws.
type partialFrameGame struct {
	shell *shell
	app   *dirtyScreen
	frame int
	err   error
}

func (g *partialFrameGame) Layout(int, int) (int, int) { return 100, 100 }

func (g *partialFrameGame) Update() error {
	// The focus-regain guard would turn a focus change into a full frame.
	// Keep the recorded focus current so the damage frames stay partial.
	g.shell.partial.focus = ebiten.IsFocused()

	if g.frame >= 5 {
		return ebiten.Termination
	}

	return nil
}

func (g *partialFrameGame) Draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{G: 255, A: 255})
	g.shell.Draw(dst)

	bounds := dst.Bounds()
	pixels := make([]byte, 4*bounds.Dx()*bounds.Dy())
	dst.ReadPixels(pixels)
	g.check(bounds, pixels)
	g.prepare()
	g.frame++
}

func (g *partialFrameGame) prepare() {
	switch g.frame {
	case 0, 1:
		g.app.click(image.Rect(10, 10, 30, 30))
	case 2:
		if g.shell.partial.buf != nil {
			g.shell.partial.buf.Dispose()
			g.shell.partial.buf = nil
		}
	case 3:
		g.app.ok = false
		g.app.rects = nil
		g.shell.scrollY = 20
	}
}
