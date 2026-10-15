# Web mode

`Serve` shows the page's latest picture in a browser and sends clicks and
typing back to the same page. The window modes are in
[platforms.md](platforms.md).

## Starting it

```
go run ./examples/login -web
```

`-web` calls `Serve` instead of `Run`, and `-addr` changes the listen address.
`Serve` prints the URL and blocks. The examples and their defaults:

| Example | Address |
|---------|---------|
| login | 127.0.0.1:8091 |
| web | 127.0.0.1:8110 |
| platform | 127.0.0.1:8115 |
| print | 127.0.0.1:8125 |

[examples/web](../examples/web) is the web mode example: a counter, a note
field, and a reset button on one page.

## Routes

| Route | What it does |
|-------|--------------|
| `GET /` | Returns the shell HTML: the frame image with a usemap, a text form, and a backspace form. `Cache-Control: no-store`. |
| `GET /frame.png` | Returns the PNG with `Cache-Control: no-store`, or 404 `no frame` when the page has no picture. |
| `GET /debug/state` | Returns JSON with `Cache-Control: no-store`: size, generation, fallback, boxes, `host.Inspector` stats when the screen has one, and a per-kind operation count map when a display list exists ([devtools.md](devtools.md)). Registered only for `ServeWithOptions` with `ServeOptions{Debug: true}` or `ServeOptions{Perf: true}`; plain `Serve` returns 404. |
| `GET /pdf` | Returns the page's PDF bytes with `Content-Type: application/pdf` and `Cache-Control: no-store`, or 500 when the render fails. The route exists only for a screen that implements `PDF` ([printing.md](printing.md)); other screens get 404. |
| `POST /click` | Reads the query or form floats `x` and `y` in CSS pixels and places the caret at the point before the click handler runs. 400 `bad coordinates` when either fails to parse, 303 to `/` on success, 500 when the page returns an error. |
| `POST /type` | Reads the urlencoded form field `text`. 400 `bad form` when the body does not parse, 303 to `/`, or 500. |
| `POST /backspace` | Deletes one character in the focused text control. 303 to `/`, or 500. |

One mutex serializes the handlers, so two requests never touch the page at
once.

## The image map

`GET /` emits one `<area>` for each box with a non-empty `data-action`, and
the href is `/click?x=<center>&y=<center>` in CSS pixels. The areas are listed
in reverse box order, so an inner element's area comes first and the browser
picks it; the engine's boxes are in document order with the inner element
last.

Plain controls and plain buttons have no `data-action`, so they get no area.
`POST /click` takes any coordinates and hit-tests them against the boxes, so a
click anywhere can still focus a control or reach a click handler.

## The frame

`Serve` draws the page first when it has not been drawn yet. `GET /` reads the
current boxes and size, and `GET /frame.png` returns the latest PNG. A page
kept as a display list has no bitmap, so `Page.PNG` paints once from the
stored template source and caches the bytes until the next `Redraw`
([screen.md](screen.md)). The shell page states that the screen is the image,
not the HTML around it.

## Hot reload

A file-backed page is polled before `GET /` and `GET /frame.png`, under the
same mutex, so an edit to the template or the theme shows on the next request.
While `Page.Watching()` is true the shell page carries a small script that
reloads the frame image every 250 ms, the same interval the window uses. A
string page has no watch, so the shell emits no script.
[hot-reload.md](hot-reload.md) has the config fields, the parse-error
behavior, and what survives a reload.

## What it drops

`Serve` has no window loop:

- No tick ([frames.md](frames.md)).
- No key events ([keys.md](keys.md)).
- No clipboard, undo, redo, select-all, delete-word, or submit route. The
  only text input routes are `/type` and `/backspace`.
- No audio context ([features.md](features.md), [window.md](window.md)).
- No file dialog, so a file input keeps the typed name ([forms.md](forms.md)).
- No dropped files. `Serve` is not the Ebiten loop, so it never reads `ebiten.DroppedFiles` ([drag-drop.md](drag-drop.md)).
- No crash recovery ([crash.md](crash.md)).

A click still focuses a control and `/type` still edits it, so a form takes
text. Values survive a redraw like any other page.

`Serve` has no window, so it lays the page out once at `Page.Size()` and never
relayouts it. Resizing the browser tab changes how the shell shows the picture,
not the layout inside it.

## Trust boundary

`Serve` is a developer preview. It assumes the browser peer is the same user on
the same machine. It returns the rendered page plus its boxes and dimensions,
mutates page state from clicks and typed input, and runs any app handler reached
by `/click`. It has no authentication, TLS, CSRF protection, or request
timeouts. The bind address is caller controlled.

Recommended protections:

- Bind to 127.0.0.1 unless the caller adds its own authentication and
  transport.
- Do not expose plain `Serve` on a shared network. Treat `/click`, `/type`,
  `/backspace`, and `/debug/state` as privileged.
- `GET /debug/state` is registered only with `ServeOptions{Debug: true}` or
  `ServeOptions{Perf: true}`. It returns box text, coordinates, and operation
  counts. Reduce or gate it before exposing a preview. Keep both off in
  shipped binaries unless debugging or profiling is needed;
  `ServeOptions{Perf: true}` alone also adds `GET /debug/pprof/*`.
- The built in server now sets `ReadHeaderTimeout` (5s), `ReadTimeout`
  (15s), `WriteTimeout` (15s), and `IdleTimeout` (60s), limits `/type` to
  256 KiB, rejects non finite `/click` coordinates, and sets `nosniff` and
  `X-Frame-Options: DENY`. `/backspace` reads at most 256 KiB. Put a scoped proxy
  or custom `http.Server` in front of it when exposure is needed.

See the security audit in
[plans/v0.0.2/security](../plans/v0.0.2/security/00-overview-threat-model-and-priority.md).
