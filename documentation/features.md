# Features

The library keeps one Go process and one HTML template. There is no JavaScript engine and no second process. `go.mod` requires `blinkless` `v0.0.0-20261010174808-34d93862f9ed` and Ebiten directly, plus `golang.org/x/image` and `golang.org/x/text`, both already in Ebiten's tree. These features did not add modules. The Android runtime and its limits are described in [android-runtime.md](android-runtime.md). The example index is [../examples/readme.md](../examples/readme.md).

## Screen paint

`Page.Redraw` fills the `html/template`, then `internal/render` parses the HTML and applies the CSS. The parsed tree and the styled document are cached by executed source, so a later redraw of the same source skips the parse and the sheet collection. `render.DisplayListDocument` returns the placement as vector operations; when `render.Replayable` accepts every operation, the page keeps that `layout.Display` and no bitmap. Any other page falls back to `bitmap.Paint` of that same list and stores the `image.Image`. The bitmap painter reports unsupported paint instead of silently omitting it. The window replays the display list with `internal/replay`, or blits the bitmap on a fallback page. `Page.PNG` rasterizes on demand for `GET /frame.png` and for tests, and caches the bytes until the next `Redraw`. `Page.SetImage` registers encoded image bytes for a template image source; the resolver rides on the render state into both paths, so a fetched image can paint as an `OpImage` or into the bitmap. A click that changes one box no longer redraws the whole page: the page reports the changed region with `TakeDirty` and the window repaints only it into a content buffer or a bounded viewport buffer for oversized pages, while a fallback page and a page with a frame callback keep the full repaint ([repaint.md](repaint.md)).

`internal/render.DisplayList` stops before the engine's paint step, so it never rasterizes. `Display.Boxes` carries the same hit-test boxes `Lay` returns, so a replayed page needs no bitmap for clicks.

Detail is in [screen.md](screen.md).

## Theming

`Config.Theme` and `Page.SetTheme` store one extra stylesheet that the render state hands to the engine as an extra sheet. The engine walks its rules last, but each sheet numbers rule order from zero, so an equal-specificity tie can still go to a late template rule; theming.md spells out the order rule. Any property the engine implements can appear in the theme, including custom properties the template reads with `var()`. `SetTheme` does not draw; the next `Redraw` applies the sheet. An empty theme is removed.

Detail is in [theming.md](theming.md).

## Hot reload

`Config.File` reads the template from disk at `New` and watches it; `Config.ThemeFile` does the same for the theme. The window polls the files every 250 ms and swaps the template without touching history, so form values, focus, hover, and active survive while their elements do, and `SetImage` and `SetTick` stay. A parse error keeps the last good picture and prints one line to stderr; the bytes are retried on the next poll. `Serve` polls before `GET /` and `GET /frame.png`, and the shell page refreshes the image while a watch is active. wasm and mobile read the file once. Setting `HTML` and `File` together is `ErrBadSource`.

Detail is in [hot-reload.md](hot-reload.md).

## Frames

`Page.SetTick` registers one function the window calls before it draws each frame. The function can change an operation in the retained display list, or call `Redraw`, so a page can animate without parsing the HTML again. `Serve` does not tick.

Detail is in [frames.md](frames.md).

## Keys

`Handlers.KeyDown` and `Handlers.KeyUp` receive each key press and release with a lowercase key name, such as `"space"` or `"arrowdown"`. A key handler does not draw; the page paints from its tick or calls `Redraw` itself. The window sends one pair per real key event and drops auto-repeat pulses. Tab and Shift+Tab move focus when the page has fields, and F11 toggles fullscreen on desktop; [interaction.md](interaction.md) covers both.

Detail is in [keys.md](keys.md).

## IPC

`Send`, `Listen`, `Handle`, and `Request` pass strings between callers in this process. `Send`, `Listen`, and `Handle` do nothing for an empty channel. `Request` returns `ErrNoHandler` for an empty channel or when nobody is handling that channel. There is no socket and no page-process bridge.

