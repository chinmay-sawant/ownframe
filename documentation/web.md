# Web mode

`Serve` is a developer picture preview. It binds to loopback by default and
mutates page state without authentication. It creates the same audio context
as `Run`, so a page that plays sound is heard on this machine. It ticks a page that
registers a tick callback, forwards browser keys, hover, wheel events, and clicks, and
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
| `GET /` | Returns the application HTML document when the page has an executed source. Text and password fields are real input elements. Textareas and selects are real elements. The boot script in the document posts input and clicks to the server. `Cache-Control: no-store`. Append `?dom=0` to force the PNG shell. A page with no executed source gets that PNG shell too: the frame image, a usemap, the key, wheel, and click script, and the no-JS text and backspace forms. |
| `GET /frame.png` | Returns the PNG with `Cache-Control: no-store`, or 404 `no frame` when the page has no picture. Used when the live document is off or the page has no executed source. |
| `GET /audio` | Returns the clip a page published with `SetAudio`, with `Cache-Control: no-store`. The content type is the stored type when it starts with `audio/`. Otherwise the bytes are sniffed: WAV (`RIFF` then `WAVE`) is `audio/wav`, MP3 (`ID3` or frame sync `FF FB`, `FF F3`, or `FF F2`) is `audio/mpeg`, and Ogg (`OggS`) is `audio/ogg`. 404 when the page has no clip. |
| `GET /{name}` | Returns the bytes any page registered with `SetImage` under that source name, so `url("logo")` and `<img src="logo">` load. The path may contain slashes, as in `emoji/1f602`. The content type is sniffed, including PNG, JPEG, GIF, WEBP, BMP, icon, SVG, AVIF, and HEIC. A format the sniffer does not know is still returned. `Cache-Control: no-store`. 404 when the name is missing or contains `.` or `..`. Fixed routes above win over this one. |
| `GET /dom` | Returns JSON `{generation,width,height,html,theme,fallback}` with `Cache-Control: no-store`. `generation` is the page generation without the retained-op frame counter: scroll, hover, and in-place tick patches change the PNG frame but leave the executed source alone, so they do not trigger a `/dom` refetch. `html` is the same document without the boot script, with live values already merged into real controls. The boot script patches id-anchored elements from it and does not replace the focused field. `theme` is the raw theme stylesheet. `fallback` true means the PNG shell. |
| `GET /events` | Opens a `text/event-stream` with `Cache-Control: no-store`. The server sends `retry: 3000`, a current snapshot `data:` frame on connect, one `data:` frame per snapshot change, and a `: ping` heartbeat every 15s. The body reuses the `/state` field names. The live document refetches `/dom` only when `domGeneration` moves; the PNG shell refetches `/frame.png` only when `generation` moves. Same-origin GET, no new privileged route. |
| `GET /state` | Returns JSON `{generation,domGeneration,width,height,ticking,audioGen,audioPlaying,audioVolume}` with `Cache-Control: no-store`. `generation` is the page generation plus the retained-op frame counter. `domGeneration` is the page generation alone. The shell polls it. The boot script fetches `/dom` only when `domGeneration` moves and patches the live document. The PNG shell swaps the frame when `generation` moves. `audioGen` is the published clip generation, or 0 when the page has no clip. `audioPlaying` and `audioVolume` (0 to 1) drive the browser player, so pause and the volume level follow the page. It exposes no boxes or field text. |
| `GET /debug/state` | Returns JSON with `Cache-Control: no-store`: size, generation, fallback, boxes, `host.Inspector` stats when the screen has one, and a per-kind operation count map when a display list exists ([devtools.md](devtools.md)). Registered only for `ServeWithOptions` with `ServeOptions{InsecureAllowDebugState: true}`; plain `Serve` returns 404. |
| `POST /click` | Accepts `id` or `action`, and optional `ox`, `oy`, `w`, `h` that place the point inside the element's box. The PNG shell can still send `x` and `y` in CSS pixels. The handler places the caret at the point before the click handler runs. Requires a matching `Origin` (scheme and host); a missing `Origin` is rejected. 400 `bad coordinates` when either fails to parse, 403 `forbidden` on origin mismatch, 303 to `/` on success, generic 500 when the page returns an error. Query parameters are bounded. |
| `POST /change` | Reads urlencoded `id` plus optional `value` or `checked` (`1`/`true`). A value replaces the whole control value and runs `BeforeEdit` and `Change`, then redraws. The live document does not stream text as one `/type` per key. `/type` and `/backspace` remain for the PNG shell. `checked` updates a checkbox or radio. An id alone moves focus. Requires a matching `Origin`. 400 `bad id` or `bad value`, 403 on origin mismatch, 204 on success. |
| `POST /type` | Reads the urlencoded form field `text`. Requires a matching `Origin`. 400 `bad form` when the body does not parse, 303 to `/`, or generic 500. |
| `POST /backspace` | Deletes one character in the focused text control. Requires a matching `Origin`. 303 to `/`, or generic 500. |
| `POST /key` | Reads `key` and `down` (`1` or `0`), plus optional `shift=1` for Tab. Special names only: arrows, page keys, Home, End, Enter, Escape, Space, Tab, Backspace, Delete, and modifiers. Tab calls `host.Focuser`; other names reach `KeyDown`/`KeyUp`. 400 `bad key` or `bad down`, 403 on origin mismatch, 204 on success. |
| `POST /hover` | Reads finite viewport `x` and `y`, maps the point through the scroll offset, and forwards hover. Requires a matching Origin. |
| `POST /wheel` | Reads finite `x`, `y`, `dx`, `dy` floats and calls `host.WheelHandler.Wheel`. Unhandled events pan the viewport through overflowing content. 400 `bad wheel`, 403 on origin mismatch, 204 on success. |

