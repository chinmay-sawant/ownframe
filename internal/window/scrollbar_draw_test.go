package window

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestDrawPaintsVerticalThumb(t *testing.T) {
	if os.Getenv("OWNFRAME_SCROLLBAR_GPU_TEST") != "1" {
		t.Skip("requires a graphics session; set OWNFRAME_SCROLLBAR_GPU_TEST=1")
	}
	s := NewGame(context.Background(), &fakeScreen{width: 100, height: 100}).(*shell)
	s.display = &layout.Display{Width: 100, Height: 200, PointsPerPixel: 1}
	ebiten.SetWindowSize(100, 100)
	g := &scrollbarDrawGame{shell: s}
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type scrollbarDrawGame struct {
	shell *shell
	err   error
	drawn bool
}

func (g *scrollbarDrawGame) Layout(int, int) (int, int) { return 100, 100 }
func (g *scrollbarDrawGame) Update() error {
	if g.drawn {
		return ebiten.Termination
	}
	return nil
}
func (g *scrollbarDrawGame) Draw(screen *ebiten.Image) {
	if g.drawn {
		return
	}
	g.shell.Draw(screen)
	pixels := make([]byte, 100*100*4)
	screen.ReadPixels(pixels)
	at := (10*100 + 95) * 4
	want := []byte{scrollThumb.R, scrollThumb.G, scrollThumb.B, scrollThumb.A}
	for i := range want {
		if pixels[at+i] != want[i] {
			g.err = fmt.Errorf("thumb pixel = %v, want %v", pixels[at:at+4], want)
			break
		}
	}
	g.drawn = true
}
