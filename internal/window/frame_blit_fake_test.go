package window

import (
	"context"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestPartialFrameLeavesTheScreen is the pixel guard for the swap: a partial
// frame blits only its planned rects, so a sentinel painted outside them
// survives, a full frame fills the whole screen and wipes it, and a scroll
// frame copies the whole buffer at the new offset.
func TestPartialFrameLeavesTheScreen(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("requires a graphics session; set OWNFRAME_FRAME_GPU_TEST=1")
	}

	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	app.display = partialDisplay(100, 100)
	app.gen = 1

	shell := newDirtyShell(app)
	shell.ctx = context.Background()
	game := &partialFrameGame{shell: shell, app: app}

	ebiten.SetWindowSize(100, 100)

	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}

	if game.err != nil {
		t.Fatal(game.err)
	}
}
