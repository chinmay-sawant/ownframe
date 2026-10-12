## v0.0.1

First public release of **go-gpui**: one Go process that shows an HTML template in a window and calls Go functions for clicks, keys, and typing. There is no JavaScript engine and no second process. `gowkhtmltopdf` parses the HTML, applies the CSS, and lays the page out. A page whose operations the replay accepts keeps the layout as vector operations and no bitmap; any other page falls back to the engine image. The same page runs in a desktop window, a browser canvas, a phone view, or an HTTP picture page.

- Module: `github.com/chinmay-sawant/go-gpui`
- Go: 1.26
- Ebiten: v2.10.4, the floor for text replay
- Engine: `gowkhtmltopdf` pinned at `v0.2.7-0.20261003121325-2111b364213b`
- Direct requires: `gowkhtmltopdf`, Ebiten, `golang.org/x/image`, `golang.org/x/text`
- Docs: `documentation/` (20 files)
- Examples: `examples/readme.md` (24 runnable examples plus shared support)
- This is the first tag. There is no previous release to compare against.

### Highlights

| Area | What you get in v0.0.1 |
|------|------------------------|
| Screen paint | `Redraw` executes the template, applies the CSS, and lays the page out. A replayable page keeps vector operations; any other page keeps the engine image. |
| Replay | Rounded fills; rounded, elliptical, and masked strokes; axis-aligned lines; table grid runs; shaped text with letter-spacing; bullets; and transformed images draw from the retained display list. A fallback frame shows a `bitmap fallback` badge. |
| Images and PNG | `Page.SetImage` registers PNG, JPEG, or SVG bytes under a source name. `Page.PNG` rasterizes on demand and caches until the next redraw. |
| Forms | Input, textarea, and select controls with ids, typing and editing, value getters and setters, the desktop file dialog, and real buttons. |
| Binding | `data-bind` two-way binding to struct fields, with `BeforeEdit` and `Change` hooks. |
| Editing | Type, Backspace, DeleteWord, Paste, SelectAll, Copy, Cut, Submit, Undo, and Redo, driven by handlers or the usual chords. |
| Navigation | `Load`, `Back`, `Forward`, `HTML`, and `data-action` routes, with history errors. |
| IPC | In-process `Send`, `Listen`, `Handle`, and `Request` with cancel semantics and thread safety. |
| Fetch and XHR | One `net/http` request for http and https only, with no cookies or cache. |
| Clipboard | A memory copy plus X11, Windows, and macOS clipboard integration. |
| Crash reports | `Run` and `BindMobile` recover a panic into a local timestamped file. |
| Theming | `Config.Theme` and `SetTheme` layer one stylesheet after the template, with custom properties and live switching. |
| Frames | `Page.SetTick` runs a callback before each frame and can change paint fields without a redraw. |
| Input | Lowercase key events with auto-repeat dropped, the usual editing chords, pointer hover and press state, wheel scrolling with draggable scrollbar thumbs, and configurable frame sizes. |
| Modes | Desktop window, WebAssembly canvas, phone view through `BindMobile`, and an HTTP picture page through `Serve`. |
| Examples | 24 runnable examples plus shared music support, covering the library and the engine. |

### Screen paint and replay

`Page.Redraw` executes the template, asks the engine for the placement as vector operations, and stores the result. When every operation is one the replay supports, the page keeps the display list and no bitmap; otherwise it falls back to the engine image. A page is never half replayed. While a fallback frame is on screen, the window paints a small `bitmap fallback` badge in the top-right corner.

`Page.Display` returns the retained display list, or nil on a fallback page. `Page.Image` returns the bitmap, or nil on a replay page. `Page.PNG` paints on demand when the page replays and caches the bytes until the next redraw, so `GET /frame.png` and tests keep working.

Replay supports fills with circular corners; strokes with circular or elliptical corners and any known side mask; axis-aligned lines and CSS outlines; collapsed table grid runs; shaped text and bullets, including letter-spacing, `text-transform`, and fake bold; and images with a payload, including scaled and rotated transforms. `OpNoop` and `OpLinkURI` paint nothing and are accepted. A page falls back when it uses elliptical fill corners, an unknown stroke mask bit, an image without a payload, a non-normal `mix-blend-mode`, a blend or isolation group, a non-identity transform on a non-image operation, or text with no font, rotation, fake oblique, font features, or autospacing.

