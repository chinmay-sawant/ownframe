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
	ebiten.SetWindowDecorated(!options.Borderless && !options.CustomChrome)
	ebiten.SetWindowFloating(options.Floating)
	// Full click-through never accepts input: Interactive is honored
	// only when MousePassthrough is off, so a click-through overlay
	// cannot be tricked into clicking. Only app code sets these;
	// page content cannot toggle transparency or passthrough.
	interactive := options.Interactive
	if options.MousePassthrough {
		interactive = nil
	}
	passthrough := options.MousePassthrough || interactive != nil
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

	game := newShell(ctx, app)
	defer game.cancel()
	defer game.renderer.Close()
	defer game.imeShutdown()
	game.transparent = options.Transparent
	game.interactive, game.passthrough = interactive, passthrough
	game.draggable = options.Draggable
	game.perf = options.Perf || game.perf
	game.wirePerf()

	setupScreenClear()

	pace, stop := paceWindow(ctx, game, options)
	defer stop()

	return ebiten.RunGameWithOptions(pace, &ebiten.RunGameOptions{
		ScreenTransparent: options.Transparent,
		InitUnfocused:     passthrough,
	})
}
