package window

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

// NewGame returns the screen loop Run and BindMobile use.
// The ctx is run-scoped: stored once with its cancel, never per request.
func NewGame(ctx context.Context, app host.Screen) ebiten.Game {
	return newShell(ctx, app)
}

// newShell builds the shell NewGame and NewGameWithOptions share, so
// neither needs a type assertion on the game value.
func newShell(ctx context.Context, app host.Screen) *shell {
	ebiten.SetVsyncEnabled(false)
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
		dev:      devState{watch: newKeyWatch(), mobile: mobileDevTools},
		pendingW: width,
		pendingH: height,
		screenW:  width,
		screenH:  height,
		perf:     perfEnabled(app),
	}
	game.wirePerf()
	game.imeInit()
	game.initA11y()
	game.configureTouchScroll()

	return game
}