`Page.SetImage(src, data)` registers encoded PNG, JPEG, or SVG bytes for a template image source such as `background-image: url("name")` or `<img src="name">`. Both paint paths resolve it, and nil removes the entry. The fetch example sets a fetched image as the page background.

go-gpui never calls the engine's PDF writers. The layout engine is still named gowkhtmltopdf, but this library uses only its HTML, CSS, and layout paths. Text replay requires Ebiten v2.10.4 or newer.

### Forms, binding, and editing

- Templates can use `input`, `textarea`, and `select` with an id. Supported input types: text, no type, password, email, search, tel, url, number, file, checkbox, and radio. A click focuses a text field or textarea, toggles a checkbox, checks a radio and unchecks its same-name siblings, or cycles a select.
- `Type`, `Backspace`, `DeleteWord`, `Paste`, `SelectAll`, `Copy`, and `Cut` edit the focused control even when the handler is nil. A password stores plaintext and paints one bullet per rune.
- Before paint, a text-like input or textarea is rewritten to a span so a long value wraps and grows the box. The rewrite keeps the author's attributes except `value` and `type`, and exposes state as `data-gpui-field`, `data-gpui-focus`, `data-gpui-selected`, `data-gpui-placeholder`, and `data-gpui-caret`. An empty field with a placeholder paints the placeholder. `maxlength` and `readonly` are kept as attributes but not enforced.
- `FormValue`, `FormChecked`, `FormSelected`, and `FocusedField` read stored state. `SetFormValue` and `SetFormChecked` write it and do not redraw.
- `data-bind="Field"` ties a control to a field on the pointer passed to `SetData`. Text-like inputs, textareas, selects, and radios bind to a string; checkboxes bind to a bool. An edit writes through before the redraw. `Handlers.BeforeEdit` runs before the built-in edit and can abort it. `Handlers.Change` runs after a control changes, bound or not.
- Under `Run`, a file input opens the desktop dialog: zenity, qarma, matedialog, or kdialog on Linux, the Windows dialog through PowerShell under WSL, `comdlg32` on Windows, and `osascript` on macOS. wasm, mobile, and the web page keep the typed name.
- A bare button gets a face and a hit box from the engine's default stylesheet. Submit-like inputs are rewritten to buttons. The login example uses real controls and a button.

### Navigation and services

- `Load` swaps the template and pushes history; `Back` and `Forward` move through it; `HTML` returns the current source; `Route` maps a `data-action` to HTML for a click. `ErrNoHistory` and `ErrEmptyHTML` cover the error cases.
- `Send`, `Listen`, `Handle`, and `Request` pass strings inside one process. `Request` returns `ErrNoHandler` when nobody handles the channel. The calls are safe from several goroutines and open no socket.
- `Fetch` sends a GET; `XHR` sends any method with headers and a body. Only http and https are accepted, redirects to another scheme return `ErrScheme`, and the client stores no cookies or cache. Pass a deadline on the context to bound a request.
- `clipboard.Write` and `clipboard.Read` keep a memory copy and use the OS clipboard on X11, Windows, and macOS with cgo. Wayland, macOS without cgo, wasm, and mobile stay on the memory copy. No helper program is started.
- `Run` and `BindMobile` recover a panic into a local UTF-8 report and return an error that names the file. `Report` writes the same file on demand and `SetCrashDir` picks the folder. `Serve` does not recover. Nothing is uploaded.

### Theming and frames

- `Config.Theme` and `Page.SetTheme` add one stylesheet after the template's own styles. Any property the engine implements can be set, including custom properties the template reads with `var()`. `SetTheme("")` removes the theme and `SetTheme` does not draw; the next `Redraw` applies it.
- `Page.SetTick` registers one callback the window calls before each frame. The callback can change paint fields on a retained operation, such as a bar width or a color, without parsing or laying out again. `SetTick(nil)` removes it. `Serve` does not tick.
- The `internal/frame` helpers find operations by box and color for the examples: `Fill`, `Fills`, `Text`, and `BoxUnits`.