A page publishes the clip with `SetAudio` and the browser plays it.

One mutex serializes the handlers and the tick loop, so two requests never
touch the page at once.

## The image map

The PNG shell emits one `<area>` for each box with a non-empty `data-action`.
The href is `/click?x=<center>&y=<center>` in CSS pixels. The areas are listed
in reverse box order, so an inner element's area comes first and the browser
picks it. The engine's boxes are in document order with the inner element
last. The live document does not use this map. It posts `id` or `action`.

Plain controls and plain buttons have no `data-action`, so they get no area.
`POST /click` takes coordinates and hit-tests them against the boxes, so a
click anywhere can still focus a control or reach a click handler.

## The frame

`Serve` draws the page first when it has not been drawn yet. The live document
is the default view. The browser shows the authored CSS. The Go page remains
the state and the handlers. `GET /` returns that document when the page has an
executed source. Text and password fields are real input elements. Textareas
and selects are real elements. The boot script posts input and clicks to the
server. `GET /dom` returns the same document without the boot script, with live
values already merged into the controls. The boot script patches id-anchored
elements from that JSON and does not replace the focused field.

`?dom=0` and pages without an executed source use the PNG. `GET /frame.png`
paints the viewport at its current scroll offset and keeps fixed layers.
Scrolled clicks and hover use the same offset. A page kept as a display list
has no bitmap, so `page.Probe().PNG` paints once from the stored template
source and caches the bytes until the next `Redraw`
([screen.md](screen.md)).

## Hot reload and ticks

A file-backed page is polled before `GET /` and `GET /frame.png`, under the
same mutex, so an edit to the template or the theme shows on the next request.
While `Page.Window().Watching()` is true the page pushes `/state` snapshots over `/events`. The live
document fetches `/dom` when `domGeneration` changes and patches id-anchored
elements. The PNG shell swaps the frame image when `generation` changes. Where
`EventSource` is unavailable both shells fall back to polling `/state` (100ms live,
50ms PNG). A string page has no watch, so
the push stays silent and only refreshes when the page redraws.
[hot-reload.md](hot-reload.md) has the config fields, the parse-error
behavior, and what survives a reload.

When the screen implements `host.Ticker`, `Serve` starts a loop at about 60 Hz.
It locks the same mutex, calls `Tick`, drains the damage with
`TakeDirtyRects` when the screen reports it, and redraws only when the page reports
dirty state or the generation changes. A bare `FrameDirty` opt-in with no
rects does not advance the frame generation, and a real commit drains its own
damage so the next idle tick does not bump twice. Each snapshot change is
pushed to `/events` subscribers; idle ticks publish nothing.
The shared `http.Server` sets no write timeout so the `/events` stream stays
open; read and idle timeouts still apply. Cancel the context to stop it.

## What it drops

`Serve` is not the native window loop:

- The desktop audio context plays on this machine and is not forwarded to the browser ([features.md](features.md), [window.md](window.md)).
- No clipboard, undo, redo, select-all, delete-word, or submit route. The
  live document's text route for fields is `/change`. `/type` and `/backspace`
  remain for the picture shell. `/key` still takes the special keys.
- No file dialog, so a file input keeps the typed name ([forms.md](forms.md)).
- No dropped files. `Serve` is not the Ebiten loop, so it never reads `ebiten.DroppedFiles` ([drag-drop.md](drag-drop.md)).
- No crash-report file. A handler panic becomes HTTP 500
  ([crash.md](crash.md)).

A click still focuses a control, and a form still takes text. Values survive a
redraw like any other page.

`Serve` has no window, so it lays the page out once at `Page.Size()` and never
relayouts it. Resizing the browser tab changes how the PNG shell shows the
picture, not the Go layout.

## Trust boundary

`Serve` is a developer preview. It assumes the browser peer is the same user on
the same machine. It returns the rendered page plus its boxes and dimensions,
mutates page state from clicks and typed input, and runs any app handler reached
by `/click`, `/key`, or `/wheel`. It has no authentication, sessions, or TLS.
Mutating routes (`/click`, `/type`, `/backspace`, `/key`, `/wheel`, `/change`) require a
matching `Origin` header (scheme and host); a missing `Origin` fails closed
there, while safe `GET`s allow it.
`/click` bounds its query parameters, `/change` bounds ids and values, the server caps header bytes at 64 KiB,
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
  `/backspace`, `/key`, `/wheel`, `/change`, and `/debug/state` as privileged.
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
