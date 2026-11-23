// Package window shows a screen in a window.
//
// On a desktop, Run opens a normal window with a title bar. The person can
// drag the edges. The minimum size comes from the screen.
// On a phone, BindMobile hands NewGame to the system view, and the screen
// size is the phone. In a browser, build with GOOS=js GOARCH=wasm and Run
// uses the page instead of a desktop window. Resizing the browser changes
// the same frame.
//
// Mouse clicks and taps call Screen.Click. Typed text calls Type.
// Every key press and release calls KeyDown and KeyUp with a lowercase
// key name.
// Backspace calls Backspace. Ctrl-Backspace calls DeleteWord.
// Enter calls Submit. Ctrl or Command with C, V, X, A, Z, and Y call
// Copy, Paste, Cut, SelectAll, Undo, and Redo.
// The wheel scrolls a page that is taller or wider than the window.
package window

import (
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"image/color"
	"time"
)

const (
	backspaceRepeatAt    = 30
	backspaceRepeatEvery = 4
)

var (
	errNilApp = errors.New("window: nil screen")
	// Window keeps its prefix only as wrapped context; identity is the
	// canonical fetch.ErrNilContext so one errors.Is branch covers Run.
	errNilContext = fmt.Errorf("window: %w", fetch.ErrNilContext)
	errNoImage    = errors.New("window: no image")
)

// Run opens the desktop window, or the browser canvas when built for wasm.
// Run must be called from main, and it returns when the window closes.
func Run(ctx context.Context, app host.Screen) error {
	if ctx == nil {
		return errNilContext
	}

	if app == nil {
		return errNilApp
	}

	return runWindow(ctx, app, Options{})
}

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
	storeContentTop(game.titlebarInset())
	defer contentTop.Store(0)
	defer game.disposeSurfaces()
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

// Options controls the desktop window. Desktop placement and mouse
// passthrough do not apply to browser canvases or mobile views.
type Options struct {
	// NativeHeader uses the operating system's title bar instead of OwnFrame's
	// desktop title bar. WSLg's native maximize can jump between monitors.
	NativeHeader bool
	// HeaderTitle overrides the page title in the shared desktop header.
	// Empty follows the page title. CustomHeader applications draw their own.
	HeaderTitle string
	// HeaderHeight is the shared header's height in CSS pixels. Zero uses 36;
	// positive heights have a minimum of 24 to keep the controls usable.
	HeaderHeight int
	// HeaderFont overrides the shared header title's font, size, and weight.
	// Nil uses the bundled proportional Go Medium face at 13 CSS pixels.
	HeaderFont text.Face
	// HeaderBackground and HeaderForeground override the shared header palette.
	// Nil follows the page's light or dark background.
	HeaderBackground color.Color
	HeaderForeground color.Color
	// HeaderHover and HeaderCloseHover override the controls' hover colors.
	HeaderHover      color.Color
	HeaderCloseHover color.Color
	// Transparent keeps unpainted pixels transparent. Use transparent HTML
	// backgrounds and a replayable page; bitmap pages may paint a background.
	Transparent bool
	// Borderless removes the title bar and window borders.
	Borderless bool
	// CustomHeader removes the OS title bar so the page can draw its own
	// minimize, maximize, and close controls, in the style of VS Code.
	// The page owns those controls through Window.Minimize,
	// Window.ToggleMaximize, Window.Maximized, and Window.Close, and drags
	// the window with Draggable pointed at its titlebar. Unlike a plain
	// Borderless overlay, the window stays resizable and keeps keyboard
	// focus. Serve, web, and mobile builds ignore the missing decorations:
	// the controls simply do nothing there.
	CustomHeader bool
	// Floating keeps the window above other normal windows.
	Floating bool
	// MousePassthrough passes all pointer input, including painted pixels,
	// to windows underneath. It also starts unfocused and ticks in background.
	// A full pass-through window never accepts input: Interactive is
	// ignored while this is set, so untrusted content cannot click.
	MousePassthrough bool
	// Interactive reports clickable window-local CSS pixels. When supplied,
	// only other pixels pass through. It runs on the window update thread.
	// Reserve only visibly rendered app pixels (never empty margins), and
	// only app code may set it: page content cannot toggle passthrough.
	Interactive func(x, y int) bool
	// Draggable starts a window drag only on a user press in this region.
	Draggable func(x, y int) bool
	// FixedSize disables resizing by the user.
	FixedSize bool
	// Perf samples frame times and runtime numbers even while DevTools is
	// closed. Opening DevTools enables sampling automatically.
	Perf bool
	// DebugAddr exposes the window's profiling data on a loopback HTTP
	// address while the window runs, for agents that cannot read the
	// devtools dock. Empty disables the listener, the default, so a
	// shipped app pays nothing. Set it to a loopback address such as
	// "127.0.0.1:0" to serve GET /debug/state with the Frame tab
	// counters and GET /debug/pprof/* with CPU, heap, goroutine, and
	// trace profiles. Non-loopback addresses are rejected. The endpoints
	// need a debug build (-tags ownframe_debug); without the tag the
	// window logs once and serves nothing. Pair it with Perf or an open
	// dock for runtime and frame numbers, otherwise the state carries
	// boxes, stats, and op counts only.
	DebugAddr string
	// BottomRight places the window in the current monitor's bottom right.
	BottomRight bool
	// Margin is the inset in CSS pixels, clamped to zero. Placement uses
	// monitor bounds, which include taskbars and docks.
	Margin int
	// PlacementID remembers the window rect across launches under this id:
	// the monitor it was on, the windowed position and size, and whether it
	// closed maximized. Letters, digits, hyphen, and underscore only; empty
	// disables it. Best-effort: a missing or corrupt file falls back to the
	// app size, and a gone monitor keeps the size at the window manager's
	// position. The code-view example sets "code-view".
	PlacementID string
	// NoIdlePace keeps the window presenting continuously even
	// while the page is idle. Pacing is on by default on the desktop: an
	// idle page stops presenting and runs page ticks at a low keep-alive
	// rate, and input or damage returns the window to uncapped drawing.
	NoIdlePace bool
	// NoAudio skips the sound device. Ebiten opens that device on the first
	// frame after an audio context exists, and a missing device ends the
	// window. A preview that never plays sound sets this.
	NoAudio bool
	// MaxFPS caps heavy presents (Draw work) per second without VSync.
	// The loop runs faster at MaxFPS+20, so a 60 cap shows 60 on the FE
	// while the app ticks at 80 for input smoothness. 0 (the default) is
	// uncapped legacy drawing; negative is also uncapped. Positive
	// 30..1000 is a cap (validated choices 60/120/144/240); anything else
	// falls back to uncapped. Idle parking is untouched: keep-alive wakes
	// are far slower than the limiter.
	MaxFPS int
}

