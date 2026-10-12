# Window and sizing

`New` takes the first frame size from `Config`. `Run` opens a window at that size. When the user resizes the window, the window relayouts the page and either scrolls or scales the picture to fit.

## Config

```go
page, err := gpui.New(gpui.Config{
    Title:     "Inbox",
    HTML:      `<h1>Inbox</h1>`,
    Theme:     `h1 { color: navy }`,
    Width:     480,
    Height:    320,
    MinWidth:  320,
    MinHeight: 200,
    MaxWidth:  1280,
    MaxHeight: 800,
})
```

`Width` and `Height` are the first frame, in CSS pixels. `MinWidth` and `MinHeight` are the smallest frame; zero means 1. `MaxWidth` and `MaxHeight` cap the picture; zero means 2560. `New` clamps the first frame into that range.

`New` returns `ErrBadSize` when `Width` or `Height` is not positive, or when a min is greater than its max. A blank template returns `ErrEmptyHTML` first ([screen.md](screen.md)). `Theme` is the extra stylesheet ([theming.md](theming.md)).

## Page size

`Size` returns the frame size in CSS pixels. `MinSize` returns the smallest frame. `SetSize(width, height)` stores the size the next `Redraw` uses, pulled into the min and max range; it does not draw. `Clamp(width, height)` returns a size pulled into the same range without storing it.

`Generation` increases by one on every successful `Redraw`, on both the replay and bitmap paths. `SetSize`, `SetData`, and `SetTheme` do not move it on their own. `internal/window/sync.go` caches the drawn artifact with the generation it came from and rebuilds it when `Generation` differs.

## The window

`Run` opens a decorated, resizable window. The initial size is the page `Size`, the OS minimum is `MinSize`, and the OS maximum is unset, so the window can grow past `MaxWidth` and `MaxHeight`. The frame stays at the max and the picture scales up to the window. On a wasm build the window is the browser canvas.

## Resize

Ebiten calls `Layout` with the outside size when the window changes. `Layout` pulls a size below 1 up to 1 and stores it. On the next update `resize` clamps that size with the page min and max, compares it with the current frame size, and calls `SetSize` and `Redraw` when it differs. A frame drawn before that redraw scales the old picture to the window.

## Fit

`stretched` in `internal/window/view.go` picks the draw path each frame.

- Frame equal to the window: draw at 1:1, one CSS pixel per window pixel.
- Layout size equal to the window and the frame larger: draw at 1:1 and scroll.
- Otherwise: scale the picture to the window with a linear filter.

The replay path draws through an offscreen buffer at the frame size and scales the buffer. The bitmap path scales the image directly. The wheel is ignored while the picture is stretched, so a scaled picture never scrolls ([scrolling.md](scrolling.md)).

## DPI

The library has no DPI code. No code reads a device scale factor or applies one to a size. `Config.Width`, `SetSize`, and the window all use Ebiten device-independent pixels, so a frame that matches the window is one CSS pixel per window pixel.

## Title

`Config.Title` is the window title. Empty means `go-gpui`. `Page.Title` returns it. `Run` reads it once when the window opens, and there is no runtime setter, so a title change after that has no effect.

## Pacing and update order

Ebiten updates the window at its default rate, about 60 TPS. The library never calls `SetTPS`, so the rate is not tunable here. Every `Update` runs keys, pointer, resize, tick, sync, and wheel in that order (`internal/window/game.go`). The tick is `Page.SetTick`; its call shape and cost are in [frames.md](frames.md).

## Errors

A tick callback that returns an error stops the window, and `Run` returns it. Cancelling the context does the same. Any handler error behaves the same way: `Update` returns it to Ebiten, the window closes, and `Run` returns it.

## Audio

`Run` and `BindMobile` create a 48 kHz Ebiten audio context once, when one is not already running (`audio_context.go`). `Serve` does not. The library has no audio API; the examples play through [examples/music](../examples/music) ([features.md](features.md)).

## Showing the page

- `Run` opens the desktop window, or the browser canvas on wasm, and blocks until it closes.
- `BindMobile` registers the page with Ebitengine's mobile view. Call it from the package that `ebitenmobile bind` compiles, and do not call `Run` from that package. It returns after registering.
- `Serve` listens on `addr` and blocks. The page at `/` shows the latest picture, and clicks and the type form call the same handlers as `Run` ([forms.md](forms.md)).

All three return `ErrNilPage` when the page is nil. `Run` and `BindMobile` recover a panic and return the crash report path ([crash.md](crash.md)). Build targets are in [platforms.md](platforms.md).

## Limits

- No fullscreen, no second window, no window icon or position, and no cursor shape. The library never calls the Ebiten functions for them.
- No programmatic scroll. The wheel and a thumb drag are the only scroll inputs ([scrolling.md](scrolling.md)).
- No runtime title setter.
- `Run` prints `opening a window` to stdout once, before the loop starts.
