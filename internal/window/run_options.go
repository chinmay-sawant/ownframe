package window

import (
	"context"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func runWindow(ctx context.Context, app host.Screen, options Options) error {
	width, height := app.Size()
	minW, minH := app.MinSize()
	maxW, maxH := windowBounds(app)
	ebiten.SetWindowTitle(app.Title())
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowSizeLimits(minW, minH, maxW, maxH)
	mode := ebiten.WindowResizingModeEnabled
	if options.FixedSize {
		mode = ebiten.WindowResizingModeDisabled
	}

	ebiten.SetWindowResizingMode(mode)
	ebiten.SetWindowDecorated(!options.Borderless)
	ebiten.SetWindowFloating(options.Floating)
	passthrough := options.MousePassthrough || options.Interactive != nil
	ebiten.SetWindowMousePassthrough(passthrough)
	if passthrough {
		ebiten.SetRunnableOnUnfocused(true)
	}

	if options.BottomRight {
		if monitor := ebiten.Monitor(); monitor != nil {
			w, h := monitor.Size()
			margin := max(0, options.Margin)
			ebiten.SetWindowPosition(max(0, w-width-margin), max(0, h-height-margin))
		}
	}

	game := NewGame(ctx, app).(*shell)
	defer game.imeShutdown()
	game.transparent = options.Transparent
	game.interactive, game.passthrough = options.Interactive, passthrough
	game.draggable = options.Draggable
	game.perf = options.Perf || game.perf
	game.wirePerf()

	setupScreenClear()

	return ebiten.RunGameWithOptions(game, &ebiten.RunGameOptions{
		ScreenTransparent: options.Transparent,
		InitUnfocused:     passthrough,
	})
}