// RunWithOptions opens a window with optional desktop settings.
func RunWithOptions(ctx context.Context, app host.Screen, options Options) error {
	if ctx == nil {
		return errNilContext
	}

	if app == nil {
		return errNilApp
	}

	return runWindow(ctx, app, options)
}

// Desktop present cap without VSync.
//
// The active drawing loop stays on VSync-off maximum and paces itself with a
// monotonic sleep, so arbitrary caps (60/120/144/240) work on any monitor
// refresh without busy-spinning. Presents (heavy Draw work) are capped to
// MaxFPS by pacing every loop iteration to the present interval: each
// iteration presents once, and the sleep holds the rate. The screen is
// retained (SetScreenClearedEveryFrame(false)), so idle keep-alive frames
// stay cheap. Updates run hotter via TPS headroom (MaxFPS+FPSHeadroom), so
// the FE shows MaxFPS but the app ticks at the loop rate for input
// smoothness.
//
// Idle parking is untouched: keep-alive wakes (125 ms) are far below the
// limiter intervals, so idle windows park efficiently with no extra sleep.
const (
	// DefaultMaxFPS is the suggested present cap (Settings/CLI choice).
	// The default runtime is uncapped: Options.MaxFPS zero means no limiter.
	DefaultMaxFPS = 60
	// FPSHeadroom is how much faster Updates run than presents when capped.
	// A 60 FPS cap shows 60 on the FE while the app ticks at 80.
	FPSHeadroom = 20
	// MinFPS and MaxAllowedFPS bound explicit caps; outside falls back to
	// uncapped (the default).
	MinFPS        = 30
	MaxAllowedFPS = 1000
)

// NormalizeMaxFPS maps an Options.MaxFPS value to its effective cap.
// 0 (the default) means uncapped. Negative means explicit uncapped (0).
// Positive values in [MinFPS, MaxAllowedFPS] are kept; anything else falls
// back to uncapped. The return is 0 for uncapped, else the present cap.
func NormalizeMaxFPS(v int) int {
	if v <= 0 {
		return 0
	}
	if v < MinFPS || v > MaxAllowedFPS {
		return 0
	}
	return v
}

// LoopFPS is the internal update (TPS) rate for a present cap: MaxFPS +
// headroom. Presents run at MaxFPS while Updates average the loop rate, so
// the FE shows MaxFPS FPS but the app ticks hotter for input smoothness.
// It returns 0 for uncapped.
func LoopFPS(maxFPS int) int {
	if maxFPS <= 0 {
		return 0
	}
	return maxFPS + FPSHeadroom
}

// PresentInterval is the minimum time between loop iterations (each does one
// heavy present). The loop is paced to the present cap; Updates run hotter
// via TPS headroom without extra presents.
func PresentInterval(maxFPS int) time.Duration {
	if maxFPS <= 0 {
		return 0
	}
	return time.Second / time.Duration(maxFPS)
}

// LoopInterval paces loop iterations to the present cap (no VSync sleep
// beyond the cap, no busy-spin).
func LoopInterval(maxFPS int) time.Duration {
	return PresentInterval(maxFPS)
}

// loopWait reports how long to sleep before the next loop iteration to hold
// the present cap. It is pure (no sleeping) so tests need no window.
func loopWait(maxFPS int, lastLoop, now time.Time) time.Duration {
	interval := LoopInterval(maxFPS)
	if interval <= 0 || lastLoop.IsZero() {
		return 0
	}
	if d := lastLoop.Add(interval).Sub(now); d > 0 {
		return d
	}
	return 0
}

// nextLoopBase advances the pacing anchor by one interval when the loop is
// keeping pace, absorbing small OS sleep overshoot instead of accumulating
// frame-rate drift. After a stall (or the first frame), it re-anchors to now
// so the loop never bursts catch-up frames.
func nextLoopBase(maxFPS int, lastLoop, now time.Time) time.Time {
	interval := LoopInterval(maxFPS)
	if interval <= 0 || lastLoop.IsZero() {
		return now
	}
	target := lastLoop.Add(interval)
	if now.Before(target) || now.Sub(target) >= interval {
		return now
	}
	return target
}
