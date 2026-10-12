# Implemented features

Verified scan of code, docs, and tests on 2026-10-03. `make test` passes.
This is a point-in-time report; [features.md](features.md) remains the canonical
feature index with links to the detailed docs.

## Core rendering

- HTML template -> CSS -> layout pipeline via `gowkhtmltopdf`, with hit-test boxes (`internal/render/paint.go`)
- Optional theme stylesheet layered after the template's own styles, swappable at runtime (`Config.Theme`, `Page.SetTheme`; `internal/page/theme.go`, `internal/render/state.go`)
- Display-list replay (fills, strokes, lines, grids, images, shaped text); non-replayable pages fall back to a bitmap with a "bitmap fallback" badge (`internal/render/display.go`, `internal/window/draw.go`)
- Rounded/elliptical/masked strokes, rounded fills, pixel-snapped borders, letter-spaced text, cached fonts/images (`internal/replay/`)
- `Page.PNG` on-demand rasterization with caching (`internal/page/page_image.go`)
- Frame animation: `Page.SetTick` runs a callback each window frame, and the retained `Display` operations can change paint fields without a Redraw (`internal/page/page_tick.go`, `internal/frame/`)

## Forms and interaction

- Click focus/toggle/radio/select activation; typing with auto-repeat Backspace, DeleteWord (`internal/page/form_*.go`, `internal/window/keys.go`)
- `data-bind` two-way binding to struct fields via reflection, `BeforeEdit`/`Change` handlers (`internal/page/bind.go`)
- Copy/Cut/Paste/SelectAll/Undo/Redo ([editing.md](editing.md)); form values survive redraw (`internal/page/page_clip.go`, `internal/page/page_edit.go`, `internal/page/form_merge.go`)
- Host CSS states: `:focus`, `:focus-visible`, `:hover`, `:active`, `:checked` (`internal/page`)
- Mouse-wheel scrolling and draggable scrollbar thumbs ([scrolling.md](scrolling.md)); click, hover, and tap input ([pointer.md](pointer.md)) (`internal/window/fit.go`, `internal/window/scrollbar_drag.go`, `internal/window/pointer.go`)

## Navigation and platform

- Browser-style history `Load`/`Back`/`Forward`/`HTML` with `ErrNoHistory`; `data-action` route map (`internal/page/page_nav.go`)
- Desktop window via Ebiten, wasm browser build, and mobile binding; sizing is in [window.md](window.md) and the run modes are in [platforms.md](platforms.md) (`internal/window/run.go`, `scripts/browser.sh`, `BindMobile`)
- Web mode HTTP server: `GET /`, `GET /frame.png`, `GET /click`, `POST /type`, `POST /backspace` ([web.md](web.md), `internal/web/server.go`)

## Services

- In-process IPC `Send`/`Listen`/`Handle`/`Request` with `ErrNoHandler` (`internal/ipc/`)
- `Fetch` GET + `XHR` arbitrary method; http/https only, no cookies (`fetch.go`, `internal/fetch/`)
- OS clipboard (X11/Windows/macOS) with in-memory fallback and `UseMemory` test mode (`internal/clipboard/`)
- Panic recovery writing local timestamped crash reports (`internal/crash/`)

## Examples and tests

- `examples/login` (sign-in, secret/secret), `examples/forms` (all control types), `examples/bind` (two-way binding), `examples/theme` (light/dark switch), `examples/wispr-flow-dashboard` (component cards, gauge, usage bars, streak heatmap, tab switching), `examples/audio-player` (per-component music player with live iTunes search, fetched artwork, and local free-music playback from Openverse with an animated seek bar and equalizer), `examples/spotify-player` (dark web-player clone with live iTunes data, local free-music playback, and an animated now-bar equalizer), `examples/dino` (keyboard-only dinosaur game with a per-frame tick and an FPS readout)
- `examples/music` is shared example support: Openverse search, an on-disk MP3 cache, MP3/WAV decode, playback through Ebiten audio, and a generated demo tune when the API is unreachable. Package `gpui` does not import it.
- 195 test files across `internal/` and `examples/`, including `internal/page`, `replay`, `render`, `window`, `ipc`, `fetch`, `crash`, and `clipboard`; `make test` passes
- Other examples: `layout`, `replay`, `shapes`, `png`, `controls`, `bind-hooks`, `editing`, `states`, `scrolling`, `history`, `web`, `ipc`, `fetch`, `clipboard`, `crash`, `platform`; index in [examples/readme.md](../examples/readme.md)

## Known gaps

Documented in [features.md](features.md) and [../PHASES.md](../PHASES.md), not
implemented: JS engine/Chromium, cookies/sessions, native menus/tray,
DevTools/auto-update, video/WebGL, a library-level audio API (the examples play
audio through `examples/music`), IME/accessibility, and some engine CSS limits
(transformed WOFF2, `:disabled`, `::placeholder`).