Detail is in [ipc.md](ipc.md).

## Navigation

`Load` replaces the template and pushes history. `Back` and `Forward` move through that history. `Route` maps a `data-action` value to HTML, and a click on that action loads it instead of calling the click handler. `HTML` returns the current template source. There is no `loadURL`, no `<a href>` navigation, and no custom protocol.

Detail is in [navigation.md](navigation.md).

## Crash reports

`Run` stores every failure in a text file and returns an error that includes the file path: panics go through the recover, and startup or window errors go through the same writer. `BindMobile` recovers a panic the same way; a browser build recovers and returns an error, but writes no file. `Report` writes the same kind of file without a panic. Nothing is uploaded.

Detail is in [crash.md](crash.md).

## Fetch and XHR

`Fetch` sends GET. `XHR` sends a method, headers, and a body. Only `http` and `https` are accepted. A redirect to any other scheme returns `ErrScheme`. The client stores no cookies and no cache.

Detail is in [fetch.md](fetch.md).

## Clipboard

`Write` and `Read` keep an in-memory copy and also talk to the OS clipboard. Linux with `DISPLAY` uses the X11 `CLIPBOARD` selection. Windows uses `CF_UNICODETEXT`. macOS with cgo uses `NSPasteboard`. Wayland, Android, iOS, wasm, and macOS without cgo stay on the memory copy. Tests call `UseMemory` and do not touch the desktop clipboard. No helper program is started.

Detail is in [clipboard.md](clipboard.md).

## Drag and drop

The window reads Ebiten's dropped files once per frame and hands them to a screen that implements `host.Dropper`. `Handlers.Drop` receives each file's name, size, directory flag, absolute path on desktop, and a `Read` function that serves the bytes during the call. A dropped directory arrives as one entry. `-web` has no window loop, so it receives no drops; the wasm canvas build does.

Detail is in [drag-drop.md](drag-drop.md).

## Printing

`Page.PDF` and `Page.WritePDF` render the last template output to PDF bytes through the engine's document writer, with the current theme injected as a style element in the head. `SavePDF` writes the bytes to a file, and `Print` writes a temporary PDF and hands it to the OS print path: `lp` or `xdg-open` on Linux, `osascript` or `open` on macOS, and PowerShell `Start-Process -Verb Print` on Windows. wasm, Android, and iOS return `ErrNoPrinter`. The zero `PDFOptions` value is A4 with 10 mm margins and no profile; a wrong profile is an error. In web mode `GET /pdf` returns the same bytes. The PDF is a re-render from source, so its pagination can differ from the window, and the display list itself is not printed.

Detail is in [printing.md](printing.md).

## Forms

An `input`, `textarea`, or `select` with an id is stored on the page. A click focuses a text field or a textarea, toggles a checkbox, checks a radio, or cycles a select. Typing edits the focused text field even when the type handler is nil. A file input opens the desktop file dialog under `Run` and stores the chosen path; wasm, mobile, and `-web` keep the typed name. `SetFormValue` and `SetFormChecked` do not redraw. A control with `data-bind` is tied to a field on the pointer passed to `SetData`; an edit writes through before the redraw, and `Handlers.Change` receives the changed control's box. `:focus`, `:hover`, and `:active` match with host state, and `:checked` follows the control's `checked` attribute; `data-ownframe-*` remains the attribute alternative. The engine's default stylesheet gives a `<button>` a face when the author does not style it, and a submit-like `input` is rewritten to a `button`. The login example uses a button.

Detail is in [forms.md](forms.md). Select all, undo, and redo are in [editing.md](editing.md).

## Interaction

