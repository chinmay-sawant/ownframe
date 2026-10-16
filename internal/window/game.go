package window

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

// NewGame returns the screen loop Run and BindMobile use.
// The ctx is run-scoped: stored once with its cancel, never per request.
func NewGame(ctx context.Context, app host.Screen) ebiten.Game {
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)

	width, height := 0, 0
	if app != nil {
		width, height = app.Size()
	}

	game := &shell{
		app:      app,
		ctx:      runCtx,
		cancel:   cancel,
		chords:   newChordWatch(),
		watched:  newKeyWatch(),
		pageZoom: 1,
		dev:      devState{watch: newKeyWatch()},
		pendingW: width,
		pendingH: height,
		screenW:  width,
		screenH:  height,
		perf:     perfEnabled(app),
	}
	game.wirePerf()
	game.imeInit()
	game.configureTouchScroll()

	return game
}
