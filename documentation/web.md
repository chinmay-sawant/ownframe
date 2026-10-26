# Web mode

`Serve` is a developer picture preview. It binds to loopback by default, has
no audio, and mutates page state without authentication. It ticks a page that
registers a tick callback, forwards browser keys, wheel events, and clicks, and
refreshes the image when the page reports damage. It is not a 1.0 ship target.
Native hosts are in [platforms.md](platforms.md).

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

[examples/web](../examples/web) is the web mode example: a counter, a note
field, and a reset button on one page.

## Routes

| Route | What it does |
|-------|--------------|
| `GET /` | Returns the interactive shell HTML: the frame image with a usemap, the key/wheel/click script, and no-JS text and backspace forms. `Cache-Control: no-store`. |
| `GET /frame.png` | Returns the PNG with `Cache-Control: no-store`, or 404 `no frame` when the page has no picture. |
| `GET /state` | Returns JSON `{generation,width,height,ticking}` with `Cache-Control: no-store`. The shell polls it and swaps the frame only when the generation or frame counter changes. It exposes no boxes or field text. |
| `GET /debug/state` | Returns JSON with `Cache-Control: no-store`: size, generation, fallback, boxes, `host.Inspector` stats when the screen has one, and a per-kind operation count map when a display list exists ([devtools.md](devtools.md)). Registered only for `ServeWithOptions` with `ServeOptions{InsecureAllowDebugState: true}`; plain `Serve` returns 404. |
| `POST /click` | Reads the query or form floats `x` and `y` in CSS pixels and places the caret at the point before the click handler runs. Requires a matching `Origin` (scheme and host); a missing `Origin` is rejected. 400 `bad coordinates` when either fails to parse, 403 `forbidden` on origin mismatch, 303 to `/` on success, generic 500 when the page returns an error. Query parameters are bounded. |
| `POST /type` | Reads the urlencoded form field `text`. Requires a matching `Origin`. 400 `bad form` when the body does not parse, 303 to `/`, or generic 500. |
| `POST /backspace` | Deletes one character in the focused text control. Requires a matching `Origin`. 303 to `/`, or generic 500. |
| `POST /key` | Reads `key` and `down` (`1` or `0`), plus optional `shift=1` for Tab. Special names only: arrows, page keys, Home, End, Enter, Escape, Space, Tab, Backspace, Delete, and modifiers. Tab calls `host.Focuser`; other names reach `KeyDown`/`KeyUp`. 400 `bad key` or `bad down`, 403 on origin mismatch, 204 on success. |
| `POST /wheel` | Reads finite `x`, `y`, `dx`, `dy` floats and calls `host.WheelHandler.Wheel`. 400 `bad wheel`, 403 on origin mismatch, 204 on success. |

One mutex serializes the handlers and the tick loop, so two requests never
touch the page at once.

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
kept as a display list has no bitmap, so `page.Probe().PNG` paints once from the
stored template source and caches the bytes until the next `Redraw`
([screen.md](screen.md)). The shell page states that the screen is the image,
not the HTML around it.

## Hot reload and ticks

A file-backed page is polled before `GET /` and `GET /frame.png`, under the
same mutex, so an edit to the template or the theme shows on the next request.
While `Page.Window().Watching()` is true the shell page polls `/state` and swaps
the frame image when the generation changes. A string page has no watch, so the
shell still polls `/state` and only swaps when the page redraws.
[hot-reload.md](hot-reload.md) has the config fields, the parse-error
behavior, and what survives a reload.

When the screen implements `host.Ticker`, `Serve` starts a loop at about 60 Hz.
It locks the same mutex, calls `Tick`, and redraws only when the page reports
dirty state or the generation changes. Cancel the context to stop it.

## What it drops

`Serve` is not the native window loop:

- No audio context ([features.md](features.md), [window.md](window.md)).
- No clipboard, undo, redo, select-all, delete-word, or submit route. The
  only text routes are `/type`, `/backspace`, and the special keys in `/key`.
- No file dialog, so a file input keeps the typed name ([forms.md](forms.md)).
- No dropped files. `Serve` is not the Ebiten loop, so it never reads `ebiten.DroppedFiles` ([drag-drop.md](drag-drop.md)).
- No crash-report file. A handler panic becomes HTTP 500
  ([crash.md](crash.md)).

A click still focuses a control and `/type` still edits it, so a form takes
text. Values survive a redraw like any other page.

`Serve` has no window, so it lays the page out once at `Page.Size()` and never
relayouts it. Resizing the browser tab changes how the shell shows the picture,
not the layout inside it.

## Trust boundary

`Serve` is a developer preview. It assumes the browser peer is the same user on
the same machine. It returns the rendered page plus its boxes and dimensions,
mutates page state from clicks and typed input, and runs any app handler reached
by `/click`, `/key`, or `/wheel`. It has no authentication, sessions, or TLS.
Mutating routes (`/click`, `/type`, `/backspace`, `/key`, `/wheel`) require a
matching `Origin` header (scheme and host); a missing `Origin` fails closed
there, while safe `GET`s allow it.
`/click` bounds its query parameters, the server caps header bytes at 64 KiB,
handler failures return a generic `500 internal server error` with the detail
logged server-side sanitized, and `/debug/state` sanitizes reload errors before
encoding them. The bind address is caller controlled, and wildcard binds are
rejected.

Recommended protections:

- Bind to 127.0.0.1 unless the caller adds its own authentication and
  transport. Any non-loopback bind needs app-level auth (an authenticated
  reverse proxy or equivalent) before exposure; `Serve` cannot establish or
  validate a session on its own.
- Do not expose plain `Serve` on a shared network. Treat `/click`, `/type`,
  `/backspace`, `/key`, `/wheel`, and `/debug/state` as privileged.
- `GET /debug/state` is registered only with `ServeOptions{InsecureAllowDebugState: true}`. It returns box text, coordinates, and operation
  counts. Reduce or gate it before exposing a preview. Keep both off in
  shipped binaries unless debugging or profiling is needed;
  `ServeOptions{InsecureAllowPprof: true}` adds `GET /debug/pprof/*`.
  Debug state exposes layout with field text. Profiling exposes process state.
  Keep both on loopback only.
- The built in server now sets `MaxHeaderBytes` (64 KiB), `ReadHeaderTimeout`
  (5s), `ReadTimeout` (15s), `WriteTimeout` (15s), and `IdleTimeout` (60s),
  limits `/type` to 256 KiB, rejects non finite `/click` coordinates, requires
  `Origin` on mutating routes, and sets `nosniff`, `X-Frame-Options: DENY`,
  `Referrer-Policy: no-referrer`, and a restrictive `Permissions-Policy`.
  `/backspace` reads at most 256 KiB. Put a scoped proxy
  or custom `http.Server` in front of it when exposure is needed.

See the security audit in
[plans/v0.0.2/security](../plans/v0.0.2/security/00-overview-threat-model-and-priority.md).