### Window, input, and modes

- `Config` sets the first frame size plus min and max. Resizing relayouts at the clamped size. A frame that matches the window draws 1:1; a page that is larger scrolls; otherwise the picture scales to the window.
- The mouse wheel moves an overflowing page 48 px per notch and pans sideways, clamped to the content. An overflowing page gets draggable scrollbar thumbs. Touch sends taps, not drags, so touch users cannot scroll.
- `Handlers.KeyDown` and `KeyUp` receive lowercase key names. The window drops auto-repeat pulses and handles the usual Ctrl and Cmd editing chords. Key handlers do not draw.
- `Page.Hover`, `Press`, `Release`, and `Click` drive `:hover`, `:active`, and focus. A tap arrives as a click.
- `Run` opens the desktop window or the wasm canvas. `BindMobile` registers the phone view through `ebitenmobile bind`. `Serve` shows the latest picture over HTTP with `GET /`, `GET /frame.png`, `GET /click`, `POST /type`, and `POST /backspace`. `Run` and `BindMobile` create a 48 kHz audio context for the examples; `Serve` does not.

### Examples

24 runnable examples plus shared `examples/music` support. Highlights: `login` (sign-in with `secret`/`secret`, real controls, an app-owned undo stack, and desktop, wasm, web, and mobile entries), `forms`, `bind`, `bind-hooks`, `controls`, `editing`, `clipboard`, `states`, `theme`, `history`, `ipc`, `fetch`, `crash`, `platform`, `web`, `png`, `replay`, `shapes`, `layout`, `scrolling`, `wispr-flow-dashboard`, `audio-player`, `spotify-player`, and `dino`. `make open` walks the examples one window at a time. Each package has a headless test.

### Documentation

`documentation/` has 20 files: an index plus guides for the paint pipeline and replay, window sizing, platform modes, web mode, frames, keys, pointer input, scrolling, theming, IPC, navigation, crash reports, fetch, the clipboard, editing, forms, and binding. `documentation/features.md` is the feature index, `documentation/features-examples.md` is a dated scan with file citations, and `examples/readme.md` indexes the examples.

### Limits

- No JavaScript engine, Chromium, preload, `contextBridge`, or Node. One process.
- No cross-process IPC, native menus, tray, notifications, or more than one window. No fullscreen, window icon, or cursor shape.
- No session, cookies, cache, or web storage.
- No DevTools, auto-update, or installer.
- No video, document canvas, WebGL, file drag-and-drop, or context menu.
- No library audio API and no `<audio>` element. The examples play through `examples/music`.
- No IME, accessibility tree, spellcheck, printing, deep links, or OS-global shortcuts.
- No sandbox, CSP, or context isolation.
- File dialogs exist on desktop `Run` only.
- No programmatic scroll, no touch scrolling or pinch, and no soft keyboard on mobile.
- Web mode has no tick, key events, clipboard, submit, undo, select-all, delete-word, audio, file dialog, or crash recovery.
- The engine implements 407 of 818 webref CSS properties. Animation, transition, 3D transforms, scroll snap, and print-only UI names are permanent non-goals. Transformed WOFF2 files are unsupported.

### Install and run

```
go run ./examples/login
go run ./examples/login -web
sh scripts/browser.sh
```

The library:

```go
page, err := gpui.New(gpui.Config{
    Title:  "Hello",
    HTML:   `<h1>{{.Title}}</h1>`,
    Width:  480,
    Height: 640,
})
if err != nil {
    return err
}
page.SetData(struct{ Title string }{"Hello"})
return gpui.Run(context.Background(), page)
```

### Verification

- `make test` passes (`go test -p 1 ./...`).
- 195 test files across `internal/` and `examples/`.
- All relative links in `documentation/` resolve.

### What's changed

* feat: ship go-gpui v0.0.1 by @chinmay-sawant in https://github.com/chinmay-sawant/go-gpui/pull/1
* feat(v0.0.1): complete form fidelity, replay coverage, and the example suite by @chinmay-sawant in https://github.com/chinmay-sawant/go-gpui/pull/2

First release, so there is no previous tag to compare against. The full history is at https://github.com/chinmay-sawant/go-gpui/commits/v0.0.1.