Tab and Shift+Tab move focus in document order when a page has fields, and Escape clears it. A mouse press in a field places the caret, a drag extends the selection, a double-click selects a word, a triple-click selects a line, and a press held in place selects the word under the point or reaches `Handlers.LongPress`. A right click opens a shell menu with cut, copy, paste, select all, undo, and redo. The hovered shape picks the cursor, a touch drag scrolls and a pinch zooms by default; `Config.DisableTouchZoom` or `Page.SetTouchZoomAllowed(false)` disables only pinch zoom. A page without that option can still pinch on devices that provide touch input. `Ctrl+=` and `Ctrl+-` zoom the page while `Ctrl+0` resets it, `Page.ScrollTo` and `Page.ScrollBy` move the offset, and F11 toggles fullscreen on desktop.

Detail is in [interaction.md](interaction.md).

## Window options

`RunWithOptions` opens the desktop window with a `WindowOptions` value.
`Transparent` leaves unpainted pixels transparent, `Borderless` removes the
frame, `Floating` keeps the window above normal windows, and `FixedSize`
stops resizing. `BottomRight` places the window at the current monitor's
bottom right, inset by `Margin` CSS pixels. `MousePassthrough` sends every
pointer event to the application underneath, `Interactive` keeps input only
where its callback returns true, and `Draggable` reserves a region for moving
the window. The zero value opens the same window as `Run`; placement,
decorations, and passthrough do not apply to a browser canvas or a phone
view.
`examples/desktop-cat` uses the transparent, click-through settings. Detail
is in [window.md](window.md).

## DevTools

The window can draw an inspector over the page. F12 or Ctrl+Shift+I toggles
it, `Config.DevTools` starts it on, and `Page.SetDevTools` changes it at
runtime. The inspector is a right-side dock with three tabs. Elements shows
the pinned or hovered element as pretty JSON, Frame shows the frame and
pipeline counters, and Ops lists the display operations colour by kind. It
outlines the box under the cursor and the operation clicked in the list. It
is window chrome, so `Page.PNG`, the display list, and the box list never
change. A custom screen implements `host.Inspector` to opt in; a screen
without it gets no overlay. `GET /debug/state` serves the same data in web
mode.

Detail is in [devtools.md](devtools.md).

## Performance

The perf checklist lives in [performance.md](performance.md), with per release numbers in [performance-baseline.md](performance-baseline.md). The Frame tab carries PERFORMANCE, DIRTY, and MEMORY counters for frame stages, repaint work, and memory, and `GET /debug/state` serves the same data in web mode. Sampling is off by default and costs nothing; `Config.Perf`, `ownframe.WithPerf(true)`, `WindowOptions.Perf`, or `ServeOptions{Perf: true}` opts in. The three benchmark apps sit in `examples/perf-benchmarks`, and `scripts/perf-idle.sh`, `scripts/perf-longrun.sh`, and `scripts/perf-baseline.sh` run the idle, soak, and baseline passes.

## Still absent

These Electron pieces are not in this branch. The scan that listed them is [../plans/v0.0.1/compare.md](../plans/v0.0.1/compare.md).

- Chromium, V8, preload, `contextBridge`, and Node.
- Cross-process IPC, native menus, tray, and more than one window. The library ships no notification API; `examples/desktop-cat` draws its own message bubble, takes agent messages over a local HTTP endpoint, and reads Chrome's media session on Windows. File dialogs exist on desktop `Run` only; wasm, mobile, and `-web` keep the typed name.
- Session, cookies, cache, and web storage.
- Auto-update, installer, and an uploaded crash dump.
- Video, document canvas, WebGL, and a context menu were the v0.0.1 absences. The context menu shipped in v0.0.2 ([interaction.md](interaction.md)); video, document canvas, and WebGL are still out. There is no library audio API or `<audio>` element; the examples play audio through `examples/music`. `Run` and `BindMobile` create a 48 kHz Ebiten audio context, `Serve` does not ([window.md](window.md)).
- An accessibility tree, spellcheck, deep links, and OS-global shortcuts. Android and iOS wire the soft keyboard through Ebiten's experimental `exp/textinput` ([interaction.md](interaction.md)); a desktop IME is not wired.
- Sandbox, CSP, and context isolation.

Paint-only hover handlers and scroll callbacks that reuse overscan are described in [frames.md](frames.md).
