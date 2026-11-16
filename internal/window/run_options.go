package window

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func runWindow(ctx context.Context, app host.Screen, options Options) error {
	width, height := app.Size()
	builtinChrome := usesTitlebar(options)
	headerH := headerHeight(options)
	if builtinChrome && options.HeaderFont == nil {
		face, err := newHeaderFont()
		if err != nil {
			return fmt.Errorf("window header font: %w", err)
		}
		options.HeaderFont = face
	}
	if builtinChrome {
		options.CustomHeader = true
		height += headerH
	}
	minW, minH := app.MinSize()
	maxW, maxH := windowBounds(app)
	if builtinChrome {
		minH += headerH
		if maxH > 0 {
			maxH += headerH
		}
	}
	ebiten.SetWindowTitle(app.Title())
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowSizeLimits(minW, minH, maxW, maxH)
	mode := ebiten.WindowResizingModeEnabled
	if options.FixedSize {
		mode = ebiten.WindowResizingModeDisabled
	}

	ebiten.SetWindowResizingMode(mode)
	ebiten.SetWindowDecorated(!options.Borderless && !options.CustomHeader)
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
	game.titlebar.enabled, game.titlebar.fixed = builtinChrome, options.FixedSize
	if builtinChrome {
		game.a11yOrigin = func() (int, int) { x, y := ebiten.WindowPosition(); return x, y + game.titlebarInset() }
	}
	game.titlebar.style = headerStyle{title: options.HeaderTitle, height: headerH, font: options.HeaderFont, background: options.HeaderBackground, foreground: options.HeaderForeground, hover: options.HeaderHover, closeHover: options.HeaderCloseHover}
	game.titlebar.starting = builtinChrome && isWSLg() && options.PlacementID == "" && !options.BottomRight
	contentTop.Store(int32(game.titlebarInset()))
	defer contentTop.Store(0)
	defer func() {
		if game.titlebar.canvas != nil {
			game.titlebar.canvas.Dispose()
		}
	}()
	defer game.cancel()
	defer game.windowDrag.native.end()
	defer game.nativeCursor.end()
	defer game.renderer.Close()
	defer game.dev.panelCache.close()
	defer game.imeShutdown()
	defer game.closeA11y()
	game.transparent = options.Transparent
	game.interactive, game.passthrough = interactive, passthrough
	game.draggable = options.Draggable
	if builtinChrome {
		game.draggable = func(x, y int) bool {
			w, _ := ebiten.WindowSize()
			if y < 0 {
				return y >= -game.titlebarInset() && x >= 0 && x < w-3*titlebarButtonWidth
			}
			return options.Draggable != nil && options.Draggable(x, y)
		}
	}
	game.perfOption = options.Perf
	game.perf = options.Perf || perfEnabled(app)
	game.wirePerf()
	game.placementID = options.PlacementID
	if game.placementID != "" {
		// A remembered rect wins over the default size and BottomRight;
		// a missing file or a gone monitor keeps this frame's geometry.
		game.lastPlacement = applyPlacement(game.placementID, width, height, minW, minH, maxW, maxH)
		game.lastPlacementValid = validPlacementRect(game.lastPlacement.W, game.lastPlacement.H)
	}

	setupScreenClear()

	stopDebug := func() {}
	if options.DebugAddr != "" {
		var err error
		_, stopDebug, err = startDebugServer(ctx, game, options.DebugAddr)
		if err != nil {
			return err
		}
	}
	defer stopDebug()

	pace, stop := paceWindow(ctx, game, options)
	defer stop()

	err := ebiten.RunGameWithOptions(pace, &ebiten.RunGameOptions{
		ScreenTransparent: options.Transparent,
		InitUnfocused:     passthrough,
	})
	savePlacementNow(game.placementID)

	return err
}
