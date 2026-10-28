# Window and sizing

`New` takes the first frame size from `Config`. `Run` opens a window at that size. When the user resizes the window, the window lays the page out again at the new size and scrolls a page larger than the window.

## Config

```go
page, err := ownframe.New(ownframe.Config{
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

`Width` and `Height` are the first frame, in CSS pixels. `MinWidth` and `MinHeight` are the smallest frame; zero means 1. `MaxWidth` and `MaxHeight` are the largest window the host asks the OS for; zero means no cap. They do not cap the picture. `New` clamps the first frame up to the minimum. Leave the maxes at zero when the window should keep its OS maximize control. X11 and Windows drop that control as soon as a maximum is set.

`New` returns `ErrBadSize` when `Width` or `Height` is not positive, or when a min is greater than a positive max. A blank template returns `ErrEmptyHTML` first ([screen.md](screen.md)). `Theme` is the extra stylesheet ([theming.md](theming.md)).

## Page size

`Size` returns the frame size in CSS pixels. `MinSize` returns the smallest frame and `MaxSize` the window bound from `Config`. `SetSize(width, height)` stores the size the next `Redraw` uses, pulled up to the minimum; it does not draw. `Clamp(width, height)` returns a size pulled up to the minimum without storing it. Neither one caps a frame at the maximum.

`page.Probe().Generation` increases by one on every successful `Redraw`, on both the replay and bitmap paths. `SetSize`, `SetData`, and `SetTheme` do not move it on their own. `internal/window/sync.go` caches the drawn artifact with the generation it came from and rebuilds it when that generation differs. The window reads the inner screen.

## The window

`Run` opens a decorated, resizable window. The initial size is the page `Size`, the OS minimum is `MinSize`, and the OS maximum is `MaxSize` when it is positive. A desktop window cannot grow past that bound; the layout follows whatever size the window has. On a WebAssembly build the browser gives the canvas size, and `MaxSize` does not apply. On a phone the system gives the screen size, so rotation and split screen work the same way.

## Resize

Ebiten calls `Layout` with the outside size when the window changes. `Layout` pulls a size below 1 up to 1, records it, and returns it. The next update commits it in `resize`:

- A size that held still for one update is laid out at once.
- A size that keeps changing relayouts at most every 100 ms, so text rewraps while the drag runs.
- Between commits the previous frame is scaled to the window.

The last size of a drag always commits, even when the throttle window is still open, and a burst of window events between two updates costs one relayout. The commit calls `SetSize` and `Redraw`, then re-resolves the hover and pressed ids against the new boxes. `internal/window/sync.go` rebuilds the drawn artifact only when the generation or the size changed, so a frame that matches both is drawn from the cache.

A file-backed page relayouts when its source changes too ([hot-reload.md](hot-reload.md)). A reload keeps the frame size, and a page that got shorter pulls the scroll offset back inside the content.

## Fit

`stretched` in `internal/window/view.go` reports that the drawn frame does not match the window. Since the cap became a window bound, that only happens while a relayout is in flight. A committed frame matches the window, so it draws at 1:1, one CSS pixel per window pixel, and scrolls when its content overflows.

The replay path draws through an offscreen buffer at the frame size and scales the buffer. The bitmap path scales the image directly. The wheel is ignored while the picture is stretched, so a scaled picture never scrolls ([scrolling.md](scrolling.md)). The `bitmap fallback` badge comes from the cached artifact, so it does not blink during a drag.

## DPI

The library has no DPI code. No code reads a device scale factor or applies one to a size. `Config.Width`, `SetSize`, and the window all use Ebiten device-independent pixels, so a frame that matches the window is one CSS pixel per window pixel.

## Title

`Config.Title` is the window title. Empty means `ownframe`. `Page.Title` returns it. `Run` reads it once when the window opens, and there is no runtime setter, so a title change after that has no effect.

## Pacing and update order

OwnFrame disables VSync and leaves active drawing uncapped. Ebiten starts the window at its default update rate, about 60 TPS; that update rate is not a drawing cap. The public API does not expose TPS. Idle park and `BindMobile` do call `SetTPS`. `BindMobile` in `run.go` calls `ebiten.SetTPS(ebiten.SyncWithFPS)`. Idle park calls `ebiten.SetTPS(ebiten.SyncWithFPS)` in `applyIdle` and `ebiten.SetTPS(g.base)` in `applyFast`, both in `internal/window/idle_frame.go`. `paceWindow` in `internal/window/idle_game.go` stores `g.base` from `ebiten.TPS()` before the first park. An active CSS animation asks for the next frame at once through `WantFrameIn(0)`; the pacer treats a deadline at or before now as a request to stay on the active drawing loop, so a moving page never enters idle park. The request stops when the last player finishes or waits. The motion driver does not change TPS; idle park and `BindMobile` keep their existing `SetTPS` calls. `OWNFRAME_PACE_DEBUG=1` prints one pacing line per second to stderr; it is dev-only and off by default, so keep it off in shipped binaries. Every `Update` runs a context check, devtools sync, IME, keys, resize, the passthrough update, pointer, touch scroll, hold, drop, tick, the frame opt-in capture, a reload poll, sync (including the Linux AT-SPI publisher hook, which is a no-op until a live bus exists), the page scroll request, a devtools refresh, wheel, the scroll-window sync, and the bitmap viewport in that order (`internal/window/update.go`). Resize runs before pointer, so a hit test uses the boxes of the frame being drawn. A dock toggle requested by pointer or keyboard changes the page flag first; the next devtools sync applies it before resize, keeping drawing and viewport geometry in the same state. The tick is `Page.SetTick`; its call shape and cost are in [frames.md](frames.md). The reload poll asks a file-backed page for changes at most every 250 ms ([hot-reload.md](hot-reload.md)).

## Errors

A tick callback that returns an error stops the window, and `Run` returns it. Cancelling the context does the same. Any handler error behaves the same way: `Update` returns it to Ebiten, the window closes, and `Run` returns it.

## Audio

`Run` and `BindMobile` create a 48 kHz Ebiten audio context once, when one is not already running (`audio_context.go`). `Serve` does not. The library has no audio API; the examples play through [examples/music](../examples/music) ([features.md](features.md)).

## Showing the page

- `Run` opens the desktop window, or the browser canvas on wasm, and blocks until it closes.
- `BindMobile` registers the page with Ebitengine's mobile view. Call it from the package that `ebitenmobile bind` compiles, and do not call `Run` from that package. It returns after registering.
- `Serve` listens on `addr` and blocks. The page at `/` shows the latest picture, and clicks and the type form call the same handlers as `Run` ([forms.md](forms.md)).

All three return `ErrNilPage` when the page is nil. `Run` stores a startup or window failure in a crash report and returns the original error. A canceled context or a deadline returns that error and writes no file. `BindMobile` recovers a panic and returns the original error; `LastReportPath` exposes the file path ([crash.md](crash.md)). Build targets are in [platforms.md](platforms.md).

## Limits

- No second window or window icon. Bottom-right placement and dragging are available through `RunWithOptions`. Fullscreen, the cursor shape, touch scroll and pinch, and programmatic scroll are in [interaction.md](interaction.md).
- No runtime title setter.

## Transparent desktop overlays

`RunWithOptions(ctx, page, WindowOptions{...})` opens a page with optional
window settings. The zero value behaves like `Run`. `Transparent` leaves
unpainted pixels transparent; the template must also have transparent
backgrounds. Use a replayable page, because bitmap rendering may paint a
background. `Borderless` removes the title bar and window borders, `Floating`
requests a window above normal windows, and `FixedSize` disables resizing.
OwnFrame disables VSync. Active drawing is uncapped; there is no VSync option.
The desktop idle policy can still park a static page between scheduled frames.

### Custom window chrome

`CustomChrome` removes the OS title bar so the page can draw its own
minimize, maximize, and close controls, in the style of VS Code. The window
stays resizable and keeps keyboard focus.

```go
err := ownframe.RunWithOptions(ctx, page, ownframe.WindowOptions{
    CustomChrome: true,
    Draggable:    titlebarStrip,
})
```

The page owns the buttons through the `Window` door: `Minimize`,
`ToggleMaximize`, `Maximized` (to swap the maximize/restore icon), and
`Close`, which ends `Run` with a nil error. Point `Draggable` at the
titlebar strip so the window still moves; a press released without moving
reaches the page as a normal click, so buttons in that strip keep working.
Serve, web, and mobile builds ignore the missing decorations: the methods
do nothing there, so shared button wiring is safe. After a maximized commit
the loop re-asserts the configured window size, which travels the normal
client-set presentation path; where the size already matches it is a no-op.
Under WSL, ToggleMaximize never stays in the maximized window state:
the first maximize per monitor learns the workarea with one window-manager
maximize, waits for the restore configure to land, then converts it to plain
client geometry (and restores the saved windowed rect on toggle), because
WSLg can present a maximized borderless window with an offset the client
cannot see, which misaligns pointer input while the layout itself is correct.
The workarea is cached per monitor, so a 1920-wide panel and a 2560-wide
panel each learn their own rect; moving the window to another monitor
re-probes once. A same-frame Restore followed by SetWindowSize loses the
resize on WSLg, so the apply waits for the windowed size to return instead
of stacking both calls.

`PlacementID` remembers the window rect across launches under one id: the
monitor name, the windowed position and size, and whether it closed
maximized. Set it to something like `"code-view"`; empty disables it. A
missing file, a corrupt file, or a gone monitor falls back to the app size
at the window manager's position. Writes are best-effort and debounced off
the update path, so a read-only config never breaks the window.

`MousePassthrough` passes every pointer event to the application underneath,
including clicks on painted pixels. It starts the window unfocused and keeps
its frame ticks running in the background. Provide a way to quit outside the
window, such as a terminal signal. This setting does not make the cat clickable.

`BottomRight` places the window at the current monitor's bottom right, with
`Margin` CSS pixels of inset. Negative margins become zero, and each
coordinate clamps at zero, so a window that does not fit sits at the
monitor's top left. Monitor bounds include taskbars and docks, so choose a
margin that clears them. The window manager controls final placement and
stacking. Linux transparency requires a compositor. Desktop placement,
stacking, decorations, and mouse passthrough do not apply to browser
canvases or mobile views.

The [desktop cat example](../examples/desktop-cat) combines these settings
with animated transparent PNG artwork and an HTML/CSS speech bubble.

### Selective click-through

Set `WindowOptions.Interactive` to a function receiving window-local CSS
coordinates. The window reads the pointer each frame and passes input through
where the function returns false. It starts unfocused and continues ticking
in the background. The callback runs on the window update thread.
Full click-through wins when both are set: `MousePassthrough` ignores
`Interactive` and the window never accepts input. Page content cannot
toggle either setting; only app code sets `WindowOptions`. Browser input routing needs a separate DOM solution.

This setting samples pointer position each frame, so a boundary transition
may take one frame. The callback reserves input regions; it does not move
the cursor or focus the window when a notification arrives.

Set `WindowOptions.Draggable` to reserve a region for moving a desktop window.
A left press in that region starts a gesture. Movement of four CSS pixels
starts dragging; a release without dragging sends the page a normal click.
Dragging retains input until release and moves the window while preserving
the grab point. Only a user's press can begin movement. This option does not
move the cursor. Browser canvases need their own DOM drag behavior.

### Overlay trust

A transparent, always-on-top, click-through window (see
[desktop cat](../examples/desktop-cat)) sits above other apps' prompts,
so it earns extra rules:

- Never accept input while fully click-through. `MousePassthrough`
  drops every pointer event, including painted pixels.
- Reserve only visibly rendered app pixels in `Interactive`: the cat
  claims opaque sprite pixels and its open speech bubble, and empty
  margins pass through. The visible artwork is the focus and ownership
  indicator; a region the user cannot see must not capture clicks.
- Only app code sets transparency and passthrough. Page content,
  notifications, and untrusted data cannot toggle them: there is no page
  API for it, and `Interactive`/`Draggable` callbacks are fixed at
  startup. Treat a request from untrusted content to change them as a
  finding, not a feature.
