# Features

v0.0.1 keeps one Go process and one HTML template. There is no JavaScript engine and no second process. `go.mod` requires `gowkhtmltopdf` and Ebiten directly, plus `golang.org/x/image` and `golang.org/x/text`, both already in Ebiten's tree. These features did not add modules. The example index is [../examples/readme.md](../examples/readme.md).

## Screen paint

`Page.Redraw` fills the `html/template`, then `internal/render.DisplayList` parses the HTML and applies the CSS. A page whose operations `render.Replayable` accepts keeps the result as a `layout.Display` and no bitmap; any other page falls back to `internal/render.Paint`, which calls `layout.Lay` and returns an `image.Image`. The window replays the display list with `internal/replay`, or blits the bitmap on a fallback page. `Page.PNG` rasterizes on demand for `GET /frame.png` and for tests, and caches the bytes until the next `Redraw`. `Page.SetImage` registers encoded image bytes for a template image source; the resolver rides on the render state into both paths, so a fetched image can paint as an `OpImage` or into the bitmap.

`internal/render.DisplayList` stops before the engine's paint step, so it never rasterizes. `Display.Boxes` carries the same hit-test boxes `Lay` returns, so a replayed page needs no bitmap for clicks.

Detail is in [screen.md](screen.md).

## Theming

`Config.Theme` and `Page.SetTheme` store one extra stylesheet that the render state hands to the engine as an extra sheet. The engine walks its rules last, but each sheet numbers rule order from zero, so an equal-specificity tie can still go to a late template rule; theming.md spells out the order rule. Any property the engine implements can appear in the theme, including custom properties the template reads with `var()`. `SetTheme` does not draw; the next `Redraw` applies the sheet. An empty theme is removed.

Detail is in [theming.md](theming.md).

## Frames

`Page.SetTick` registers one function the window calls before it draws each frame. The function can change an operation in the retained display list, or call `Redraw`, so a page can animate without parsing the HTML again. `Serve` does not tick.

Detail is in [frames.md](frames.md).

## Keys

`Handlers.KeyDown` and `Handlers.KeyUp` receive each key press and release with a lowercase key name, such as `"space"` or `"arrowdown"`. A key handler does not draw; the page paints from its tick or calls `Redraw` itself. The window sends one pair per real key event and drops auto-repeat pulses.

Detail is in [keys.md](keys.md).

## IPC

`Send`, `Listen`, `Handle`, and `Request` pass strings between callers in this process. `Send`, `Listen`, and `Handle` do nothing for an empty channel. `Request` returns `ErrNoHandler` for an empty channel or when nobody is handling that channel. There is no socket and no page-process bridge.

Detail is in [ipc.md](ipc.md).

## Navigation

`Load` replaces the template and pushes history. `Back` and `Forward` move through that history. `Route` maps a `data-action` value to HTML, and a click on that action loads it instead of calling the click handler. `HTML` returns the current template source. There is no `loadURL`, no `<a href>` navigation, and no custom protocol.

Detail is in [navigation.md](navigation.md).

## Crash reports

`Run` and `BindMobile` recover a panic, write a text file, and return an error that includes the file path. `Report` writes the same kind of file without a panic. Nothing is uploaded.

Detail is in [crash.md](crash.md).

## Fetch and XHR

`Fetch` sends GET. `XHR` sends a method, headers, and a body. Only `http` and `https` are accepted. A redirect to any other scheme returns `ErrScheme`. The client stores no cookies and no cache.

Detail is in [fetch.md](fetch.md).

## Clipboard

`Write` and `Read` keep an in-memory copy and also talk to the OS clipboard. Linux with `DISPLAY` uses the X11 `CLIPBOARD` selection. Windows uses `CF_UNICODETEXT`. macOS with cgo uses `NSPasteboard`. Wayland, Android, iOS, wasm, and macOS without cgo stay on the memory copy. Tests call `UseMemory` and do not touch the desktop clipboard. No helper program is started.

Detail is in [clipboard.md](clipboard.md).

## Forms

An `input`, `textarea`, or `select` with an id is stored on the page. A click focuses a text field or a textarea, toggles a checkbox, checks a radio, or cycles a select. Typing edits the focused text field even when the type handler is nil. A file input opens the desktop file dialog under `Run` and stores the chosen path; wasm, mobile, and `-web` keep the typed name. `SetFormValue` and `SetFormChecked` do not redraw. A control with `data-bind` is tied to a field on the pointer passed to `SetData`; an edit writes through before the redraw, and `Handlers.Change` receives the changed control's box. `:focus`, `:hover`, and `:active` match with host state, and `:checked` follows the control's `checked` attribute; `data-gpui-*` remains the attribute alternative. The engine's default stylesheet gives a `<button>` a face when the author does not style it, and a submit-like `input` is rewritten to a `button`. The login example uses a button.

Detail is in [forms.md](forms.md). Select all, undo, and redo are in [editing.md](editing.md).

## Still absent

These Electron pieces are not in this branch. The scan that listed them is [../plans/v0.0.1/compare.md](../plans/v0.0.1/compare.md).

- Chromium, V8, preload, `contextBridge`, and Node.
- Cross-process IPC, native menus, tray, notifications, and more than one window. File dialogs exist on desktop `Run` only; wasm, mobile, and `-web` keep the typed name.
- Fullscreen, cursor shape, touch scrolling and pinch, and programmatic scrolling.
- Session, cookies, cache, and web storage.
- DevTools, auto-update, installer, and an uploaded crash dump.
- Video, document canvas, WebGL, file drag-and-drop, and a context menu. There is no library audio API or `<audio>` element; the examples play audio through `examples/music`. `Run` and `BindMobile` create a 48 kHz Ebiten audio context, `Serve` does not ([window.md](window.md)).
- IME, an accessibility tree, spellcheck, printing, deep links, and OS-global shortcuts.
- Sandbox, CSP, and context isolation.
